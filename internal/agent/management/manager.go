/*
MIT License

Copyright (c) 2022 Kubeshop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package management

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/kubeshop/kusk-gateway/internal/agent/mocking"
)

func New(address string, log logr.Logger) *ConfigManager {
	cacheManager := &cacheManager{
		fleetConfigs:                make(map[string]*Snapshot),
		fleetConfigsMutex:           &sync.RWMutex{},
		fleetNodesConnections:       make(map[string]map[string]chan *Snapshot),
		fleetNodesConnectionsMutex:  &sync.RWMutex{},
		randomNodeStreamIDGenerator: rand.New(rand.NewSource(100)),
	}
	logger := log.WithName("agent-config-manager")
	return &ConfigManager{
		cacheManager: cacheManager,
		l:            logger,
		address:      address,
	}
}

// ###############################  Cache Manager ########################################################
// cacheManager provides the snapshots cache and the methods to update it with the new configuration for the agents in the fleet
type cacheManager struct {
	// active cache snapshot per fleet, the key - fleetID
	fleetConfigs map[string]*Snapshot
	// Map of [fleets] to the map of [node connection id] and their connection channels
	fleetNodesConnections       map[string]map[string]chan *Snapshot
	fleetNodesConnectionsMutex  *sync.RWMutex
	fleetConfigsMutex           *sync.RWMutex
	randomNodeStreamIDGenerator *rand.Rand
	// need to embed this to implement the interface for GRPC types
	UnimplementedConfigManagerServer
}

// getSnapshot returns a snapshot for the fleet
func (c *cacheManager) getSnapshot(fleetID string) *Snapshot {
	c.fleetConfigsMutex.RLock()
	defer c.fleetConfigsMutex.RUnlock()
	return c.fleetConfigs[fleetID]
}

// createAndSetSnapshot creates snapshot from mockingConfig and adds snapshot to the fleet mapping
func (c *cacheManager) createAndSetSnapshot(fleetID string, mockConfig *mocking.MockConfig) {
	c.fleetConfigsMutex.Lock()
	defer c.fleetConfigsMutex.Unlock()
	c.fleetConfigs[fleetID] = &Snapshot{
		MockConfig: MockConfigToProtoMockConfig(mockConfig),
	}
}

func (c *cacheManager) updateNodes(fleetID string) {
	nodeConnections, ok := c.fleetNodesConnections[fleetID]
	// No fleet nodes connections are registered
	if !ok {
		return
	}
	snapshot := c.getSnapshot(fleetID)
	c.fleetNodesConnectionsMutex.RLock()
	defer c.fleetNodesConnectionsMutex.RUnlock()
	for _, ch := range nodeConnections {
		ch <- snapshot
	}
}

func (c *cacheManager) registerClientConnection(fleetID string, nodeStreamID string) <-chan *Snapshot {
	ch := make(chan *Snapshot)
	c.fleetNodesConnectionsMutex.Lock()
	defer c.fleetNodesConnectionsMutex.Unlock()
	// Create nodes connections map if missing
	if _, ok := c.fleetNodesConnections[fleetID]; !ok {
		c.fleetNodesConnections[fleetID] = make(map[string]chan *Snapshot)
	}
	// register node connection in the connections map
	c.fleetNodesConnections[fleetID][nodeStreamID] = ch
	return ch
}

func (c *cacheManager) unregisterClientConnection(fleetID string, nodeStreamID string) {
	c.fleetNodesConnectionsMutex.Lock()
	defer c.fleetNodesConnectionsMutex.Unlock()
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

	// Get snapshot and send to the client once on the start of client connection
	snapshot := c.getSnapshot(clientParams.FleetID)
	// If no snapshot found - break it, manager could be restarting so no snapshots yet.
	if snapshot == nil {
		return fmt.Errorf("no snapshot found for the fleet %s", clientParams.FleetID)
	}
	// Otherwise send it
	if err := stream.Send(snapshot); err != nil {
		return err
	}

	// We register in the connections map and permanently wait for the channel message with the new Snapshot to send id to the client.
	// Generate a random stream ID to avoid the races when the client reconnects faster than this goroutine is terminated.
	// Node new connection will have the different stream ID in the map while the older is removed.
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
// It contains cacheManager for the data and runs GRPC service for updates to agent nodes
type ConfigManager struct {
	cacheManager *cacheManager
	address      string
	l            logr.Logger
}

// ApplyNewFleetConfig adds new configuration to the cache manager and triggers agents update
func (m *ConfigManager) ApplyNewFleetConfig(fleetID string, mockConfig *mocking.MockConfig) {
	m.cacheManager.createAndSetSnapshot(fleetID, mockConfig)
	// Trigger update to all fleet nodes
	m.cacheManager.updateNodes(fleetID)
	m.l.Info("Applied new Agent fleet config", "fleet", fleetID)
	return
}

func (m *ConfigManager) Start() error {
	// Starts GRPC service
	grpcServer := newGRPCServer()
	m.l.Info(fmt.Sprintf("Starting Agent Configuration Management service on %s", m.address))
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
