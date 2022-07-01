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

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_extensions_filters_network_http_connection_manager_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"

	"github.com/kubeshop/kusk-gateway/internal/cloudentity"
	"github.com/kubeshop/kusk-gateway/pkg/options"
)

func ParseAuthUpstreamOptions(authUpstreamOptions *options.AuthUpstream, args *parseAuthOptionsArguments, scheme string) error {
	upstreamServiceHost := authUpstreamOptions.Host.Hostname
	upstreamServicePort := authUpstreamOptions.Host.Port

	clusterName := args.GenerateClusterName(upstreamServiceHost, upstreamServicePort)

	var authHeaders []*envoy_config_core_v3.HeaderValue
	if scheme == options.SchemeCloudEntity {
		var (
			// fetch auth service host and port once
			// TODO: fetch kusk gateway auth service dynamically
			cloudEntityHostname string = "kusk-gateway-auth-service.kusk-system.svc.cluster.local."
			cloudEntityPort     uint32 = 19000
		)

		args.CloudEntityBuilder.AddAPI(upstreamServiceHost, upstreamServicePort, args.Name, args.Name, args.RoutePath, args.Method)
		authHeaders = []*envoy_config_core_v3.HeaderValue{
			{
				Key:   cloudentity.HeaderAuthorizerURL,
				Value: fmt.Sprintf("https://%s:%d", upstreamServiceHost, upstreamServicePort),
			},
			{
				Key:   cloudentity.HeaderAPIGroup,
				Value: args.Name,
			},
		}
		upstreamServiceHost = cloudEntityHostname
		upstreamServicePort = cloudEntityPort
	}

	if !args.EnvoyConfiguration.ClusterExist(clusterName) {
		args.EnvoyConfiguration.AddCluster(
			clusterName,
			upstreamServiceHost,
			upstreamServicePort,
		)
	}

	pathPrefix := ""
	if authUpstreamOptions.PathPrefix != nil {
		pathPrefix = *authUpstreamOptions.PathPrefix
	}

	typedConfig, err := NewFilterHTTPExternalAuthorizationFilter(
		upstreamServiceHost,
		upstreamServicePort,
		clusterName,
		pathPrefix,
		authHeaders,
	)
	if err != nil {
		return err
	}

	filter := &envoy_extensions_filters_network_http_connection_manager_v3.HttpFilter{
		Name: wellknown.HTTPExternalAuthorization,
		ConfigType: &envoy_extensions_filters_network_http_connection_manager_v3.HttpFilter_TypedConfig{
			TypedConfig: typedConfig,
		},
	}

	return args.HTTPConnectionManagerBuilder.AddFilter(filter)
}
