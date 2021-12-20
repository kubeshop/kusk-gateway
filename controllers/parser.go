package controllers

import (
	"fmt"
	"strings"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoytypematcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/getkin/kin-openapi/openapi3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/kubeshop/kusk-gateway/envoy/config"
	"github.com/kubeshop/kusk-gateway/envoy/types"
	"github.com/kubeshop/kusk-gateway/options"
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
func UpdateConfigFromAPIOpts(envoyConfiguration *config.EnvoyConfiguration, opts *options.Options, spec *openapi3.T) error {
	for _, vhost := range opts.Hosts {
		vh := types.NewVirtualHost(string(vhost))
		envoyConfiguration.AddVirtualHost(vh)
	}

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

			routePath := path
			if finalOpts.Path != nil {
				routePath = generateRoutePath(finalOpts.Path.Prefix, path)
			}

			corsPolicy, err := generateCORSPolicy(finalOpts.CORS)
			if err != nil {
				return err
			}

			rt := &route.Route{
				Name: generateRouteName(routePath, method),
				Match: generateRouteMatch(
					routePath,
					method,
					extractParams(operation.Parameters),
					corsPolicy,
				),
			}

			// This block creates redirect route
			// We either create the redirect or the route with proxy to upstream
			// Redirect takes a precedence.
			if finalOpts.Redirect != nil {
				routeRedirect, err := generateRedirect(finalOpts.Redirect)
				if err != nil {
					return err
				}

				rt.Action = routeRedirect
			} else {
				upstreamHostname, upstreamPort := getUpstreamHost(finalOpts.Upstream)
				clusterName := generateClusterName(upstreamHostname, upstreamPort)
				if !envoyConfiguration.ClusterExist(clusterName) {
					envoyConfiguration.AddCluster(clusterName, upstreamHostname, upstreamPort)
				}

				routeRoute, err := generateRoute(
					clusterName,
					corsPolicy,
					&generateRouteOpts{
						RewriteRegex: &opts.Path.Rewrite,
						QoS:          opts.QoS,
					},
				)
				if err != nil {
					return err
				}

				rt.Action = routeRoute
			}

			for _, vh := range envoyConfiguration.GetVirtualHosts() {
				if err := envoyConfiguration.AddRouteToVHost(vh.Name, rt); err != nil {
					return fmt.Errorf("failure adding redirect route: %w", err)
				}
			}
		}
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
	for _, vhost := range opts.Hosts {
		vh := types.NewVirtualHost(string(vhost))
		envoyConfiguration.AddVirtualHost(vh)
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
				Name: generateRouteName(routePath, strMethod),
				Match: generateRouteMatch(
					routePath,
					string(method),
					nil,
					corsPolicy,
				),
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

				var rewriteOpts options.RewriteRegex
				if methodOpts.Path != nil {
					rewriteOpts = methodOpts.Path.Rewrite
				}
				routeRoute, err := generateRoute(
					clusterName,
					corsPolicy,
					&generateRouteOpts{
						RewriteRegex: &rewriteOpts,
						QoS:          methodOpts.QoS,
					},
				)
				if err != nil {
					return err
				}

				rt.Action = routeRoute
			}

			for _, vh := range envoyConfiguration.GetVirtualHosts() {
				if err := envoyConfiguration.AddRouteToVHost(vh.Name, rt); err != nil {
					return fmt.Errorf("failure adding redirect route: %w", err)
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

	redirect, err := types.NewRouteRedirectBuilder().
		HostRedirect(redirectOpts.HostRedirect).
		PortRedirect(redirectOpts.PortRedirect).
		SchemeRedirect(redirectOpts.SchemeRedirect).
		RegexRedirect(redirectOpts.RewriteRegex.Pattern, redirectOpts.RewriteRegex.Substitution).
		PathRedirect(redirectOpts.PathRedirect).
		ResponseCode(redirectOpts.ResponseCode).
		StripQuery(redirectOpts.StripQuery).
		ValidateAndReturn()

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

// Can be moved to operationID, but generally we just need unique string
func generateRouteName(path string, method string) string {
	return fmt.Sprintf("%s-%s", path, strings.ToUpper(method))
}

func generateRoutePath(prefix, path string) string {
	if prefix == "" {
		return path
	}

	// Avoids path joins (removes // in e.g. /path//subpath, or //subpath)
	return fmt.Sprintf(`%s/%s`, strings.TrimSuffix(prefix, "/"), strings.TrimPrefix(path, "/"))
}

type generateRouteOpts struct {
	RewriteRegex *options.RewriteRegex
	QoS          *options.QoSOptions
}

func generateRoute(
	clusterName string,
	corsPolicy *route.CorsPolicy,
	opts *generateRouteOpts,
) (*route.Route_Route, error) {

	var rewritePathRegex *envoytypematcher.RegexMatchAndSubstitute
	if opts.RewriteRegex != nil {
		rewritePathRegex = types.GenerateRewriteRegex(opts.RewriteRegex.Pattern, opts.RewriteRegex.Substitution)
	}

	var (
		requestTimeout, requestIdleTimeout int64  = 0, 0
		retries                            uint32 = 0
	)
	if opts.QoS != nil {
		retries = opts.QoS.Retries
		requestTimeout = int64(opts.QoS.RequestTimeout)
		requestIdleTimeout = int64(opts.QoS.IdleTimeout)
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

	if err := routeRoute.Route.Validate(); err != nil {
		return nil, fmt.Errorf("incorrect Route Action: %w", err)
	}

	return routeRoute, nil
}
