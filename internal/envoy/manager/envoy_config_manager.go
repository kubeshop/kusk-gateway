// MIT License
//
// Copyright (c) 2022 Kubeshop
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// package manager provide GRPC server configuration and configuration cache manager.
package manager

import (
	"context"
	"net"
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

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func NewEnvoyConfigManager(ctx context.Context, address string, logger logr.Logger) *EnvoyConfigManager {
	snapshotCache := cache.NewSnapshotCache(true, cache.IDHash{}, NewEnvoySnapshotCacheLogger(logger))
	cacheManager := NewCacheManager(snapshotCache, logger)
	callbacks := NewCallbacks(cacheManager, logger)
	server := server.NewServer(ctx, cacheManager, callbacks)

	return &EnvoyConfigManager{
		XDSServer:    &server,
		cacheManager: cacheManager,
		logger:       logger.WithName("EnvoyConfigManager"),
		address:      address,
	}
}

// EnvoyConfigManager holds cacheManager and XDS service
// Only its methods must be called to update Envoy configuration
type EnvoyConfigManager struct {
	XDSServer    *server.Server
	cacheManager *cacheManager
	address      string
	logger       logr.Logger
}

func (em *EnvoyConfigManager) Start() error {
	// Starts GRPC service
	grpcServer := newGRPCServer()
	listener, err := net.Listen("tcp", em.address)
	if err != nil {
		return err
	}

	registerServer(grpcServer, *em.XDSServer)

	em.logger.Info("control plane server listening", "address", em.address)
	return grpcServer.Serve(listener)
}

func (em *EnvoyConfigManager) ApplyNewFleetSnapshot(fleet string, snapshot *cache.Snapshot) error {
	return em.cacheManager.applyNewFleetSnapshot(fleet, snapshot)
}

func registerServer(grpcServer *grpc.Server, server server.Server) {
	// register services
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, server)

	secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, server)
	// TODO(MBana): Not too sure about this one, but I'm leaving it as a reference that an unimplemented SDS could be used.
	// secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, &secretservice.UnimplementedSecretDiscoveryServiceServer{})

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
