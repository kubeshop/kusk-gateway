package config

import (
	"fmt"
	"sort"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	cacheTypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/gofrs/uuid"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/kubeshop/kusk-gateway/envoy/types"
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

type EnvoyConfiguration struct {
	// vhosts maps vhost domain name or domain pattern to the list of vhosts assigned to the listener
	vHosts   map[string]*types.VirtualHost
	clusters map[string]*cluster.Cluster
	listener *listener.Listener
}

func New() *EnvoyConfiguration {
	return &EnvoyConfiguration{
		clusters: make(map[string]*cluster.Cluster),
		vHosts:   make(map[string]*types.VirtualHost),
	}
}

func (e *EnvoyConfiguration) AddListener(l *listener.Listener) {
	e.listener = l
}

func (e *EnvoyConfiguration) GetVirtualHosts() map[string]*types.VirtualHost {
	return e.vHosts
}

func (e *EnvoyConfiguration) AddVirtualHost(vh *types.VirtualHost) {
	e.vHosts[vh.Name] = vh
}

// AddRouteToVHost appends new route with proxying to the upstream to the list of routes by path and method
func (e *EnvoyConfiguration) AddRouteToVHost(vhost string, rt *route.Route) error {
	virtualHost, ok := e.vHosts[vhost]

	if !ok {
		return fmt.Errorf("envoy configuration doesnt have virtualhost: %s", vhost)
	}

	if err := virtualHost.AddRoute(rt); err != nil {
		return fmt.Errorf("route %s already exists for vhost %s", rt.GetName(), vhost)
	}

	return nil
}

func (e *EnvoyConfiguration) ClusterExist(name string) bool {
	_, exist := e.clusters[name]
	return exist
}

// AddCluster creates Envoy cluster which is the representation of backend service
// For the simplicity right now we don't support endpoints assignments separately, i.e. one cluster - one endpoint, not multiple load balanced
// Cluster with the same name will be overwritten
func (e *EnvoyConfiguration) AddCluster(clusterName, upstreamServiceHost string, upstreamServicePort uint32) {
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
		ConnectTimeout:       &durationpb.Duration{Seconds: 5},
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_LOGICAL_DNS},
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment:       upstreamEndpoint,
		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
	}
}

func (e *EnvoyConfiguration) GenerateSnapshot() (*cache.Snapshot, error) {
	var clusters []cacheTypes.Resource
	for _, c := range e.clusters {
		clusters = append(clusters, c)
	}
	// We're using uuid V1 to provide time sortable snapshot version
	snapshotVersion, _ := uuid.NewV1()
	snap, err := cache.NewSnapshot(snapshotVersion.String(),
		map[resource.Type][]cacheTypes.Resource{
			resource.ClusterType:  clusters,
			resource.RouteType:    {e.makeRouteConfiguration(RouteName)},
			resource.ListenerType: {e.listener},
		},
	)
	if err != nil {
		return nil, err
	}
	return &snap, snap.Consistent()
}

func (e *EnvoyConfiguration) makeRouteConfiguration(routeConfigName string) *route.RouteConfiguration {
	var vhosts []*route.VirtualHost
	for _, vhost := range e.vHosts {
		vhost.Routes = sortRoutesByPathMatcher(vhost.Routes)
		vhosts = append(vhosts, vhost.VirtualHost)
	}
	return &route.RouteConfiguration{
		Name:         routeConfigName,
		VirtualHosts: vhosts,
	}
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
