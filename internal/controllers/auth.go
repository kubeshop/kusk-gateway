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
	auth_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/kubeshop/kusk-gateway/pkg/options"
)

var _ = auth_v3.AuthorizationRequest{}

func convertAuthOptions(auth *options.Auth) *auth_v3.HttpService {
	uri := fmt.Sprintf("http://%s:%d", auth.AuthUpstream.Host.Hostname, auth.AuthUpstream.Host.Port)
	httpUpstreamType := &envoy_config_core_v3.HttpUri_Cluster{}

	pathPrefix := ""
	if auth.PathPrefix != nil {
		pathPrefix = *auth.PathPrefix
	}

	return &auth_v3.HttpService{
		ServerUri: &envoy_config_core_v3.HttpUri{
			// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/http_uri.proto#envoy-v3-api-msg-config-core-v3-httpuri
			Uri:              uri,
			HttpUpstreamType: httpUpstreamType,
			Timeout: &durationpb.Duration{
				Seconds: 60,
			},
		},
		PathPrefix: pathPrefix,
	}
}

func configureAuthz(filterConf map[string]*anypb.Any, authOptions *options.Auth) error {
	// Do nothing for now.
	return nil

	// if authOptions != nil {
	// 	httpService := convertAuthOptions(authOptions)
	// 	anyHTTPService, err := anypb.New(httpService)
	// 	if err != nil {
	// 		return fmt.Errorf("configureAuthz: failure marshalling `auth` configuration: %w ", err)
	// 	}

	// 	filterConf["envoy.filters.http.ext_authz.v3.HttpService"] = anyHTTPService
	// }
}
