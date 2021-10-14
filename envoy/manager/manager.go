// package manager provide GRPC server configuration and configuration cache manager.
package manager

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const DefaultFleetName string = "default"

func New(ctx context.Context, address string, log Logger) *EnvoyConfigManager {
	snapshotCache := cache.NewSnapshotCache(true, cache.IDHash{}, log)
	cacheManager := cacheManager{snapshotCache, make(map[string]*cache.Snapshot), sync.RWMutex{}}
	callbacks := Callbacks{cacheMgr: &cacheManager, log: log}
	server := server.NewServer(ctx, &cacheManager, &callbacks)
	return &EnvoyConfigManager{
		XDSServer:    &server,
		cacheManager: &cacheManager,
		log:          log,
		address:      address,
	}
}

// EnvoyConfigManager holds cacheManager and XDS service
// Only its methods must be called to update Envoy configuration
type EnvoyConfigManager struct {
	XDSServer    *server.Server
	cacheManager *cacheManager
	address      string
	log          Logger
}

func (em *EnvoyConfigManager) Start() error {
	// Starts GRPC service
	grpcServer := newGRPCServer()
	listener, err := net.Listen("tcp", em.address)
	if err != nil {
		return err
	}

	registerServer(grpcServer, *em.XDSServer)

	log.Printf("control plane server listening on %s\n", em.address)
	return grpcServer.Serve(listener)
}

func (em *EnvoyConfigManager) ApplyNewFleetSnapshot(fleet string, snapshot *cache.Snapshot) error {
	return em.cacheManager.applyNewFleetSnapshot(fleet, snapshot)
}

// cacheManager provides cache and methods to update it with new configuration for Envoy fleet
// it is invisible for clients importing the package
type cacheManager struct {
	cache.SnapshotCache
	// active snapshot per fleet
	fleetSnapshot map[string]*cache.Snapshot
	mu            sync.RWMutex
}

func (cm *cacheManager) IsNodeExist(nodeId string) bool {
	if status := cm.GetStatusInfo(nodeId); status != nil {
		return true
	}
	return false
}

// setNodeSnapshot sets new node snapshot from active fleet configuration snapshot
func (cm *cacheManager) setNodeSnapshot(nodeId string, fleet string) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	snapshot, ok := cm.fleetSnapshot[fleet]
	if !ok {
		return fmt.Errorf("no such %s Envoy fleet (cluster) configuration exist", fleet)
	}
	return cm.SetSnapshot(context.Background(), nodeId, *snapshot)
}

// applyNewFleetSnapshot assigns active snapshot and updates all nodes with it
func (cm *cacheManager) applyNewFleetSnapshot(fleet string, newSnapshot *cache.Snapshot) error {
	if err := newSnapshot.Consistent(); err != nil {
		return fmt.Errorf("inconsistent snapshot %v", newSnapshot)
	}
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.fleetSnapshot[fleet] = newSnapshot
	errs := []error{}
	for _, nodeId := range cm.GetStatusKeys() {
		err := cm.SetSnapshot(context.Background(), nodeId, *newSnapshot)
		if err != nil {
			errs = append(errs, err)
		}
	}
	// Convert list of errors to strings.
	if len(errs) > 0 {
		errors := make([]string, len(errs))
		for i := 0; i < len(errs); i++ {
			errors[i] = errs[i].Error()
		}
		return fmt.Errorf(strings.Join(errors, "\n"))
	}
	return nil
}

func registerServer(grpcServer *grpc.Server, server server.Server) {
	// register services
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, server)
	secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, server)
	runtimeservice.RegisterRuntimeDiscoveryServiceServer(grpcServer, server)
}

func newGRPCServer() *grpc.Server {
	// gRPC golang library sets a very small upper bound for the number gRPC/h2
	// streams over a single TCP connection. If a proxy multiplexes requests over
	// a single connection to the management server, then it might lead to
	// availability problems. Keepalive timeouts based on connection_keepalive parameter https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/examples#dynamic
	var grpcOptions []grpc.ServerOption
	const (
		grpcKeepaliveTime        = 30 * time.Second
		grpcKeepaliveTimeout     = 5 * time.Second
		grpcKeepaliveMinTime     = 30 * time.Second
		grpcMaxConcurrentStreams = 1000000
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
