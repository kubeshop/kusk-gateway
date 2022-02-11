// package manager provide GRPC server configuration and configuration cache manager.
package manager

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	pb "github.com/kubeshop/kusk-gateway/internal/mocking/management"
	mt "github.com/kubeshop/kusk-gateway/internal/mocking/types"
)

func New(ctx context.Context, address string, log Logger) *MockingConfigManager {
	// snapshotCache := cache.NewSnapshotCache(true, cache.IDHash{}, log)
	mockingCacheManager := &cacheManager{fleetSnapshot: make(map[string]*mt.MockConfig)}
	// callbacks := Callbacks{cacheMgr: &cacheManager, log: log}
	// server := server.NewServer(ctx, &cacheManager, &callbacks)
	return &MockingConfigManager{
		// XDSServer:    &server,
		// cacheManager: &cacheManager,
		mockingCacheManager: mockingCacheManager,
		log:                 log,
		address:             address,
	}
}

// cacheManager provides the Mocking snapshots cache and the methods to update it with new mocking configuration for the specific Envoy fleet
type cacheManager struct {
	// active cache snapshot per fleet
	fleetSnapshot map[string]*mt.MockConfig
	mu            sync.RWMutex
	// need to embed this to implement the interface
	pb.UnimplementedConfigManagerServer
}

func (cm *cacheManager) GetMockSnapshot(*pb.ClientParams, ConfigManager_GetMockSnapshotServer) error

// MockingConfigManager manages Mocking service configuration for the fleets
type MockingConfigManager struct {
	// MockingServer    *server.Server
	mockingCacheManager *cacheManager
	address             string
	log                 Logger
}

func (m *MockingConfigManager) Start() error {
	// Starts GRPC service
	grpcServer := newGRPCServer()
	listener, err := net.Listen("tcp", m.address)
	if err != nil {
		return err
	}

	// registerServer(grpcServer, *em.XDSServer)

	log.Printf("control plane server listening on %s\n", m.address)
	return grpcServer.Serve(listener)
}

// func (em *EnvoyConfigManager) ApplyNewFleetSnapshot(fleet string, snapshot *cache.Snapshot) error {
// 	return em.cacheManager.applyNewFleetSnapshot(fleet, snapshot)
// }

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

// func registerServer(grpcServer *grpc.Server, server server.Server) {
// 	// register services
// 	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
// 	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
// 	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, server)
// 	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, server)
// 	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, server)
// 	secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, server)
// 	runtimeservice.RegisterRuntimeDiscoveryServiceServer(grpcServer, server)
// }

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
