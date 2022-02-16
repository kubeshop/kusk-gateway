package management

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/go-logr/logr"
	"github.com/kubeshop/kusk-gateway/internal/helper/mocking"
)

func New(ctx context.Context, address string, log logr.Logger) *ConfigManager {
	cacheManager := &cacheManager{
		fleetConfigs:                make(map[string]*Snapshot),
		fleetConfigsMutex:           &sync.RWMutex{},
		fleetNodesConnections:       make(map[string]map[string]chan *Snapshot),
		fleetNodesMutex:             &sync.RWMutex{},
		randomNodeStreamIDGenerator: rand.New(rand.NewSource(100)),
	}
	logger := log.WithName("helper-config-manager")
	return &ConfigManager{
		cacheManager: cacheManager,
		l:            logger,
		address:      address,
	}
}

// ###############################  Cache Manager ########################################################
// cacheManager provides the snapshots cache and the methods to update it with the new configuration for the helpers in the fleet
type cacheManager struct {
	// active cache snapshot per fleet, the key - fleetID
	fleetConfigs map[string]*Snapshot
	// Map of [fleets] to the map of [node connection id] and their connection channels
	fleetNodesConnections       map[string]map[string]chan *Snapshot
	fleetNodesMutex             *sync.RWMutex
	fleetConfigsMutex           *sync.RWMutex
	randomNodeStreamIDGenerator *rand.Rand
	// need to embed this to implement the interface for GRPC types
	UnimplementedConfigManagerServer
}

// UpdateFleetConfigSnapshot creates the new fleet configuration snapshot and updates the related internal cache
func (c *cacheManager) UpdateFleetConfigSnapshot(fleetID string, mockConfig *mocking.MockConfig) {
	pbMockConfig := MockConfigToProtoMockConfig(mockConfig)
	snapshot := &Snapshot{
		MockConfig: pbMockConfig,
	}
	c.fleetConfigsMutex.Lock()
	c.fleetConfigs[fleetID] = snapshot
	c.fleetConfigsMutex.Unlock()
	// Trigger update to all fleet nodes
	c.sendUpdateToFleetNodes(fleetID)
}

func (c *cacheManager) sendUpdateToFleetNodes(fleetID string) {
	snapshot := c.getOrSetSnapshot(fleetID)
	nodeConnections, ok := c.fleetNodesConnections[fleetID]
	// No fleet nodes connections are registered
	if !ok {
		return
	}
	c.fleetNodesMutex.RLock()
	for _, ch := range nodeConnections {
		ch <- snapshot
	}
	c.fleetNodesMutex.RUnlock()
}

// getOrSetSnapshot returns snapshot for the fleet or creates one empty if the fleet entry is missing.
func (c *cacheManager) getOrSetSnapshot(fleetID string) *Snapshot {
	var snapshot *Snapshot
	var ok bool
	c.fleetConfigsMutex.Lock()
	defer c.fleetConfigsMutex.Unlock()
	// Fleet exists (found snapshot)
	if snapshot, ok = c.fleetConfigs[fleetID]; !ok {
		// Not found fleet snapshot in the cache - return empty snapshot and create it
		snapshot = &Snapshot{}
		c.fleetConfigs[fleetID] = snapshot
	}
	return snapshot
}

func (c *cacheManager) registerClientConnection(fleetID string, nodeStreamID string) <-chan *Snapshot {
	ch := make(chan *Snapshot)
	c.fleetNodesMutex.Lock()
	defer c.fleetNodesMutex.Unlock()
	// Create nodes connections map if missing
	_, ok := c.fleetNodesConnections[fleetID]
	if !ok {
		c.fleetNodesConnections[fleetID] = make(map[string]chan *Snapshot)
	}
	// register node connection in the connections map
	c.fleetNodesConnections[fleetID][nodeStreamID] = ch
	return ch
}

func (c *cacheManager) unregisterClientConnection(fleetID string, nodeStreamID string) {
	c.fleetNodesMutex.Lock()
	defer c.fleetNodesMutex.Unlock()
	// Remove this node stream from the fleet's streams
	nodesConnections, ok := c.fleetNodesConnections[fleetID]
	if ok {
		// Close the channel before
		ch, ok := c.fleetNodesConnections[fleetID][nodeStreamID]
		if ok {
			close(ch)
		}
		delete(nodesConnections, nodeStreamID)
	}
}

