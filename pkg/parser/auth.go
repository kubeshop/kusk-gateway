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
	envoy_config_filter_http_ext_authz_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/kubeshop/kusk-gateway/internal/envoy/config"
	"github.com/kubeshop/kusk-gateway/pkg/options"
)

// RouteAuthzDisabled returns a per-route config to disable authorization.
func RouteAuthzDisabled() (*anypb.Any, error) {
	return anypb.New(
		&envoy_config_filter_http_ext_authz_v3.ExtAuthzPerRoute{
			Override: &envoy_config_filter_http_ext_authz_v3.ExtAuthzPerRoute_Disabled{
				Disabled: true,
			},
		},
	)
}

// ParseAuthOptions
func ParseAuthOptions(finalOpts options.SubOptions, envoyConfiguration *config.EnvoyConfiguration, httpConnectionManagerBuilder *config.HCMBuilder) error {
	upstreamServiceHost := finalOpts.Auth.AuthUpstream.Host.Hostname
	upstreamServicePort := finalOpts.Auth.AuthUpstream.Host.Port

	clusterName := GenerateClusterName(upstreamServiceHost, upstreamServicePort)

	if !envoyConfiguration.ClusterExist(clusterName) {
		envoyConfiguration.AddCluster(
			clusterName,
			upstreamServiceHost,
			upstreamServicePort,
		)
	}

	pathPrefix := ""
	if finalOpts.Auth.PathPrefix != nil {
		pathPrefix = *finalOpts.Auth.PathPrefix
	}

	httpExternalAuthorizationFilter, err := config.NewHTTPExternalAuthorizationFilter(
		upstreamServiceHost,
		upstreamServicePort,
		clusterName,
		pathPrefix,
	)
	if err != nil {
		return err
	}

	httpConnectionManagerBuilder.AppendFilterHTTPExternalAuthorizationFilterToStart(httpExternalAuthorizationFilter)

	return nil
}
