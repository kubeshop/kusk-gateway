package management

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/go-logr/logr"
	"github.com/kubeshop/kusk-gateway/internal/helper/mocking"
)

func New(ctx context.Context, address string, log logr.Logger) *ConfigManager {
	cacheManager := &cacheManager{fleetConfigs: make(map[string]*Snapshot), mu: &sync.RWMutex{}}
	logger := log.WithName("helper-config-manager")
	// callbacks := Callbacks{cacheMgr: &cacheManager, log: log}
	return &ConfigManager{
		cacheManager: cacheManager,
		l:            logger,
		address:      address,
	}
}

// This is the snapshot of the configuration that helper node receives.
// type fleetConfig struct {
// 	mockConfigs *mocking.MockConfig
// }

// ###############################  Cache Manager ########################################################
// cacheManager provides the snapshots cache and the methods to update it with the new configuration for the specific Envoy fleet
type cacheManager struct {
	// active cache snapshot per fleet
	fleetConfigs map[string]*Snapshot
	mu           *sync.RWMutex
	// need to embed this to implement the interface for GRPC types
	UnimplementedConfigManagerServer
}

// UpdateFleetConfigSnapshot creates the new fleet configuration snapshot and updates the related internal cache
func (c *cacheManager) UpdateFleetConfigSnapshot(fleetID string, mockConfig *mocking.MockConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	protoMockConfig := MockConfigToProtoMockConfig(mockConfig)
	c.fleetConfigs[fleetID] = &Snapshot{
		MockConfig: protoMockConfig,
	}
	// TODO: update nodes
}

// GetSnapshot is a server side function to return the protobuf encoded snapshot to the client.
func (c cacheManager) GetSnapshot(clientParams *ClientParams, srv ConfigManager_GetSnapshotServer) error {
	snapshot := &Snapshot{}
	if fleetConfig, ok := c.fleetConfigs[clientParams.FleetID]; ok {
		snapshot = fleetConfig
	}
	if err := srv.Send(snapshot); err != nil {
		return err
	}
	return nil
}

// func (cm *cacheManager) IsNodeExist(nodeID string) bool {
// 	if status := cm.GetStatusInfo(nodeID); status != nil {
// 		return true
// 	}
// 	return false
// }

// func (cm *cacheManager) getNodesWithCluster(cluster string) []string {
// 	var nodesIDs []string
// 	for _, nodeID := range cm.GetStatusKeys() {
// 		if cm.GetStatusInfo(nodeID).GetNode().Cluster == cluster {
// 			nodesIDs = append(nodesIDs, nodeID)
// 		}
// 	}
// 	return nodesIDs
// }

// // setNodeSnapshot sets new node snapshot from active fleet configuration snapshot
// func (cm *cacheManager) setNodeSnapshot(nodeID string, fleet string) error {
// 	cm.mu.RLock()
// 	snapshot, ok := cm.fleetSnapshot[fleet]
// 	cm.mu.RUnlock()
// 	if !ok {
// 		return fmt.Errorf("no such %s Envoy fleet (cluster) configuration exist", fleet)
// 	}
// 	return cm.SetSnapshot(context.Background(), nodeID, *snapshot)
// }

// // applyNewFleetSnapshot assigns active snapshot and updates all nodes with it
// func (cm *cacheManager) applyNewFleetSnapshot(fleet string, newSnapshot *cache.Snapshot) error {
// 	if err := newSnapshot.Consistent(); err != nil {
// 		return fmt.Errorf("inconsistent snapshot %v", newSnapshot)
// 	}
// 	cm.mu.Lock()
// 	cm.fleetSnapshot[fleet] = newSnapshot
// 	cm.mu.Unlock()
// 	errs := []error{}
// 	// Update caches for existing nodes with only this fleet
// 	for _, nodeID := range cm.getNodesWithCluster(fleet) {
// 		if err := cm.setNodeSnapshot(nodeID, fleet); err != nil {
// 			errs = append(errs, err)
// 		}
// 	}
// 	// Convert list of errors to strings.
// 	if len(errs) > 0 {
// 		errors := make([]string, len(errs))
// 		for i := 0; i < len(errs); i++ {
// 			errors[i] = errs[i].Error()
// 		}
// 		return fmt.Errorf(strings.Join(errors, "\n"))
// 	}
// 	return nil
// }

// ###############################  Config Manager ########################################################
// ConfigManager manages all fleet configuration for the fleets
// It contains cacheManager for the data and runs GRPC service to updates helper nodes
type ConfigManager struct {
	// MockingServer    *server.Server
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
		grpcKeepaliveTime        = 30 * time.Second
		grpcKeepaliveTimeout     = 5 * time.Second
		grpcKeepaliveMinTime     = 30 * time.Second
		grpcMaxConcurrentStreams = 100
	)
	grpcOptions = append(grpcOptions,
		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    grpcKeepaliveTime,
			Timeout: grpcKeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             grpcKeepaliveMinTime,
			PermitWithoutStream: true,
		}),
	)
	return grpc.NewServer(grpcOptions...)
}
