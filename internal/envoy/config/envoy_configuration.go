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

package config

import (
	"fmt"
	"sort"

	"github.com/davecgh/go-spew/spew"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_extensions_transport_sockets_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_cache_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	envoy_resource_v3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	google_uuid "github.com/google/uuid"

	"github.com/go-logr/logr"
	"github.com/gofrs/uuid"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/kubeshop/kusk-gateway/internal/envoy/types"
)

// type SecretType string

// const (
// 	TokenSecretType SecretType = "token_secret"
// 	HMACSecretType  SecretType = "hmac_secret"
// )

// EnvoyConfiguration
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
	secrets  map[string]*envoy_extensions_transport_sockets_tls_v3.Secret
	logger   logr.Logger
}

func NewEnvoyConfiguration(logger logr.Logger) *EnvoyConfiguration {
	return &EnvoyConfiguration{
		clusters: make(map[string]*cluster.Cluster),
		vHosts:   make(map[string]*types.VirtualHost),
		secrets:  map[string]*envoy_extensions_transport_sockets_tls_v3.Secret{},
		logger:   logger.WithName("EnvoyConfiguration"),
	}
}

func (e *EnvoyConfiguration) AddListener(l *listener.Listener) {
	e.listener = l
}

func (e *EnvoyConfiguration) GetVirtualHosts() map[string]*types.VirtualHost {
	return e.vHosts
}

func (e *EnvoyConfiguration) GetVirtualHost(name string) *types.VirtualHost {
	return e.vHosts[name]
}

func (e *EnvoyConfiguration) AddVirtualHost(vh *types.VirtualHost) {
	// Don't add if already present
	if _, ok := e.vHosts[vh.Name]; ok {
		return
	}
	e.vHosts[vh.Name] = vh
}

// AddRouteToVHost appends new route with proxying to the upstream to the list of routes by path and method
func (e *EnvoyConfiguration) AddRouteToVHost(vhost string, rt *route.Route) error {
	virtualHost, ok := e.vHosts[vhost]

	if !ok {
		return fmt.Errorf("envoy configuration doesnt have virtualhost: %s", vhost)
	}

	if err := virtualHost.AddRoute(rt); err != nil {
		return fmt.Errorf("can't add route %s to vhost %s: %w", rt.GetName(), vhost, err)
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
	loadAssignment := createLoadAssignment(clusterName, upstreamServiceHost, upstreamServicePort)

	e.clusters[clusterName] = &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       &durationpb.Duration{Seconds: 5},
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_LOGICAL_DNS},
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment:       loadAssignment,
		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
	}
}

// AddClusterWithTLS - AddCluster with SNI or rather `AutoSni` enabled`.
// Example SNI : "kubeshop-kusk-gateway-oauth2.eu.auth0.com" -> "eu.auth0.com"
func (e *EnvoyConfiguration) AddClusterWithTLS(clusterName, upstreamServiceHost string, upstreamServicePort uint32) error {
	loadAssignment := createLoadAssignment(clusterName, upstreamServiceHost, upstreamServicePort)

	upstreamTlsContext := &envoy_extensions_transport_sockets_tls_v3.UpstreamTlsContext{}
	anyUpstreamTlsContext, err := anypb.New(upstreamTlsContext)
	if err != nil {
		return fmt.Errorf("EnvoyConfiguration.AddClusterWithTLS: failed on `anypb.New(upstreamTlsContext)`, %w", err)
	}

	transportSocket := &envoy_config_core_v3.TransportSocket{
		Name: "envoy.transport_sockets.tls",
		ConfigType: &envoy_config_core_v3.TransportSocket_TypedConfig{
			TypedConfig: anyUpstreamTlsContext,
		},
	}

	cluster := &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       &durationpb.Duration{Seconds: 5},
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_LOGICAL_DNS},
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment:       loadAssignment,
		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
		TransportSocket:      transportSocket,
		UpstreamHttpProtocolOptions: &envoy_config_core_v3.UpstreamHttpProtocolOptions{
			// Set transport socket `SNI <https://en.wikipedia.org/wiki/Server_Name_Indication>`_ for new
			// upstream connections based on the downstream HTTP host/authority header or any other arbitrary
			// header when :ref:`override_auto_sni_header <envoy_v3_api_field_config.core.v3.UpstreamHttpProtocolOptions.override_auto_sni_header>`
			// is set, as seen by the :ref:`router filter <config_http_filters_router>`.
			AutoSni: true,
		},
	}

	if err := cluster.ValidateAll(); err != nil {
		return fmt.Errorf("EnvoyConfiguration.AddClusterWithTLS: failed to validate cluster=%v, %w", cluster, err)
	}

	e.clusters[clusterName] = cluster

	return nil
}

func createLoadAssignment(clusterName string, upstreamServiceHost string, upstreamServicePort uint32) *endpoint.ClusterLoadAssignment {
	upstreamEndpoint := &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*endpoint.LbEndpoint{
				{
					HostIdentifier: &endpoint.LbEndpoint_Endpoint{
						Endpoint: &endpoint.Endpoint{
							Address: &envoy_config_core_v3.Address{
								Address: &envoy_config_core_v3.Address_SocketAddress{
									SocketAddress: &envoy_config_core_v3.SocketAddress{
										Protocol: envoy_config_core_v3.SocketAddress_TCP,
										Address:  upstreamServiceHost,
										PortSpecifier: &envoy_config_core_v3.SocketAddress_PortValue{
											PortValue: upstreamServicePort,
										},
									},
								},
							},
						},
					},
				},
			},
		}},
	}

	// TODO: Should be calling `ValidateAll()` here.
	// if err := upstreamEndpoint.ValidateAll(); err != nil {
	// 	return fmt.Errorf("EnvoyConfiguration.createLoadAssignment: failed to validate clusterLoadAssignment=%v, %w", upstreamEndpoint, err)
	// }

	return upstreamEndpoint
}

