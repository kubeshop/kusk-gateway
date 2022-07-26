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

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_auth_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	envoy_type_matcher_v3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	envoy_type_v3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

// https://github.com/envoyproxy/envoy/tree/main/examples/ext_authz
// https://github.com/envoyproxy/envoy/blob/main/docs/root/configuration/http/http_filters/ext_authz_filter.rst
// https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_authz_filter#config-http-filters-ext-authz
func NewHTTPExternalAuthorizationFilter(upstreamHostname string, upstreamPort uint32, clusterName string, pathPrefix string, authHeaders []*envoy_config_core_v3.HeaderValue) (*anypb.Any, error) {
	uri := fmt.Sprintf("%s:%d", upstreamHostname, upstreamPort)

	httpUpstreamType := &envoy_config_core_v3.HttpUri_Cluster{
		Cluster: clusterName,
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

	var authorizationRequest *envoy_auth_v3.AuthorizationRequest
	if len(authHeaders) != 0 {
		authorizationRequest = &envoy_auth_v3.AuthorizationRequest{
			AllowedHeaders: &envoy_type_matcher_v3.ListStringMatcher{
				Patterns: []*envoy_type_matcher_v3.StringMatcher{
					{
						MatchPattern: &envoy_type_matcher_v3.StringMatcher_SafeRegex{
							SafeRegex: &envoy_type_matcher_v3.RegexMatcher{
								EngineType: &envoy_type_matcher_v3.RegexMatcher_GoogleRe2{
									GoogleRe2: &envoy_type_matcher_v3.RegexMatcher_GoogleRE2{},
								},
								Regex: ".*",
							},
						},
					},
				},
			},
			HeadersToAdd: authHeaders,
		}
	}

	httpService := &envoy_auth_v3.HttpService{
		ServerUri:             serverUri,
		PathPrefix:            pathPrefix,
		AuthorizationRequest:  authorizationRequest,
		AuthorizationResponse: authorizationResponse,
	}
	services := &envoy_auth_v3.ExtAuthz_HttpService{
		HttpService: httpService,
	}
	authorization := &envoy_auth_v3.ExtAuthz{
		Services:            services,
		TransportApiVersion: envoy_config_core_v3.ApiVersion_V3,
		StatusOnError: &envoy_type_v3.HttpStatus{
			Code: envoy_type_v3.StatusCode_ServiceUnavailable,
		},
	}
	anyAuthorization, err := anypb.New(authorization)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal authorization=%+v configuration: %w", authorization, err)
	}

	return anyAuthorization, nil
}
