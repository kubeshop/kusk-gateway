// package manager provides manager that starts GRPC server with the configuration cache.
package manager

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/go-logr/logr"
	pb "github.com/kubeshop/kusk-gateway/internal/helper/management"
	"github.com/kubeshop/kusk-gateway/internal/helper/mocking"
)

func New(ctx context.Context, address string, log logr.Logger) *ConfigManager {
	cacheManager := &cacheManager{fleetMockConfigs: make(map[string]*mocking.MockConfig), mu: &sync.RWMutex{}}
	logger := log.WithName("helper-config-manager")
	// callbacks := Callbacks{cacheMgr: &cacheManager, log: log}
	return &ConfigManager{
		cacheManager: cacheManager,
		l:            logger,
		address:      address,
	}
}

// cacheManager provides the snapshots cache and the methods to update it with the new configuration for the specific Envoy fleet
type cacheManager struct {
	// active cache snapshot per fleet
	fleetMockConfigs map[string]*mocking.MockConfig
	mu               *sync.RWMutex
	// need to embed this to implement the interface
	pb.UnimplementedConfigManagerServer
}

// SetFleetSnapshot creates the new fleet configuration snapshot and updates the related internal cache
func (c *cacheManager) UpdateFleetConfigSnapshot(fleetID string, mockConfig *mocking.MockConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fleetMockConfigs[fleetID] = mockConfig
}

// GetSnapshot is a server side function to return the protobuf encoded snapshot to the client.
func (c cacheManager) GetSnapshot(ctx context.Context, clientParams *pb.ClientParams) (*pb.Snapshot, error) {
	snapshot := &pb.Snapshot{}
	// Retrieve mocking configuration and fill it in if found
	if fleetMockConfig, ok := c.fleetMockConfigs[clientParams.FleetID]; ok {
		snapshot.MockConfig = &pb.MockConfig{
			MockResponses: make(map[string]*pb.MockResponse, len(*fleetMockConfig)),
		}
		for mockID, mockResponse := range *fleetMockConfig {
			// Create protobuf MockResponses with the status code
			snapshot.MockConfig.MockResponses[mockID] = &pb.MockResponse{
				StatusCode:    uint32(mockResponse.StatusCode),
				MediaTypeData: make(map[string][]byte, len(mockResponse.MediaTypeData)),
			}
			// Fill its MediaTypeData mapping
			for mediaType, mediaTypeData := range mockResponse.MediaTypeData {
				snapshot.MockConfig.MockResponses[mockID].MediaTypeData[mediaType] = mediaTypeData
			}
		}
	}
	return snapshot, nil
}

// ConfigManager manages Mocking service configuration for the fleets
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

// ApplyNewFleetConfig adds new mocking configuration to the cache manager and triggers helpers update
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
	pb.RegisterConfigManagerServer(grpcServer, *m.cacheManager)

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
