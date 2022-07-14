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

package parser

import (
	"fmt"
	"time"

	//v3 "github.com/cncf/xds/go/xds/core/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_config_filter_http_ext_authz_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	envoy_config_filter_http_oauth2_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/oauth2/v3"
	envoy_extensions_filters_http_router_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	http "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_extensions_transport_sockets_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_type_matcher_v3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/google/uuid"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/kubeshop/kusk-gateway/pkg/options"
)

const (
	ResourceApiVersion = envoy_config_core_v3.ApiVersion_V3
)

// https://github.com/envoyproxy/envoy/tree/main/examples/ext_authz
// https://github.com/envoyproxy/envoy/blob/main/docs/root/configuration/http/http_filters/ext_authz_filter.rst
// https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_authz_filter#config-http-filters-ext-authz
func NewFilterHTTPExternalAuthorization(upstreamHostname string, upstreamPort uint32, clusterName string, pathPrefix string) (*anypb.Any, error) {
	const (
		transportApiVersion = envoy_config_core_v3.ApiVersion_V3
	)

	uri := fmt.Sprintf("%s:%d", upstreamHostname, upstreamPort)

	httpUpstreamType := &envoy_config_core_v3.HttpUri_Cluster{
		Cluster: clusterName,
	}
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/http_uri.proto
	serverUri := &envoy_config_core_v3.HttpUri{
		Uri:              uri,
		HttpUpstreamType: httpUpstreamType,
		Timeout:          TimeoutDefault(),
	}
	authorizationResponse := &envoy_config_filter_http_ext_authz_v3.AuthorizationResponse{
		AllowedUpstreamHeaders: &envoy_type_matcher_v3.ListStringMatcher{
			Patterns: []*envoy_type_matcher_v3.StringMatcher{
				StringMatcherContains("x-current-user", true),
			},
		},
	}
	httpService := &envoy_config_filter_http_ext_authz_v3.HttpService{
		ServerUri:             serverUri,
		PathPrefix:            pathPrefix,
		AuthorizationResponse: authorizationResponse,
	}
	services := &envoy_config_filter_http_ext_authz_v3.ExtAuthz_HttpService{
		HttpService: httpService,
	}
	withRequestBody := &envoy_config_filter_http_ext_authz_v3.BufferSettings{
		MaxRequestBytes:     10,
		AllowPartialMessage: true,
		PackAsBytes:         true,
	}
	authorization := &envoy_config_filter_http_ext_authz_v3.ExtAuthz{
		Services:               services,
		TransportApiVersion:    transportApiVersion,
		IncludePeerCertificate: true,
		WithRequestBody:        withRequestBody,
		StatusOnError:          &envoy_type.HttpStatus{Code: envoy_type.StatusCode_Forbidden},
		// Pretty sure we always want this. Why have an
		// external auth service if it is not going to affect
		// routing decisions?
		ClearRouteCache: false,
	}

	anyAuthorization, err := anypb.New(authorization)
	if err != nil {
		return nil, fmt.Errorf("parser.NewFilterHTTPExternalAuthorization: cannot marshal filter authorization=%+#v, %w", authorization, err)
	}

	return anyAuthorization, nil
}

