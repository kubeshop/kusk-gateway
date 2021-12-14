// package config provides structures to create and update routing configuration for Envoy Fleet
// it is not used for Fleet creation, only for configuration snapshot creation.

package envoy

import (
	"fmt"
	"sort"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/types/known/anypb"
)

// TODO: move to params
const (
	ListenerName string = "listener_0"
	ListenerPort uint32 = 8080
	RouteName    string = "local_route"
)

// Simplified objects hierarchy configuration as for the static Envoy config
// Top level objects are "listeners" and "clusters"
//
// listeners:
//   - address: (address:port)
// .........
//     VirtualHost:
//      Routes (RoutesConfiguration):
//       route:
//        - match:
//            path: /bla
//            headers (method)
//          route:
//            cluster: clusterRef-cluster1
//        - match:
//            path: /blabla
//            headers (method)
//          route:
//            cluster: clusterRef-cluster1
// clusters:
// - name: cluster1
//     load_assignment:
//       cluster_name: cluster1
//       endpoints:
//        - lb_endpoints:
//          - endpoint:
//              address:
//                address: backendsvc1-dns-name
//                port_value: backendsvc1-port
//

type envoyConfiguration struct {
	// vhosts maps vhost domain name or domain pattern to the list of vhosts assigned to the listener
	vhosts   map[string]*route.VirtualHost
	clusters map[string]*cluster.Cluster
	listener *listener.Listener
}

func New() *envoyConfiguration {
	return &envoyConfiguration{
		clusters: make(map[string]*cluster.Cluster),
		vhosts:   make(map[string]*route.VirtualHost),
	}
}

// AddRoute appends new route with proxying to the upstream to the list of routes by path and method
func (e *envoyConfiguration) AddRoute(
	vhosts []string,
	rt *route.Route,
) error {
	// Add this route to the list of vhost it applies to
	for _, vhost := range vhosts {
		vhostConfig, ok := e.vhosts[vhost]
		// add if new vhost entry
		if !ok {
			vhostConfig = &route.VirtualHost{
				Name:    vhost,
				Domains: []string{vhost},
			}
			e.vhosts[vhost] = vhostConfig
		}
		if routeExists(rt.GetName(), vhostConfig.Routes) {
			return fmt.Errorf("route %s already exists for vhost %s", rt.GetName(), vhostConfig.Name)
		}
		vhostConfig.Routes = append(vhostConfig.Routes, rt)
	}
	return nil
}

func routeExists(name string, routes []*route.Route) bool {
	for _, rt := range routes {
		if rt.Name == name {
			return true
		}
	}
	return false
}

func (e *envoyConfiguration) ClusterExist(name string) bool {
	_, exist := e.clusters[name]
	return exist
}

// AddCluster creates Envoy cluster which is the representation of backend service
// For the simplicity right now we don't support endpoints assignments separately, i.e. one cluster - one endpoint, not multiple load balanced
// Cluster with the same name will be overwritten
func (e *envoyConfiguration) AddCluster(clusterName string, upstreamServiceHost string, upstreamServicePort uint32) {
	upstreamEndpoint := &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*endpoint.LbEndpoint{{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol: core.SocketAddress_TCP,
									Address:  upstreamServiceHost,
									PortSpecifier: &core.SocketAddress_PortValue{
										PortValue: upstreamServicePort,
									},
								},
							},
						},
					},
				},
			}},
		}},
	}

	e.clusters[clusterName] = &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_LOGICAL_DNS},
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment:       upstreamEndpoint,
		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
	}
}

func (e *envoyConfiguration) makeRouteConfiguration(routeConfigName string) *route.RouteConfiguration {
	var vhosts []*route.VirtualHost
	for _, vhost := range e.vhosts {
		vhost.Routes = sortRoutesByPathMatcher(vhost.Routes)
		vhosts = append(vhosts, vhost)
	}
	return &route.RouteConfiguration{
		Name:         routeConfigName,
		VirtualHosts: vhosts,
	}
}

