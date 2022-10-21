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

package auth

import (
	"fmt"

	envoy_extensions_filters_http_ext_authz_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	envoy_extensions_filters_network_http_connection_manager_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/kubeshop/kusk-gateway/pkg/options"
)

func ParseOAuth2Options(oauth2Options *options.OAuth2, arguments *parseAuthOptionsArguments) error {
	typedConfig, err := NewFilterHTTPOAuth2(oauth2Options, arguments)
	if err != nil {
		return err
	}

	filter := &envoy_extensions_filters_network_http_connection_manager_v3.HttpFilter{
		Name: "envoy.filters.http.oauth2",
		ConfigType: &envoy_extensions_filters_network_http_connection_manager_v3.HttpFilter_TypedConfig{
			TypedConfig: typedConfig,
		},
	}

	if err := arguments.HTTPConnectionManagerBuilder.AddFilter(filter); err != nil {
		arguments.Logger.WithName("auth.ParseOAuth2Options").Error(err, "failed to add filter", "filter", fmt.Sprintf("%+#v", filter))
		return err
	}

	return nil
}

// RouteAuthzDisabled
// returns a per-route config to disable authorization.
func RouteAuthzDisabled() (*anypb.Any, error) {
	return anypb.New(
		&envoy_extensions_filters_http_ext_authz_v3.ExtAuthzPerRoute{
			Override: &envoy_extensions_filters_http_ext_authz_v3.ExtAuthzPerRoute_Disabled{
				Disabled: true,
			},
		},
	)
}
