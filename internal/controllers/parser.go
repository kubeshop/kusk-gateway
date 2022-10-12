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

package controllers

import (
	"fmt"
	"strconv"
	"strings"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubeshop/kusk-gateway/internal/cloudentity"
	"github.com/kubeshop/kusk-gateway/internal/envoy/auth"
	"github.com/kubeshop/kusk-gateway/internal/envoy/config"
	"github.com/kubeshop/kusk-gateway/internal/envoy/cors"
	"github.com/kubeshop/kusk-gateway/internal/envoy/types"
	"github.com/kubeshop/kusk-gateway/internal/mocking"
	"github.com/kubeshop/kusk-gateway/internal/validation"
	"github.com/kubeshop/kusk-gateway/pkg/options"
	parseSpec "github.com/kubeshop/kusk-gateway/pkg/spec"
)

/* This is the copy of https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/route_matching to remind how Envoy matches the route.
When Envoy matches a route, it uses the following procedure:

    The HTTP request’s Host or :authority header is matched to a virtual host.

    Each route entry in the virtual host is checked, in order. If there is a match, the route is used and no further route checks are made.

    Independently, each virtual cluster in the virtual host is checked, in order. If there is a match, the virtual cluster is used and no further virtual cluster checks are made.

From the istio issue tracker:

The virtual hosts order does not influence the domain matching order

It is the domain matters

Domain search order:
1. Exact domain names: www.foo.com.
2. Suffix domain wildcards: *.foo.com or *-bar.foo.com.
3. Prefix domain wildcards: foo.* or foo-*.
4. Special wildcard * matching any domain.
*/

