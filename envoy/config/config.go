// package config provides structures to create and update routing configuration for Envoy Fleet
// it is not used for Fleet creation, only for configuration snapshot creation.

package config

import (
	"fmt"
	"regexp"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoytypematcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/durationpb"
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

// RouteConfiguration is a convenient configuration object used for route creation
type RouteConfiguration struct {
	name             string
	path             string
	method           string
	clusterName      string
	corsPolicy       *route.CorsPolicy
	rewritePathRegex *envoytypematcher.RegexMatchAndSubstitute
	timeout          int64
	idleTimeout      int64
	retries          uint32
	redirectAction   *route.RedirectAction
}

// AddRoute appends new route with proxying to the backend to the list of routes by path and method
func (e *envoyConfiguration) AddRoute(config *RouteConfiguration) {
	// routeAction defines route parameters whose fields will be filled out further down
	routeAction := &route.Route_Route{
		Route: &route.RouteAction{
			ClusterSpecifier: &route.RouteAction_Cluster{
				Cluster: config.clusterName,
			},
		},
	}
	// headerMatcher allows as to match route by method (:method header)
	var headerMatcher []*route.HeaderMatcher
	// If CORS specified, we add OPTIONS method to the route and enable CORS in the route
	if config.corsPolicy != nil {
		routeAction.Route.Cors = config.corsPolicy
		// header matcher with OPTIONS or main method
		headerMatcher = []*route.HeaderMatcher{
			{
				Name: ":method",
				HeaderMatchSpecifier: &route.HeaderMatcher_SafeRegexMatch{
					SafeRegexMatch: &envoytypematcher.RegexMatcher{
						EngineType: &envoytypematcher.RegexMatcher_GoogleRe2{},
						Regex:      fmt.Sprintf("^OPTIONS|%s$", config.method),
					},
				},
			},
		}
	} else {
		// otherwise simple exact match by method
		headerMatcher = []*route.HeaderMatcher{
			{
				Name: ":method",
				HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{StringMatch: &envoytypematcher.StringMatcher{
					MatchPattern: &envoytypematcher.StringMatcher_Exact{
						Exact: config.method,
					},
				}},
			},
		}
	}
	var routeMatcher *route.RouteMatch
	// if regex in the path - matcher is using RouteMatch_Regex with /{pattern} replaced by /([A-z0-9]+) regex
	// if simple path - RouteMatch_Path with path
	if rePathParams.MatchString(config.path) {
		replacementRegex := string(rePathParams.ReplaceAll([]byte(config.path), []byte("/([A-z0-9]+)")))
		routeMatcher = &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_SafeRegex{
				SafeRegex: &envoytypematcher.RegexMatcher{
					EngineType: &envoytypematcher.RegexMatcher_GoogleRe2{
						GoogleRe2: &envoytypematcher.RegexMatcher_GoogleRE2{},
					},
					Regex: replacementRegex,
				},
			},
			Headers: headerMatcher,
		}
	} else {
		routeMatcher = &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Path{
				Path: config.path,
			},
			Headers: headerMatcher,
		}
	}

	if config.rewritePathRegex != nil {
		routeAction.Route.RegexRewrite = config.rewritePathRegex
	}

	if config.timeout != 0 {
		routeAction.Route.Timeout = &durationpb.Duration{Seconds: config.timeout}
	}
	if config.idleTimeout != 0 {
		routeAction.Route.IdleTimeout = &durationpb.Duration{Seconds: config.idleTimeout}
	}

	if config.retries > 0 {
		routeAction.Route.RetryPolicy = &route.RetryPolicy{
			RetryOn:    "5xx",
			NumRetries: &wrappers.UInt32Value{Value: config.retries},
		}
	}

	// finally create the route and append it to the list
	rt := &route.Route{
		Name:   config.name,
		Match:  routeMatcher,
		Action: routeAction,
	}
	e.routes = append(e.routes, rt)
}

// AddRouteRedirect appends new route with the redirect to the list of routes by path and method
func (e *envoyConfiguration) AddRouteRedirect(config *RouteConfiguration) {

	headerMatcher := []*route.HeaderMatcher{
		{
			Name: ":method",
			HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{StringMatch: &envoytypematcher.StringMatcher{
				MatchPattern: &envoytypematcher.StringMatcher_Exact{
					Exact: config.method,
				},
			}},
		},
	}
	var routeMatcher *route.RouteMatch
	// if regex in the path - matcher is using RouteMatch_Regex with /{pattern} replaced by /([A-z0-9]+) regex
	// if simple path - RouteMatch_Path with path
	if rePathParams.MatchString(config.path) {
		replacementRegex := string(rePathParams.ReplaceAll([]byte(config.path), []byte("/([A-z0-9]+)")))
		routeMatcher = &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_SafeRegex{
				SafeRegex: &envoytypematcher.RegexMatcher{
					EngineType: &envoytypematcher.RegexMatcher_GoogleRe2{
						GoogleRe2: &envoytypematcher.RegexMatcher_GoogleRE2{},
					},
					Regex: replacementRegex,
				},
			},
			Headers: headerMatcher,
		}
	} else {
		routeMatcher = &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Path{
				Path: config.path,
			},
			Headers: headerMatcher,
		}
	}
	// finally create the route and append it to the list
	rt := &route.Route{
		Name:  config.name,
		Match: routeMatcher,
		Action: &route.Route_Redirect{
			Redirect: config.redirectAction,
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
		HttpFilters: []*hcm.HttpFilter{
			{
				Name: wellknown.CORS,
			},
			{
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
	snapshot_version, _ := uuid.NewV1()
	snap, err := cache.NewSnapshot(snapshot_version.String(),
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
