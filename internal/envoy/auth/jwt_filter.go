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
	"net/url"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	envoy_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_jwt_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/jwt_authn/v3"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/kubeshop/kusk-gateway/pkg/options"
)

// See https://github.com/projectcontour/contour/blob/main/internal/envoy/v3/listener.go#L746 for example usage of the filter.

func NewFilterHTTPJWT(jwtOptions *options.JWT, args *ParseAuthArguments, paths []string) (*anypb.Any, error) {
	if len(jwtOptions.JWTProviders) == 0 {
		return nil, nil
	}

	jwtConfig := envoy_jwt_v3.JwtAuthentication{
		Providers:      map[string]*envoy_jwt_v3.JwtProvider{},
		RequirementMap: map[string]*envoy_jwt_v3.JwtRequirement{},
		Rules:          []*envoy_jwt_v3.RequirementRule{},
	}

	for _, provider := range jwtOptions.JWTProviders {
		// Default cache timeout.
		cacheDuration := durationpb.New(time.Second * 300)
		// Default Remote JWKS timeout.
		timeout := durationpb.New(time.Second * 1)

		cluster := provider.JWKS
		if !args.EnvoyConfiguration.ClusterExist(cluster) {
			url, err := url.Parse(cluster)
			if err != nil {
				return nil, err
			}

			upstreamServiceHost := url.Hostname()
			upstreamServicePort := uint32(443)

			port := url.Port()
			if port != "" {
				port, err := strconv.ParseUint(port, 10, 32)
				if err != nil {
					return nil, err
				}

				upstreamServicePort = uint32(port)
			}

			args.Logger.Info("NewFilterHTTPJWT: adding cluster", "cluster", cluster, "upstreamServiceHost", upstreamServiceHost, "upstreamServicePort", upstreamServicePort)
			if err := args.EnvoyConfiguration.AddClusterWithTLS(cluster, upstreamServiceHost, upstreamServicePort); err != nil {
				return nil, err
			}
		}

		jwtConfig.Providers[provider.Name] = &envoy_jwt_v3.JwtProvider{
			Issuer:    provider.Issuer,
			Audiences: provider.Audiences,
			JwksSourceSpecifier: &envoy_jwt_v3.JwtProvider_RemoteJwks{
				RemoteJwks: &envoy_jwt_v3.RemoteJwks{
					HttpUri: &envoy_core_v3.HttpUri{
						Uri: cluster,
						HttpUpstreamType: &envoy_core_v3.HttpUri_Cluster{
							Cluster: cluster,
							// Cluster: DNSNameClusterName(&provider.RemoteJWKS.Cluster),
						},
						Timeout: timeout,
					},
					CacheDuration: cacheDuration,
				},
			},
			Forward: provider.ForwardJWT,
		}

		// Set up a requirement map so that per-route filter config can refer
		// to a requirement by name. This is nicer than specifying rules here,
		// because it likely results in less Envoy config overall (don't have
		// to duplicate every route match in the jwt_authn config), and it means
		// we don't have to implement another sorter to sort JWT rules -- the
		// sorting already being done to routes covers it.
		jwtConfig.RequirementMap[provider.Name] = &envoy_jwt_v3.JwtRequirement{
			RequiresType: &envoy_jwt_v3.JwtRequirement_ProviderName{
				ProviderName: provider.Name,
			},
		}

		args.Logger.Info(
			"NewFilterHTTPJWT: adding `paths` to `rules`",
			"paths", spew.Sprint(paths),
			"providerName", provider.Name,
		)

		for _, path := range paths {
			jwtConfig.Rules = append(jwtConfig.Rules, &envoy_jwt_v3.RequirementRule{
				Match: &envoy_config_route_v3.RouteMatch{
					PathSpecifier: &envoy_config_route_v3.RouteMatch_Path{
						Path: path,
					},
				},
				RequirementType: &envoy_jwt_v3.RequirementRule_Requires{
					Requires: &envoy_jwt_v3.JwtRequirement{
						RequiresType: &envoy_jwt_v3.JwtRequirement_ProviderAndAudiences{
							ProviderAndAudiences: &envoy_jwt_v3.ProviderWithAudiences{
								ProviderName: provider.Name,
								Audiences:    provider.Audiences,
							},
						},
					},
				},
			})
		}
	}

	if err := jwtConfig.ValidateAll(); err != nil {
		return nil, err
	}

	return anypb.New(&jwtConfig)
}