// UpdateConfigFromAPIOpts updates Envoy configuration from OpenAPI spec and x-kusk options
func UpdateConfigFromAPIOpts(
	envoyConfiguration *config.EnvoyConfiguration,
	proxy validation.ValidationUpdater,
	opts *options.Options,
	spec *openapi3.T,
	httpConnectionManagerBuilder *config.HCMBuilder,
	clBuilder *cloudentity.Builder,
	name string,
	kubernetesClient client.Client,
) error {
	logger := ctrl.Log.WithName("internal/controllers/parser.go:UpdateConfigFromAPIOpts")

	// Add new vhost if already not present.
	for _, vhost := range opts.Hosts {
		if envoyConfiguration.GetVirtualHost(string(vhost)) == nil {
			vh := types.NewVirtualHost(string(vhost))
			// Add the same domain as virtual host
			vh.AddDomain(string(vhost))
			envoyConfiguration.AddVirtualHost(vh)
		}
	}

	// store proxied services in map to de-duplicate
	proxiedServices := map[string]*validation.Service{}

	// Iterate on all paths and build routes
	// The overriding works in the following way:
	// 1. For each path we get SubOptions from the opts map and merge in top level SubOpts
	// 2. For each method we get SubOptions for that method from the opts map and merge in path SubOpts
	for path, pathItem := range spec.Paths {
		// x-kusk options per operation (http method)
		for method, operation := range pathItem.Operations() {
			finalOpts := opts.OperationFinalSubOptions[method+path]
			if finalOpts.Disabled != nil && *finalOpts.Disabled {
				continue
			}

			var routesToAddToVirtualHost []*route.Route

			routePath := path
			if finalOpts.Path != nil {
				routePath = generateRoutePath(finalOpts.Path.Prefix, path)
			}

			corsPolicy, err := generateCORSPolicy(finalOpts.CORS)
			if err != nil {
				return err
			}

			rt := &route.Route{
				Name: types.GenerateRouteName(routePath, method),
				Match: generateRouteMatch(
					routePath,
					method,
					extractParams(operation.Parameters),
					corsPolicy,
				),
			}

			if opts.CORS != nil {
				cors.ConfigureCORSOnRoute(logger, corsPolicy, rt, opts.CORS.Origins)
			}

			if finalOpts.Cache != nil && finalOpts.Cache.Enabled != nil && *finalOpts.Cache.Enabled {
				rt.ResponseHeadersToAdd = append(rt.ResponseHeadersToAdd, &envoy_config_core_v3.HeaderValueOption{
					Header: &envoy_config_core_v3.HeaderValue{
						Key:   "Cache-Control",
						Value: "max-age=" + strconv.Itoa(*finalOpts.Cache.MaxAge),
					},
					Append: wrapperspb.Bool(false),
				},
				)
			}

			if finalOpts.Auth != nil {
				logger.Info("parsing `auth` options", "`auth`", finalOpts.Auth)
				arguments := auth.NewParseAuthOptionsArguments(
					ctrl.Log,
					envoyConfiguration,
					httpConnectionManagerBuilder,
					name,
					routePath,
					method,
					clBuilder,
					generateClusterName, // each cluster can be uniquely identified by dns name + port (i.e. canonical Host, which is hostname:port)
					kubernetesClient,
				)

				err := auth.ParseAuthOptions(finalOpts, arguments)
				if err != nil {
					return err
				}
			}

			// // Validate and Proxy to the upstream
			if finalOpts.Validation != nil && finalOpts.Validation.Request != nil && finalOpts.Validation.Request.Enabled != nil && *finalOpts.Validation.Request.Enabled {
				var (
					upstreamHostname string
					upstreamPort     uint32
				)
				if finalOpts.Mocking != nil && *finalOpts.Mocking.Enabled {
					upstreamHostname = types.GenerateRouteName(routePath, method)
					upstreamPort = 0
				} else {
					hostPortPair, err := getUpstreamHost(finalOpts.Upstream)
					if err != nil {
						return err
					}
					upstreamHostname = hostPortPair.Host
					upstreamPort = hostPortPair.Port
				}
				// create proxied service if needed
				serviceID := validation.GenerateServiceID(upstreamHostname, upstreamPort)
				if _, ok := proxiedServices[serviceID]; !ok {
					proxiedService, err := validation.NewService(serviceID, upstreamHostname, upstreamPort, spec, opts)
					if err != nil {
						return fmt.Errorf("failed to create proxied service: %w", err)
					}

					proxiedServices[serviceID] = proxiedService
				}

				// attach service id and operation id headers so that validator will know which service should
				// serve this request
				operationID := validation.GenerateOperationID(method, path)

				headers := []*envoy_config_core_v3.HeaderValue{
					{
						Key:   validation.HeaderOperationName,
						Value: "validate",
					},
					{
						Key:   validation.HeaderOperationID,
						Value: operationID,
					},
					{
						Key:   validation.HeaderServiceID,
						Value: serviceID,
					},
				}

				extProc := mapExternalProcessorConfig(headers)

				anyExtProc, err := anypb.New(extProc)
				if err != nil {
					return fmt.Errorf("failure marshalling ext_proc configuration: %w ", err)
				}

				if rt.TypedPerFilterConfig == nil {
					rt.TypedPerFilterConfig = make(map[string]*any.Any)
				}
				rt.TypedPerFilterConfig[WellknownExtProc] = anyExtProc
			}

			// Create the decision what to do with the request, in order.
			// Some inherited options might be conflicting, so we implicitly define the decision order - the first detected wins:
			// Redirect -> Mock -> Validate and Proxy to the upstream -> Proxy (Route) to the upstream
			switch {
			// Redirect
			case finalOpts.Redirect != nil:
				routeRedirect, err := generateRedirect(finalOpts.Redirect)
				if err != nil {
					return fmt.Errorf("cannot generate redirect: %w", err)
				}
				rt.Action = routeRedirect

				routesToAddToVirtualHost = append(routesToAddToVirtualHost, rt)
			// Mock
			case finalOpts.Mocking != nil && *finalOpts.Mocking.Enabled:
				for respCode, respRef := range operation.Responses {
					// We don't handle non 2xx codes, skip if found
					if !strings.HasPrefix(respCode, "2") {
						continue
					}
					// Note that we don't handle wildcards, e.g. '2xx' - this is allowed in OpenAPI, but we need the exact status code.
					statusCode, err := strconv.Atoi(respCode)
					if err != nil {
						return fmt.Errorf("cannot convert the response code %s to int: %w", respCode, err)
					}

					// if there are more examples of different content types, require headers
					// to differentiate which should be returned
					requireAcceptHeader := len(respRef.Value.Content) > 1

					for mediaType, mediaTypeValue := range respRef.Value.Content {
						exampleContent := parseSpec.GetExampleResponse(mediaTypeValue)
						if exampleContent == nil {
							continue
						}

						mockedRouteBuilder, err := mocking.NewRouteBuilder(mediaType, rt)
						if err != nil {
							return fmt.Errorf("cannot build mocked route: %w", err)
						}

						mockedRoute, err := mockedRouteBuilder.BuildMockedRoute(&mocking.BuildMockedRouteArgs{
							RoutePath:           routePath,
							Method:              method,
							StatusCode:          uint32(statusCode),
							ExampleContent:      exampleContent,
							RequireAcceptHeader: requireAcceptHeader,
						})
						if err != nil {
							return fmt.Errorf("cannot build mocked route: %w", err)
						}

						routesToAddToVirtualHost = append(routesToAddToVirtualHost, mockedRoute)
					}

					// if there is more than one mediatype, ensure that json is the default when no Accept header passed
					// by appending the match to the end of the chain
					// if no json response present, take the first response in the list and use that as the default
					if requireAcceptHeader {
						singleMediaType := "application/json"
						var singleResponse *openapi3.MediaType

						// Grab the json response if present
						if json, ok := respRef.Value.Content[singleMediaType]; ok {
							singleResponse = json
						} else {
							// Otherwise grab the first response from the list of responses
							for mediaType, mediaTypeValue := range respRef.Value.Content {
								singleMediaType = mediaType
								singleResponse = mediaTypeValue
								break
							}
						}

						exampleContent := parseSpec.GetExampleResponse(singleResponse)
						if exampleContent == nil {
							break
						}

						mockedRouteBuilder, err := mocking.NewRouteBuilder(singleMediaType, rt)
						if err != nil {
							return fmt.Errorf("cannot build mocked route: %w", err)
						}

						mockedRoute, err := mockedRouteBuilder.BuildMockedRoute(&mocking.BuildMockedRouteArgs{
							RoutePath:           routePath,
							Method:              method,
							StatusCode:          uint32(statusCode),
							ExampleContent:      exampleContent,
							RequireAcceptHeader: false,
						})
						if err != nil {
							return fmt.Errorf("cannot build mocked route: %w", err)
						}

						mockedRoute.Name = mockedRoute.Name + "-no-accept-header"
						routesToAddToVirtualHost = append(routesToAddToVirtualHost, mockedRoute)
					}
				}
			// Default - proxy to the upstream
			default:
				hostPortPair, err := getUpstreamHost(finalOpts.Upstream)
				if err != nil {
					return err
				}

				clusterName := generateClusterName(hostPortPair.Host, hostPortPair.Port)
				if !envoyConfiguration.ClusterExist(clusterName) {
					envoyConfiguration.AddCluster(clusterName, hostPortPair.Host, hostPortPair.Port)
				}

				var rewriteOpts *options.RewriteRegex
				if finalOpts.Upstream != nil && finalOpts.Upstream.Rewrite.Pattern != "" {
					rewriteOpts = &finalOpts.Upstream.Rewrite
				}

				routeRoute, err := generateRoute(
					clusterName,
					corsPolicy,
					rewriteOpts,
					finalOpts.QoS,
					finalOpts.Websocket,
				)
				if err != nil {
					return err
				}

				// https://github.com/kubeshop/kusk-gateway/issues/404
				// to help with issues around direct IP access e.g. CloudFlare
				routeRoute.Route.HostRewriteSpecifier = &route.RouteAction_HostRewriteLiteral{
					HostRewriteLiteral: hostPortPair.Host,
				}

				rt.Action = routeRoute

				routesToAddToVirtualHost = append(routesToAddToVirtualHost, rt)
			}

			// For the list of vhosts that we create exactly THIS configuration for, update the routes
			for _, vh := range opts.Hosts {
				for _, rt := range routesToAddToVirtualHost {
					if rt.TypedPerFilterConfig == nil {
						rt.TypedPerFilterConfig = map[string]*any.Any{}
					}

					if finalOpts.RateLimit != nil {
						rl := mapRateLimitConf(finalOpts.RateLimit, generateRateLimitStatPrefix(string(vh), path, method, operation.OperationID))
						anyRateLimit, err := anypb.New(rl)
						if err != nil {
							return fmt.Errorf("failure marshalling ratelimiting configuration: %w ", err)
						}

						rt.TypedPerFilterConfig["envoy.filters.http.local_ratelimit"] = anyRateLimit
					}

					if finalOpts.Auth == nil {
						perRouteAuth, err := auth.RouteAuthzDisabled()
						if err != nil {
							return fmt.Errorf("cannot create per-route config to disable authorization: vh=%q, %w", string(vh), err)
						}

						rt.TypedPerFilterConfig[wellknown.HTTPExternalAuthorization] = perRouteAuth

						logger.Info("disabled `auth` for route", "finalOpts.Auth", finalOpts.Auth, "vh", fmt.Sprintf("%q", string(vh)))
					}

					if finalOpts.Validation == nil || finalOpts.Validation.Request == nil || finalOpts.Validation.Request.Enabled == nil || *finalOpts.Validation.Request.Enabled == false {
						extProc, err := externalProcessorConfigDisabled()
						if err != nil {
							return fmt.Errorf("cannot create per-route config to disable external processing: vh=%q, %w", string(vh), err)
						}

						rt.TypedPerFilterConfig[WellknownExtProc] = extProc
					}

					if err := envoyConfiguration.AddRouteToVHost(string(vh), rt); err != nil {
						return fmt.Errorf("failure adding the route to vhost %s: %w ", string(vh), err)
					}
				}
			}
		}
	}

	if opts.OpenAPIPath != "" {
		for _, vh := range opts.Hosts {
			mockedRouteBuilder, err := mocking.NewRouteBuilder("application/json", &route.Route{})
			if err != nil {
				return fmt.Errorf("cannot build mocked route: %w", err)
			}

			if !strings.HasPrefix(opts.OpenAPIPath, "/") {
				opts.OpenAPIPath = fmt.Sprintf("/%s", opts.OpenAPIPath)
			}
			openapiRt, err := mockedRouteBuilder.BuildMockedRoute(&mocking.BuildMockedRouteArgs{
				RoutePath:      opts.OpenAPIPath,
				Method:         "GET",
				StatusCode:     uint32(200),
				ExampleContent: parseSpec.PostProcessedDef(*spec, *opts),
			})
			if err != nil {
				return fmt.Errorf("cannot build postprocessed api route: %w", err)
			}

			if opts.Auth != nil {
				if openapiRt.TypedPerFilterConfig == nil {
					openapiRt.TypedPerFilterConfig = map[string]*any.Any{}
				}

				perRouteAuth, err := auth.RouteAuthzDisabled()
				if err != nil {
					return fmt.Errorf("cannot create per-route config to disable authorization: openapi-path=%q, %w", opts.OpenAPIPath, err)
				}

				openapiRt.TypedPerFilterConfig[wellknown.HTTPExternalAuthorization] = perRouteAuth

				logger.Info("disabled `auth` for route", "openapi-path", opts.OpenAPIPath, "vh", fmt.Sprintf("%q", string(vh)))
			}

			if err := envoyConfiguration.AddRouteToVHost(string(vh), openapiRt); err != nil {
				return fmt.Errorf("failure adding the route to vhost %s: %w ", string(vh), err)
			}
		}
	}

	// update the validation proxy in the end
	if len(proxiedServices) > 0 {
		proxiedServicesArray := make([]*validation.Service, 0, len(proxiedServices))
		for _, service := range proxiedServices {
			proxiedServicesArray = append(proxiedServicesArray, service)
		}
		proxy.UpdateServices(proxiedServicesArray)
	}

	return nil
}

