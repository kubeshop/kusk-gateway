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

package controllers

import (
	"fmt"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_auth_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	envoy_hcm_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_type_matcher_v3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/kubeshop/kusk-gateway/pkg/options"
)

func fromAuthOptionsToHttpService(auth *options.Auth) *envoy_auth_v3.HttpService {
	// Hardcoding for now to see if it actually works.
	uri := fmt.Sprintf("http://%s:%d", auth.AuthUpstream.Host.Hostname, auth.AuthUpstream.Host.Port)

	pathPrefix := ""
	if auth.PathPrefix != nil {
		pathPrefix = *auth.PathPrefix
	}

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

	_ = envoy_hcm_v3.HttpFilter_TypedConfig{}

	return httpService
}

func configureAuthz(filterConf map[string]*anypb.Any, authOptions *options.Auth) error {
	// // Do nothing for now.
	// return nil

	if authOptions != nil {
		httpService := fromAuthOptionsToHttpService(authOptions)
		anyHTTPService, err := anypb.New(httpService)
		if err != nil {
			return fmt.Errorf("configureAuthz: failure marshalling `auth` configuration: %w ", err)
		}

		// "envoy.filters.http.ext_authz"
		filterName := wellknown.HTTPExternalAuthorization
		filterConf[filterName] = anyHTTPService
		// filterConf["envoy.filters.http.ext_authz.v3.HttpService"] = anyHTTPService
	}

	return nil
}
