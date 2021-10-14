// package config provides structures to create and update routing configuration for Envoy Fleet
// it is not used for Fleet creation, only for configuration snapshot creation.

package config

import (
	"regexp"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoytypematcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
)

var (
	// compiles e.g. /pets/{petID}/legs/{leg1}
	rePathParams = regexp.MustCompile(`/{[A-z0-9]+}`)
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
	vhosts   []string
	routes   []*route.Route
	clusters map[string]*cluster.Cluster
	listener *listener.Listener
}

func New() *envoyConfiguration {
	return &envoyConfiguration{
		clusters: make(map[string]*cluster.Cluster),
	}
}

// AddRoute appends new route to the list of routes by path and method
func (e *envoyConfiguration) AddRoute(name string, path string, method string, clusterName string) {
	// match by this header
	headerMatcher := []*route.HeaderMatcher{
		{
			Name: ":method",
			HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
				StringMatch: &envoytypematcherv3.StringMatcher{
					MatchPattern: &envoytypematcherv3.StringMatcher_Exact{Exact: method},
					IgnoreCase:   false,
				},
			},
		},
	}

	var routeMatcher *route.RouteMatch

	// if regex - matcher is using RouteMatch_Regex with /{pattern} replaced by /([A-z0-9]+) regex
	// if simple path - RouteMatch_Path with path
	if rePathParams.MatchString(path) {
		replacementRegex := string(rePathParams.ReplaceAll([]byte(path), []byte("/([A-z0-9]+)")))
		routeMatcher = &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_SafeRegex{
				SafeRegex: &envoytypematcherv3.RegexMatcher{
					EngineType: &envoytypematcherv3.RegexMatcher_GoogleRe2{
						GoogleRe2: &envoytypematcherv3.RegexMatcher_GoogleRE2{},
					},
					Regex: replacementRegex,
				},
			},
			Headers: headerMatcher,
		}
	} else {
		routeMatcher = &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Path{
				Path: path,
			},
			Headers: headerMatcher,
		}
	}
	// finally create the route and append it to the list
	rt := &route.Route{
		Name:  name,
		Match: routeMatcher,
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: clusterName,
				},
				// TODO: clarify if we need to rewrite Host to service hostname at all
				// HostRewriteSpecifier: &route.RouteAction_HostRewriteLiteral{
				// 	HostRewriteLiteral: upstreamServiceHost,
				// },
			},
		},
	}
	e.routes = append(e.routes, rt)
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
	return &route.RouteConfiguration{
		Name: routeConfigName,
		VirtualHosts: []*route.VirtualHost{{
			Name:    "local_service",
			Domains: e.vhosts,
			Routes:  e.routes,
		}},
	}
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
		HttpFilters: []*hcm.HttpFilter{{
			Name: wellknown.Router,
		}},
	}
	pbst, err := ptypes.MarshalAny(manager)
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

func (e *envoyConfiguration) GenerateSnapshot() (*cache.Snapshot, error) {
	var clusters []types.Resource
	for _, cluster := range e.clusters {
		clusters = append(clusters, cluster)
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