func NewFilterHTTPOAuth2(oauth2 *options.OAuth2) (*anypb.Any, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	clusterName := fmt.Sprintf("%s-%s", "token_endpoint", uuid.String())
	httpUpstreamType := &envoy_config_core_v3.HttpUri_Cluster{
		Cluster: clusterName,
	}
	tokenEndpoint := &envoy_config_core_v3.HttpUri{
		Uri:              oauth2.TokenEndpoint,
		HttpUpstreamType: httpUpstreamType,
		Timeout:          TimeoutDefault(),
	}
	authorizationEndpoint := oauth2.AuthorizationEndpoint

	tokenSecret := &envoy_extensions_transport_sockets_tls_v3.SdsSecretConfig{
		Name: "token_secret",
		SdsConfig: &envoy_config_core_v3.ConfigSource{
			// Authorities:           []*v3.Authority{},
			ConfigSourceSpecifier: nil,
			InitialFetchTimeout:   TimeoutDefault(),
			ResourceApiVersion:    ResourceApiVersion,
		},
	}
	// tokenFormation := &envoy_config_filter_http_oauth2_v3.OAuth2Credentials_HmacSecret{
	// 	HmacSecret: &envoy_extensions_transport_sockets_tls_v3.SdsSecretConfig{
	// 		Name: "hmac_secret",
	// 		SdsConfig: &envoy_config_core_v3.ConfigSource{
	// 			Authorities:           []*v3.Authority{},
	// 			ConfigSourceSpecifier: nil,
	// 			InitialFetchTimeout:   TimeoutDefault(),
	// 			ResourceApiVersion:    resourceApiVersion,
	// 		},
	// 	},
	// }
	cookieNames := &envoy_config_filter_http_oauth2_v3.OAuth2Credentials_CookieNames{
		// Cookie name to hold OAuth bearer token value. When the authentication server validates the
		// client and returns an authorization token back to the OAuth filter, no matter what format
		// that token is, if :ref:`forward_bearer_token <envoy_v3_api_field_extensions.filters.http.oauth2.v3.OAuth2Config.forward_bearer_token>`
		// is set to true the filter will send over the bearer token as a cookie with this name to the
		// upstream. Defaults to ``BearerToken``.
		BearerToken: "BearerToken",
		// Cookie name to hold OAuth HMAC value. Defaults to ``OauthHMAC``.
		OauthHmac: "OauthHMAC",
		// Cookie name to hold OAuth expiry value. Defaults to ``OauthExpires``.
		OauthExpires: "OauthExpires",
	}
	credentials := &envoy_config_filter_http_oauth2_v3.OAuth2Credentials{
		// The client_id to be used in the authorize calls. This value will be URL encoded when sent to the OAuth server.
		ClientId: oauth2.Credentials.ClientID,
		// The secret used to retrieve the access token. This value will be URL encoded when sent to the OAuth server.
		TokenSecret: tokenSecret,
		// // Configures how the secret token should be created.
		// //
		// // Types that are assignable to TokenFormation:
		// //	*OAuth2Credentials_HmacSecret
		// TokenFormation: tokenFormation,
		// The cookie names used in OAuth filters flow.
		CookieNames: cookieNames,
	}

	redirectUri := oauth2.RedirectURI
	redirectPathMatcher := PathMatcherContains(oauth2.RedirectPathMatcher)
	signoutPath := PathMatcherContains(oauth2.SignoutPath)
	forwardBearerToken := oauth2.ForwardBearerToken
	passThroughMatcher := []*envoy_config_route_v3.HeaderMatcher{
		{},
	}
	authScopes := oauth2.AuthScopes
	resources := oauth2.Resources

	config := &envoy_config_filter_http_oauth2_v3.OAuth2Config{
		// Endpoint on the authorization server to retrieve the access token from.
		TokenEndpoint: tokenEndpoint,
		// The endpoint redirect to for authorization in response to unauthorized requests.
		AuthorizationEndpoint: authorizationEndpoint,
		// Credentials used for OAuth.
		Credentials: credentials,
		// The redirect URI passed to the authorization endpoint. Supports header formatting
		// tokens. For more information, including details on header value syntax, see the
		// documentation on :ref:`custom request headers <config_http_conn_man_headers_custom_request_headers>`.
		//
		// This URI should not contain any query parameters.
		RedirectUri: redirectUri,
		// Matching criteria used to determine whether a path appears to be the result of a redirect from the authorization server.
		RedirectPathMatcher: redirectPathMatcher,
		// The path to sign a user out, clearing their credential cookies.
		SignoutPath: signoutPath,
		// Forward the OAuth token as a Bearer to upstream web service.
		ForwardBearerToken: forwardBearerToken,
		// Any request that matches any of the provided matchers will be passed through without OAuth validation.
		PassThroughMatcher: passThroughMatcher,
		// Optional list of OAuth scopes to be claimed in the authorization request. If not specified,
		// defaults to "user" scope.
		// OAuth RFC https://tools.ietf.org/html/rfc6749#section-3.3
		AuthScopes: authScopes,
		// Optional resource parameter for authorization request
		// RFC: https://tools.ietf.org/html/rfc8707
		Resources: resources,
	}
	oAuth2 := &envoy_config_filter_http_oauth2_v3.OAuth2{
		// Leave this empty to disable OAuth2 for a specific route, using per filter config.
		Config: config,
	}

	anyOAuth2, err := anypb.New(oAuth2)
	if err != nil {
		return nil, fmt.Errorf("parsers.NewFilterHTTPOAuth2: cannot marshal filter oAuth2=%+#v, %w", oAuth2, err)
	}

	return anyOAuth2, nil
}

