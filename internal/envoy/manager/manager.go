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

func (cm *cacheManager) IsNodeExist(nodeID string) bool {
	if status := cm.GetStatusInfo(nodeID); status != nil {
		return true
	}
	return false
}

func (cm *cacheManager) getNodesWithCluster(cluster string) []string {
	var nodesIDs []string
	for _, nodeID := range cm.GetStatusKeys() {
		if cm.GetStatusInfo(nodeID).GetNode().Cluster == cluster {
			nodesIDs = append(nodesIDs, nodeID)
		}
	}
	return nodesIDs
}

// setNodeSnapshot sets new node snapshot from active fleet configuration snapshot
func (cm *cacheManager) setNodeSnapshot(nodeID string, fleet string) error {
	cm.mu.RLock()
	snapshot, ok := cm.fleetSnapshot[fleet]
	cm.mu.RUnlock()
	if !ok {
		return fmt.Errorf("no such %s Envoy fleet (cluster) configuration exist", fleet)
	}
	return cm.SetSnapshot(context.Background(), nodeID, *snapshot)
}

// applyNewFleetSnapshot assigns active snapshot and updates all nodes with it
func (cm *cacheManager) applyNewFleetSnapshot(fleet string, newSnapshot *cache.Snapshot) error {
	if err := newSnapshot.Consistent(); err != nil {
		return fmt.Errorf("inconsistent snapshot %v", newSnapshot)
	}
	cm.mu.Lock()
	cm.fleetSnapshot[fleet] = newSnapshot
	cm.mu.Unlock()
	errs := []error{}
	// Update caches for existing nodes with only this fleet
	for _, nodeID := range cm.getNodesWithCluster(fleet) {
		if err := cm.setNodeSnapshot(nodeID, fleet); err != nil {
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
