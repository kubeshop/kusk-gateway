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
package config

import (
	"fmt"

	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_auth_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	ratelimit "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_type_matcher_v3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	RouteName string = "local_route"
)

type hcmBuilder struct {
	httpConnectionManager *hcm.HttpConnectionManager
}

// x-kusk:
//   auth:
//     scheme: basic
//     auth-upstream:
//       host:
//         hostname: envoy-auth-basic-http-service.svc.cluster.local
//         port: 9092

func NewHCMBuilder() (*hcmBuilder, error) {
	rl := &ratelimit.LocalRateLimit{
		StatPrefix: "http_local_rate_limiter",
	}
	anyRateLimit, err := anypb.New(rl)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal ratelimit configuration: %w", err)
	}

	uri := fmt.Sprintf("http://%s:%d", "envoy-auth-basic-http-service.svc.cluster.local", 9092)

	pathPrefix := ""

	httpUpstreamType := &envoy_config_core_v3.HttpUri_Cluster{
		Cluster: "envoy-auth-basic-http-service",
	}
	serverUri := &envoy_config_core_v3.HttpUri{
		// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/http_uri.proto#envoy-v3-api-msg-config-core-v3-httpuri
		Uri:              uri,
		HttpUpstreamType: httpUpstreamType,
		Timeout: &durationpb.Duration{
			Seconds: 60,
		},
	}
	authorizationResponse := &envoy_auth_v3.AuthorizationResponse{
		AllowedUpstreamHeaders: &envoy_type_matcher_v3.ListStringMatcher{
			Patterns: []*envoy_type_matcher_v3.StringMatcher{
				{
					MatchPattern: &envoy_type_matcher_v3.StringMatcher_Exact{
						Exact: "x-current-user",
					},
					IgnoreCase: true,
				},
			},
		},
	}
	httpService := &envoy_auth_v3.HttpService{
		ServerUri:             serverUri,
		PathPrefix:            pathPrefix,
		AuthorizationResponse: authorizationResponse,
	}
	services := &envoy_auth_v3.ExtAuthz_HttpService{
		HttpService: httpService,
	}
	authorization := &envoy_auth_v3.ExtAuthz{
		Services:            services,
		TransportApiVersion: envoy_config_core_v3.ApiVersion_V3,
	}
	anyAuthorization, err := anypb.New(authorization)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal authorization configuration: %w", err)
	}

	return &hcmBuilder{
		httpConnectionManager: &hcm.HttpConnectionManager{
			CodecType:  hcm.HttpConnectionManager_AUTO,
			StatPrefix: "http",
			RouteSpecifier: &hcm.HttpConnectionManager_Rds{
				Rds: &hcm.Rds{
					ConfigSource:    makeConfigSource(),
					RouteConfigName: RouteName,
				},
			},
			HttpFilters: []*hcm.HttpFilter{
				{
					Name: "envoy.filters.http.local_ratelimit",
					ConfigType: &hcm.HttpFilter_TypedConfig{
						TypedConfig: anyRateLimit,
					},
				},
				{
					Name: wellknown.CORS,
				},
				{
					Name: wellknown.HTTPExternalAuthorization,
					ConfigType: &hcm.HttpFilter_TypedConfig{
						TypedConfig: anyAuthorization,
					},
				},
				{
					Name: wellknown.Router,
				},
			},
		},
	}, nil
}

func (h *hcmBuilder) Validate() error {
	return h.httpConnectionManager.Validate()
}

func (h *hcmBuilder) AddAccessLog(al *accesslog.AccessLog) *hcmBuilder {
	h.httpConnectionManager.AccessLog = append(h.httpConnectionManager.AccessLog, al)
	return h
}

func (h *hcmBuilder) GetHTTPConnectionManager() *hcm.HttpConnectionManager {
	return h.httpConnectionManager
}

func makeConfigSource() *envoy_config_core_v3.ConfigSource {
	source := &envoy_config_core_v3.ConfigSource{}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	source.ConfigSourceSpecifier = &envoy_config_core_v3.ConfigSource_ApiConfigSource{
		ApiConfigSource: &envoy_config_core_v3.ApiConfigSource{
			TransportApiVersion:       resource.DefaultAPIVersion,
			ApiType:                   envoy_config_core_v3.ApiConfigSource_GRPC,
			SetNodeOnFirstMessageOnly: true,
			GrpcServices: []*envoy_config_core_v3.GrpcService{{
				TargetSpecifier: &envoy_config_core_v3.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &envoy_config_core_v3.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"},
				},
			}},
		},
	}
	return source
}