const (
	localhost = "127.0.0.1"

	// XdsCluster is the cluster name for the control server (used by non-ADS set-up).
	XdsCluster = "xds_cluster"

	// AlsCluster is the clustername for gRPC access log service (ALS)
	AlsCluster = "als_cluster"

	// Ads mode for resources: one aggregated xDS service
	Ads = "ads"

	// Xds mode for resources: individual xDS services.
	Xds = "xds"

	// Rest mode for resources: polling using Fetch.
	Rest = "rest"

	// Delta mode for resources: individual delta xDS services.
	Delta = "delta"

	// Delta Ads mode for resource: one aggregated delta xDS service.
	DeltaAds = "delta-ads"
)

var (
	// RefreshDelay for the polling config source.
	RefreshDelay = 500 * time.Millisecond
)

// data source configuration
func configSource(mode string) *envoy_config_core_v3.ConfigSource {
	source := &envoy_config_core_v3.ConfigSource{}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	switch mode {
	case Ads:
		source.ConfigSourceSpecifier = &envoy_config_core_v3.ConfigSource_Ads{
			Ads: &envoy_config_core_v3.AggregatedConfigSource{},
		}
	case DeltaAds:
		source.ConfigSourceSpecifier = &envoy_config_core_v3.ConfigSource_Ads{
			Ads: &envoy_config_core_v3.AggregatedConfigSource{},
		}
	case Xds:
		source.ConfigSourceSpecifier = &envoy_config_core_v3.ConfigSource_ApiConfigSource{
			ApiConfigSource: &envoy_config_core_v3.ApiConfigSource{
				TransportApiVersion:       ResourceApiVersion,
				ApiType:                   envoy_config_core_v3.ApiConfigSource_GRPC,
				SetNodeOnFirstMessageOnly: true,
				GrpcServices: []*envoy_config_core_v3.GrpcService{{
					TargetSpecifier: &envoy_config_core_v3.GrpcService_EnvoyGrpc_{
						EnvoyGrpc: &envoy_config_core_v3.GrpcService_EnvoyGrpc{ClusterName: XdsCluster},
					},
				}},
			},
		}
	case Rest:
		source.ConfigSourceSpecifier = &envoy_config_core_v3.ConfigSource_ApiConfigSource{
			ApiConfigSource: &envoy_config_core_v3.ApiConfigSource{
				ApiType:             envoy_config_core_v3.ApiConfigSource_REST,
				TransportApiVersion: ResourceApiVersion,
				ClusterNames:        []string{XdsCluster},
				RefreshDelay:        durationpb.New(RefreshDelay),
			},
		}
	case Delta:
		source.ConfigSourceSpecifier = &envoy_config_core_v3.ConfigSource_ApiConfigSource{
			ApiConfigSource: &envoy_config_core_v3.ApiConfigSource{
				TransportApiVersion:       ResourceApiVersion,
				ApiType:                   envoy_config_core_v3.ApiConfigSource_DELTA_GRPC,
				SetNodeOnFirstMessageOnly: true,
				GrpcServices: []*envoy_config_core_v3.GrpcService{{
					TargetSpecifier: &envoy_config_core_v3.GrpcService_EnvoyGrpc_{
						EnvoyGrpc: &envoy_config_core_v3.GrpcService_EnvoyGrpc{ClusterName: XdsCluster},
					},
				}},
			},
		}
	}
	return source
}

func IsRouterFilter(filter *http.HttpFilter) bool {
	return filter.GetTypedConfig().MessageIs(&envoy_extensions_filters_http_router_v3.Router{}) || filter.Name == wellknown.Router
}

func TimeoutDefault() *durationpb.Duration {
	// return &durationpb.Duration{
	// 	Seconds: 120,
	// }
	// Envoy will wait indefinitely for the first xDS config.
	return durationpb.New(time.Second * 0)
}