// GetSnapshot is a server side function to return the protobuf encoded snapshot to the client.
// It will run in its own goroutine for the each call (client).
func (c cacheManager) GetSnapshot(clientParams *ClientParams, stream ConfigManager_GetSnapshotServer) error {
	// Get snapshot and send it once at the start
	snapshot := c.getOrSetSnapshot(clientParams.FleetID)
	if err := stream.Send(snapshot); err != nil {
		return err
	}
	// On success we register in the connections array and permanently wait for the channel message with the Snapshot to send id to the client.
	// Generate a random stream ID to avoid the situations when the client reconnects faster than this goroutine is terminated to avoid races.
	nodeStreamID := fmt.Sprintf("%s:%d", clientParams.NodeName, c.randomNodeStreamIDGenerator.Uint32())
	receiveChan := c.registerClientConnection(clientParams.FleetID, nodeStreamID)
	defer c.unregisterClientConnection(clientParams.FleetID, nodeStreamID)
	// Endlessly stream Snapshots to the client until it closes the connection or returns error.
	for {
		select {
		case <-stream.Context().Done(): // if the client closes the connection - exit
			return stream.Context().Err() // stream will be closed immediately after return
		case snapshot := <-receiveChan:
			if err := stream.Send(snapshot); err != nil {
				return err
			}
		}
	}
}

// ###############################  Config Manager ########################################################
// ConfigManager manages all fleet configuration for the fleets
// It contains cacheManager for the data and runs GRPC service to updates helper nodes
type ConfigManager struct {
	cacheManager *cacheManager
	address      string
	l            logr.Logger
}

func (m *ConfigManager) UpdateFleetNodes(fleetID string) {
	// TODO: write me
	return
}

// ApplyNewFleetConfig adds new configuration to the cache manager and triggers helpers update
func (m *ConfigManager) ApplyNewFleetConfig(fleetID string, mockConfig *mocking.MockConfig) {
	m.cacheManager.UpdateFleetConfigSnapshot(fleetID, mockConfig)
	m.UpdateFleetNodes(fleetID)
	m.l.Info("Applied new fleet config", "fleet", fleetID, "config", mockConfig)
	return
}

func (m *ConfigManager) Start() error {
	// Starts GRPC service
	grpcServer := newGRPCServer()
	m.l.Info(fmt.Sprintf("Starting Helper Configuration Management service on %s", m.address))
	listener, err := net.Listen("tcp", m.address)
	if err != nil {
		return err
	}
	RegisterConfigManagerServer(grpcServer, *m.cacheManager)

	return grpcServer.Serve(listener)
}

func newGRPCServer() *grpc.Server {
	// gRPC golang library sets a very small upper bound for the number gRPC/h2
	// streams over a single TCP connection.
	var grpcOptions []grpc.ServerOption
	const (
		grpcMaxConnectionIdle     = 30 * time.Second // If a client is idle for 30 seconds, send a GOAWAY
		grpcMaxConnectionAge      = 30 * time.Minute // If any connection is alive for more than this time, send a GOAWAY - i.e. force client to reconnect.
		grpcMaxConnectionAgeGrace = 5 * time.Second  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
		grpcKeepaliveTime         = 15 * time.Second // Ping the client if it is idle for this number of seconds to ensure the connection is still active
		grpcKeepaliveTimeout      = 5 * time.Second  // Wait for the ping ack before assuming the connection is dead
		grpcKeepaliveMinTime      = 5 * time.Second  // If a client pings more than once this every seconds, terminate the connection
		grpcMaxConcurrentStreams  = 1000
	)
	grpcOptions = append(grpcOptions,
		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     grpcMaxConnectionIdle,
			MaxConnectionAge:      grpcMaxConnectionAge,
			MaxConnectionAgeGrace: grpcMaxConnectionAgeGrace,
			Time:                  grpcKeepaliveTime,
			Timeout:               grpcKeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             grpcKeepaliveMinTime,
			PermitWithoutStream: true,
		}),
	)
	return grpc.NewServer(grpcOptions...)
}
