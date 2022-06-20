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
	"strconv"
	"strings"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	ratelimit "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoytypematcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	envoy_type_v3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/getkin/kin-openapi/openapi3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/kubeshop/kusk-gateway/internal/envoy/config"
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
func UpdateConfigFromAPIOpts(envoyConfiguration *config.EnvoyConfiguration, proxy *validation.Proxy, opts *options.Options, spec *openapi3.T) error {
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

	// fetch validation service host and port once
	// TODO: fetch kusk gateway validator service dynamically
	var validatorHostname string = "kusk-gateway-validator-service.kusk-system.svc.cluster.local."
	var validatorPort uint32 = 17000

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
				// TODO: make them compatible
				if finalOpts.Validation != nil && finalOpts.Validation.Request != nil && finalOpts.Validation.Request.Enabled != nil {
					return fmt.Errorf("mocking and validation are enabled but are mutually exclusive")
				}

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

						mockedRouteBuilder, err := mocking.NewRouteBuilder(mediaType)
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

						mockedRouteBuilder, err := mocking.NewRouteBuilder(singleMediaType)
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

			// // Validate and Proxy to the upstream
			case finalOpts.Validation != nil && finalOpts.Validation.Request != nil && finalOpts.Validation.Request.Enabled != nil && *finalOpts.Validation.Request.Enabled:
				upstreamHostname, upstreamPort := getUpstreamHost(finalOpts.Upstream)

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

				rt.RequestHeadersToAdd = append(rt.RequestHeadersToAdd, &envoy_config_core_v3.HeaderValueOption{
					Header: &envoy_config_core_v3.HeaderValue{
						Key:   validation.HeaderServiceID,
						Value: serviceID,
					},
					Append: wrapperspb.Bool(false),
				})

				rt.RequestHeadersToAdd = append(rt.RequestHeadersToAdd, &envoy_config_core_v3.HeaderValueOption{
					Header: &envoy_config_core_v3.HeaderValue{
						Key:   validation.HeaderOperationID,
						Value: operationID,
					},
					Append: wrapperspb.Bool(false),
				})

				clusterName := generateClusterName(validatorHostname, validatorPort)
				if !envoyConfiguration.ClusterExist(clusterName) {
					envoyConfiguration.AddCluster(clusterName, validatorHostname, validatorPort)
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

				rt.Action = routeRoute

				routesToAddToVirtualHost = append(routesToAddToVirtualHost, rt)

			// Default - proxy to the upstream
			default:
				upstreamHostname, upstreamPort := getUpstreamHost(finalOpts.Upstream)

				clusterName := generateClusterName(upstreamHostname, upstreamPort)
				if !envoyConfiguration.ClusterExist(clusterName) {
					envoyConfiguration.AddCluster(clusterName, upstreamHostname, upstreamPort)
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
				rt.Action = routeRoute

				routesToAddToVirtualHost = append(routesToAddToVirtualHost, rt)
			}

			// For the list of vhosts that we create exactly THIS configuration for, update the routes
			for _, vh := range opts.Hosts {
				for _, rt := range routesToAddToVirtualHost {
					filterConf := map[string]*anypb.Any{}

					if finalOpts.RateLimit != nil {
						rl := mapRateLimitConf(finalOpts.RateLimit, generateRateLimitStatPrefix(string(vh), path, method, operation.OperationID))
						anyRateLimit, err := anypb.New(rl)
						if err != nil {
							return fmt.Errorf("failure marshalling ratelimiting configuration: %w ", err)
						}
						filterConf = map[string]*anypb.Any{
							"envoy.filters.http.local_ratelimit": anyRateLimit,
						}
					}
					rt.TypedPerFilterConfig = filterConf

					if err := envoyConfiguration.AddRouteToVHost(string(vh), rt); err != nil {
						return fmt.Errorf("failure adding the route to vhost %s: %w ", string(vh), err)
					}
				}
			}
		}

	}

	if opts.OpenAPIPath != "" {
		for _, vh := range opts.Hosts {

			mockedRouteBuilder, err := mocking.NewRouteBuilder("application/json")
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
				ExampleContent: parseSpec.PostProcessedDef(spec, opts),
			})
			if err != nil {
				return fmt.Errorf("cannot build postprocessed api route: %w", err)
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

// extract Params returns a map mapping the name of a paths parameter to its schema
// where the schema elements we care about are its type and enum if its defined
func extractParams(parameters openapi3.Parameters) map[string]types.ParamSchema {
	params := map[string]types.ParamSchema{}

	for _, parameter := range parameters {
		// Prevent populating map with empty parameter names
		if parameter.Value != nil && parameter.Value.Name != "" {
			params[parameter.Value.Name] = types.ParamSchema{}

			// Extract the schema if it's not nil and assign the map value
			if parameter.Value.Schema != nil && parameter.Value.Schema.Value != nil {
				schemaValue := parameter.Value.Schema.Value

				// It is acceptable for Type and / or Enum to have their zero value
				// It means the user has not defined it, and we will construct the regex path accordingly
				params[fmt.Sprintf("{%s}", parameter.Value.Name)] = types.ParamSchema{
					Type: schemaValue.Type,
					Enum: schemaValue.Enum,
				}
			}
		}
	}

	return params
}

// UpdateConfigFromOpts updates Envoy configuration from Options only
func UpdateConfigFromOpts(envoyConfiguration *config.EnvoyConfiguration, opts *options.StaticOptions) error {
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

			if methodOpts.Redirect != nil {
				// Generating Redirect
				routeRedirect, err := generateRedirect(methodOpts.Redirect)
				if err != nil {
					return err
				}

				rt.Action = routeRedirect
			} else {
				upstreamHostname, upstreamPort := getUpstreamHost(methodOpts.Upstream)
				clusterName := generateClusterName(upstreamHostname, upstreamPort)
				if !envoyConfiguration.ClusterExist(clusterName) {
					envoyConfiguration.AddCluster(clusterName, upstreamHostname, upstreamPort)
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

func generateRouteMatch(path string, method string, pathParameters map[string]types.ParamSchema, corsPolicy *route.CorsPolicy) *route.RouteMatch {
	headerMatcherConfig := []*route.HeaderMatcher{
		types.GetHeaderMatcherConfig([]string{strings.ToUpper(method)}, corsPolicy != nil),
	}

	routeMatcherBuilder := types.NewRouteMatcherBuilder(path, pathParameters)
	return routeMatcherBuilder.GetRouteMatcher(headerMatcherConfig)
}

func generateRedirect(redirectOpts *options.RedirectOptions) (*route.Route_Redirect, error) {
	if redirectOpts == nil {
		return nil, nil
	}

	builder := types.NewRouteRedirectBuilder().
		HostRedirect(redirectOpts.HostRedirect).
		PortRedirect(redirectOpts.PortRedirect).
		SchemeRedirect(redirectOpts.SchemeRedirect).
		PathRedirect(redirectOpts.PathRedirect).
		ResponseCode(redirectOpts.ResponseCode).
		StripQuery(redirectOpts.StripQuery)

	if redirectOpts.RewriteRegex != nil {
		builder = builder.RegexRedirect(redirectOpts.RewriteRegex.Pattern, redirectOpts.RewriteRegex.Substitution)
	}

	redirect, err := builder.ValidateAndReturn()
	if err != nil {
		return nil, err
	}

	return redirect, nil
}

func generateCORSPolicy(corsOpts *options.CORSOptions) (*route.CorsPolicy, error) {
	if corsOpts == nil {
		return nil, nil
	}

	return types.GenerateCORSPolicy(
		corsOpts.Origins,
		corsOpts.Methods,
		corsOpts.Headers,
		corsOpts.ExposeHeaders,
		corsOpts.MaxAge,
		corsOpts.Credentials,
	)
}

func getUpstreamHost(upstreamOpts *options.UpstreamOptions) (hostname string, port uint32) {
	if upstreamOpts.Service != nil {
		return fmt.Sprintf("%s.%s.svc.cluster.local.", upstreamOpts.Service.Name, upstreamOpts.Service.Namespace), upstreamOpts.Service.Port
	}
	return upstreamOpts.Host.Hostname, upstreamOpts.Host.Port
}

// each cluster can be uniquely identified by dns name + port (i.e. canonical Host, which is hostname:port)
func generateClusterName(name string, port uint32) string {
	return fmt.Sprintf("%s-%d", name, port)
}

func generateMockID(path string, method string, operationID string) string {
	return fmt.Sprintf("%s-%s-%s", path, method, operationID)
}

func generateRateLimitStatPrefix(host, path, method, operationID string) string {
	return fmt.Sprintf("%s-%s-%s-%s", host, path, method, operationID)
}

func generateRoutePath(prefix, path string) string {
	if prefix == "" {
		return path
	}

	// Avoids path joins (removes // in e.g. /path//subpath, or //subpath)
	return fmt.Sprintf(`%s/%s`, strings.TrimSuffix(prefix, "/"), strings.TrimPrefix(path, "/"))
}

func generateRoute(
	clusterName string,
	corsPolicy *route.CorsPolicy,
	rewriteRegex *options.RewriteRegex,
	QoS *options.QoSOptions,
	websocket *bool,
) (*route.Route_Route, error) {

	var rewritePathRegex *envoytypematcher.RegexMatchAndSubstitute
	if rewriteRegex != nil {
		rewritePathRegex = types.GenerateRewriteRegex(rewriteRegex.Pattern, rewriteRegex.Substitution)
	}

	var (
		requestTimeout, requestIdleTimeout int64  = 0, 0
		retries                            uint32 = 0
	)
	if QoS != nil {
		retries = QoS.Retries
		requestTimeout = int64(QoS.RequestTimeout)
		requestIdleTimeout = int64(QoS.IdleTimeout)
	}

	routeRoute := &route.Route_Route{
		Route: &route.RouteAction{
			ClusterSpecifier: &route.RouteAction_Cluster{
				Cluster: clusterName,
			},
		},
	}

	if corsPolicy != nil {
		routeRoute.Route.Cors = corsPolicy
	}
	if rewritePathRegex != nil {
		routeRoute.Route.RegexRewrite = rewritePathRegex
	}

	if requestTimeout != 0 {
		routeRoute.Route.Timeout = &durationpb.Duration{Seconds: requestTimeout}
	}
	if requestIdleTimeout != 0 {
		routeRoute.Route.IdleTimeout = &durationpb.Duration{Seconds: requestIdleTimeout}
	}

	if retries != 0 {
		routeRoute.Route.RetryPolicy = &route.RetryPolicy{
			RetryOn:    "5xx",
			NumRetries: &wrapperspb.UInt32Value{Value: retries},
		}
	}
	if websocket != nil && *websocket {
		routeRoute.Route.UpgradeConfigs = append(routeRoute.Route.UpgradeConfigs, &route.RouteAction_UpgradeConfig{UpgradeType: "websocket"})
	}
	if err := routeRoute.Route.Validate(); err != nil {
		return nil, fmt.Errorf("incorrect Route Action: %w", err)
	}

	return routeRoute, nil
}

func mapRateLimitConf(rlOpt *options.RateLimitOptions, statPrefix string) *ratelimit.LocalRateLimit {
	var seconds int64
	switch rlOpt.Unit {
	case "second":
		seconds = 1
	case "minute":
		seconds = 60
	case "hour":
		seconds = 60 * 60
	}

	responseCode := rlOpt.ResponseCode
	if responseCode == 0 {
		// HTTP Status too many requests
		responseCode = 429
	}

	rl := &ratelimit.LocalRateLimit{
		StatPrefix: statPrefix,
		Status: &envoy_type_v3.HttpStatus{
			Code: envoy_type_v3.StatusCode(responseCode),
		},
		TokenBucket: &envoy_type_v3.TokenBucket{
			MaxTokens: rlOpt.RequestsPerUnit,
			TokensPerFill: &wrapperspb.UInt32Value{
				Value: rlOpt.RequestsPerUnit,
			},
			FillInterval: &durationpb.Duration{
				Seconds: seconds,
			},
		},
		FilterEnabled: &envoy_config_core_v3.RuntimeFractionalPercent{
			DefaultValue: &envoy_type_v3.FractionalPercent{
				Numerator:   100,
				Denominator: envoy_type_v3.FractionalPercent_HUNDRED,
			},
			RuntimeKey: "local_rate_limit_enabled",
		},
		FilterEnforced: &envoy_config_core_v3.RuntimeFractionalPercent{
			DefaultValue: &envoy_type_v3.FractionalPercent{
				Numerator:   100,
				Denominator: envoy_type_v3.FractionalPercent_HUNDRED,
			},
			RuntimeKey: "local_rate_limit_enforced",
		},
		Stage:                                 0,
		LocalRateLimitPerDownstreamConnection: rlOpt.PerConnection,
	}

	return rl
}