func (e *envoyConfiguration) GenerateSnapshot() (*cache.Snapshot, error) {
	var clusters []types.Resource
	for _, c := range e.clusters {
		clusters = append(clusters, c)
	}
	// We're using uuid V1 to provide time sortable snapshot version
	snapshotVersion, _ := uuid.NewV1()
	snap, err := cache.NewSnapshot(snapshotVersion.String(),
		map[resource.Type][]types.Resource{
			resource.ClusterType:  clusters,
			resource.RouteType:    {e.makeRouteConfiguration(RouteName)},
			resource.ListenerType: {makeHTTPListener(ListenerName, RouteName)},
		},
	)
	if err != nil {
		return nil, err
	}
	return &snap, snap.Consistent()
}

func makeHTTPListener(listenerName string, routeConfigName string) *listener.Listener {
	// HTTP filter configuration
	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "http",
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource:    makeConfigSource(),
				RouteConfigName: routeConfigName,
			},
		},
		HttpFilters: []*hcm.HttpFilter{
			{
				Name: wellknown.CORS,
			},
			{
				Name: wellknown.Router,
			}},
	}

	pbst, err := anypb.New(manager)
	if err != nil {
		panic(err)
	}

	return &listener.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "0.0.0.0",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: ListenerPort,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{{
			Filters: []*listener.Filter{{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
	}
}

func makeConfigSource() *core.ConfigSource {
	source := &core.ConfigSource{}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	source.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
		ApiConfigSource: &core.ApiConfigSource{
			TransportApiVersion:       resource.DefaultAPIVersion,
			ApiType:                   core.ApiConfigSource_GRPC,
			SetNodeOnFirstMessageOnly: true,
			GrpcServices: []*core.GrpcService{{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"},
				},
			}},
		},
	}
	return source
}

// sortRoutesByPathMatcher creates route list ordered by:
// * path matcher
// * regex path matcher, longest regex path first
// * prefix path matcher, longest path first
// Envoy matches path by the first win in the routes order, so we need to be specific as possible.
func sortRoutesByPathMatcher(routes []*route.Route) []*route.Route {
	result := make([]*route.Route, 0, len(routes))
	resultRegexPathsRoutes := []*route.Route{}
	resultPrefixPathsRoutes := []*route.Route{}
	for _, rt := range routes {
		switch t := rt.Match.PathSpecifier.(type) {
		case *route.RouteMatch_Path:
			// We don't care if path is longer or shorter, path matcher will match correctly
			result = append(result, rt)
		case *route.RouteMatch_SafeRegex:
			resultRegexPathsRoutes = append(resultRegexPathsRoutes, rt)
		case *route.RouteMatch_Prefix:
			resultPrefixPathsRoutes = append(resultPrefixPathsRoutes, rt)
		default:
			// We won't handle this as an error, this qualifies for a panic
			panic(fmt.Sprintf("unsupported route path matcher type: %T", t))
		}
	}
	// Sort by regex length, longest first.
	// Note that this doesn't mean the regex will match the longest path exactly 100%, regex itself adds to the length to compare.
	// However longer regex usually are more specific and should be prioritized.
	sort.SliceStable(
		resultRegexPathsRoutes,
		func(i, j int) bool {
			return len(resultRegexPathsRoutes[i].Match.GetSafeRegex().GetRegex()) > len(resultRegexPathsRoutes[j].Match.GetSafeRegex().GetRegex())
		},
	)
	result = append(result, resultRegexPathsRoutes...)
	// Sort by the path prefix length, longest first.
	sort.SliceStable(
		resultPrefixPathsRoutes,
		func(i, j int) bool {
			return len(resultPrefixPathsRoutes[i].Match.GetPrefix()) > len(resultPrefixPathsRoutes[j].Match.GetPrefix())
		},
	)
	result = append(result, resultPrefixPathsRoutes...)

	return result
}