// UpdateConfigFromOpts updates Envoy configuration from Options only
func UpdateConfigFromOpts(
	envoyConfiguration *config.EnvoyConfiguration,
	opts *options.StaticOptions,
	httpConnectionManagerBuilder *config.HCMBuilder,
	clBuilder *cloudentity.Builder,
	name string,
	kubernetesClient client.Client,
) error {
	logger := ctrl.Log.WithName("internal/controllers/parser.go:UpdateConfigFromOpts")

	// Add new vhost if already not present.
	for _, vhost := range opts.Hosts {
		if envoyConfiguration.GetVirtualHost(string(vhost)) == nil {
			vh := types.NewVirtualHost(string(vhost))
			// Add the same domain as virtual host
			vh.AddDomain(string(vhost))
			envoyConfiguration.AddVirtualHost(vh)
		}
	}

	// Iterate on all paths and build routes
	for path, methods := range opts.Paths {
		for method, methodOpts := range methods {
			strMethod := string(method)

			routePath := generateRoutePath("", path)

			corsPolicy, err := generateCORSPolicy(methodOpts.CORS)
			if err != nil {
				return err
			}

			// routeMatcher defines how we match a route by the provided path and the headers
			rt := &route.Route{
				Name:  types.GenerateRouteName(routePath, strMethod),
				Match: generateRouteMatch(routePath, string(method), nil, corsPolicy),
			}

			if methodOpts.CORS != nil {
				cors.ConfigureCORSOnRoute(logger, corsPolicy, rt, methodOpts.CORS.Origins)
			}

			if methodOpts.Redirect != nil {
				// Generating Redirect
				routeRedirect, err := generateRedirect(methodOpts.Redirect)
				if err != nil {
					return err
				}

				rt.Action = routeRedirect
			} else {
				hostPortPair, err := getUpstreamHost(methodOpts.Upstream)
				if err != nil {
					return err
				}

				clusterName := generateClusterName(hostPortPair.Host, hostPortPair.Port)
				if !envoyConfiguration.ClusterExist(clusterName) {
					envoyConfiguration.AddCluster(clusterName, hostPortPair.Host, hostPortPair.Port)
				}

				var rewriteOpts *options.RewriteRegex
				if methodOpts.Upstream.Rewrite.Pattern != "" {
					rewriteOpts = &methodOpts.Upstream.Rewrite
				}
				routeRoute, err := generateRoute(
					clusterName,
					corsPolicy,
					rewriteOpts,
					methodOpts.QoS,
					methodOpts.Websocket,
				)
				if err != nil {
					return err
				}

				rt.Action = routeRoute
			}

			// For the list of vhosts that we create exactly THIS configuration for, update the routes
			for _, vh := range opts.Hosts {
				if err := envoyConfiguration.AddRouteToVHost(string(vh), rt); err != nil {
					return fmt.Errorf("failure adding the route to vhost %s: %w ", string(vh), err)
				}
			}
		}
	}

	return nil
}