func (e *EnvoyConfiguration) GenerateSnapshot() (*cache.Snapshot, error) {
	if err := e.listener.ValidateAll(); err != nil {
		return nil, fmt.Errorf("EnvoyConfiguration.GenerateSnapshot: failed to validate listener=%v, %w", e.listener, err)
	}

	clusters := []envoy_cache_types.Resource{}
	for _, cluster := range e.clusters {
		if err := cluster.ValidateAll(); err != nil {
			return nil, fmt.Errorf("EnvoyConfiguration.GenerateSnapshot: failed to validate cluster=%v, %w", cluster, err)
		}
		clusters = append(clusters, cluster)
	}

	secrets := []*envoy_extensions_transport_sockets_tls_v3.Secret{}
	for _, secret := range e.secrets {
		if err := secret.ValidateAll(); err != nil {
			return nil, fmt.Errorf("EnvoyConfiguration.GenerateSnapshot: failed to validate secret=%v, %w", secret, err)
		}
		secrets = append(secrets, secret)
	}

	// We're using uuid V1 to provide time sortable snapshot version
	snapshotVersion, err := uuid.NewV1()
	if err != nil {
		return nil, fmt.Errorf("EnvoyConfiguration.GenerateSnapshot: failed on uuid.NewV1, %w", err)
	}

	newVersion := snapshotVersion.String()
	e.logger.Info("GenerateSnapshot: generating new snapshot", "newVersion", newVersion)

	snap, err := cache.NewSnapshot(newVersion,
		map[envoy_resource_v3.Type][]envoy_cache_types.Resource{
			envoy_resource_v3.ClusterType:  clusters,
			envoy_resource_v3.RouteType:    {e.makeRouteConfiguration(RouteName)},
			envoy_resource_v3.ListenerType: {e.listener},
			envoy_resource_v3.SecretType:   asResources(secrets),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("EnvoyConfiguration.GenerateSnapshot: failed to create new snapshot, newVersion=%v, %w", newVersion, err)
	}

	return snap, snap.Consistent()
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

// AddGenericSecretString
// Add a generic secrets as a string. This should be used to add the `token_secret`, also known as `TokenSecretType`.
func (e *EnvoyConfiguration) AddGenericSecretString(inlineString string) (string, error) {
	const SecretType = types.TokenSecretType

	name, err := e.generateSecretName(SecretType)
	if err != nil {
		return "", fmt.Errorf("EnvoyConfiguration.AddGenericSecretBytes: failed to generate %q name, %w", SecretType, err)
	}
	secret := &envoy_extensions_transport_sockets_tls_v3.Secret{
		Name: name,
		Type: &envoy_extensions_transport_sockets_tls_v3.Secret_GenericSecret{
			GenericSecret: &envoy_extensions_transport_sockets_tls_v3.GenericSecret{
				Secret: &envoy_config_core_v3.DataSource{
					Specifier: &envoy_config_core_v3.DataSource_InlineString{
						InlineString: inlineString,
					},
				},
			},
		},
	}
	return e.addGenericSecret(SecretType, secret)
}

// AddGenericSecretBytes
// Add a generic secrets as a string. This should be used to add the `hmac_secret`, also known as `HMACSecretType`.
func (e *EnvoyConfiguration) AddGenericSecretBytes(inlineBytes []byte) (string, error) {
	const SecretType = types.HMACSecretType

	name, err := e.generateSecretName(SecretType)
	if err != nil {
		return "", fmt.Errorf("EnvoyConfiguration.AddGenericSecretBytes: failed to generate %q name, %w", SecretType, err)
	}
	secret := &envoy_extensions_transport_sockets_tls_v3.Secret{
		Name: name,
		Type: &envoy_extensions_transport_sockets_tls_v3.Secret_GenericSecret{
			GenericSecret: &envoy_extensions_transport_sockets_tls_v3.GenericSecret{
				Secret: &envoy_config_core_v3.DataSource{
					Specifier: &envoy_config_core_v3.DataSource_InlineBytes{
						InlineBytes: inlineBytes,
					},
				},
			},
		},
	}
	return e.addGenericSecret(SecretType, secret)
}

// addGenericSecret
// The `name` field in the secret along with a uuid-v4 is used as the key to value in the map.
func (e *EnvoyConfiguration) addGenericSecret(secretType types.SecretType, secret *envoy_extensions_transport_sockets_tls_v3.Secret) (string, error) {
	if secret == nil {
		return "", fmt.Errorf("EnvoyConfiguration.addGenericSecret: %q is nil", secretType)
	}

	if err := secret.ValidateAll(); err != nil {
		return "", fmt.Errorf("EnvoyConfiguration.addGenericSecret: %q failed validation %v, %w", secretType, spew.Sprint(secret), err)
	}

	key := secret.Name
	e.secrets[key] = secret
	e.logger.Info("AddClientSecret: added secret", "key", key, "secret", spew.Sprint(secret), "secretType", secretType)

	return key, nil
}

// addGenericSecret
// The `name` field in the secret along with a uuid-v4 is used as the key to value in the map.
func (e *EnvoyConfiguration) generateSecretName(secretType types.SecretType) (string, error) {
	uuid, err := google_uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("EnvoyConfiguration.generateSecretName: failed to get UUID for %q, %w", secretType, err)
	}

	keyPrefix := ""
	if secretType == types.TokenSecretType {
		keyPrefix = "token-secret"
	} else if secretType == types.HMACSecretType {
		keyPrefix = "hmac-secret"
	}

	return fmt.Sprintf("%s-%s", keyPrefix, uuid.String()), nil
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
