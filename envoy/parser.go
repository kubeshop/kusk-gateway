package envoy

import (
	"fmt"
	"strings"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoytypematcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/durationpb"

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
func (e *envoyConfiguration) UpdateConfigFromAPIOpts(opts *options.Options, spec *openapi3.T) error {
	var vhosts []string
	for _, vhost := range opts.Hosts {
		vhosts = append(vhosts, string(vhost))
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

			routeName := generateRouteName(routePath, method)
			params := extractParams(operation.Parameters)

			corsPolicy, err := generateCORSPolicy(finalOpts.CORS)
			if err != nil {
				return err
			}
			// routeMatcher defines how we match a route by the provided path and the headers
			routeMatcher := generateRouteMatch(routePath, method, params, corsPolicy)

			// This block creates redirect route
			// We either create the redirect or the route with proxy to upstream
			// Redirect takes a precedence.
			if finalOpts.Redirect != nil {
				routeRedirect, err := generateRedirect(finalOpts.Redirect)
				if err != nil {
					return err
				}

				rt := &route.Route{
					Name:   routeName,
					Match:  routeMatcher,
					Action: routeRedirect,
				}

				if err := e.AddRoute(vhosts, rt); err != nil {
					return fmt.Errorf("failure adding redirect route: %w", err)
				}

				return nil
			}

			//var routeRoute *route.Route_Route
			upstreamHostname, upstreamPort := getUpstreamHost(finalOpts.Upstream)
			clusterName := generateClusterName(upstreamHostname, upstreamPort)
			if !e.ClusterExist(clusterName) {
				e.AddCluster(clusterName, upstreamHostname, upstreamPort)
			}
			var rewritePathRegex *envoytypematcher.RegexMatchAndSubstitute
			if finalOpts.Path != nil {
				rewritePathRegex = GenerateRewriteRegex(finalOpts.Path.Rewrite.Pattern, finalOpts.Path.Rewrite.Substitution)
			}

			var (
				requestTimeout, requestIdleTimeout int64  = 0, 0
				retries                            uint32 = 0
			)
			if finalOpts.QoS != nil {
				retries = finalOpts.QoS.Retries
				requestTimeout = int64(finalOpts.QoS.RequestTimeout)
				requestIdleTimeout = int64(finalOpts.QoS.IdleTimeout)
			}

			routeRoute, err := generateRoute(
				clusterName,
				corsPolicy,
				rewritePathRegex,
				requestTimeout,
				requestIdleTimeout,
				retries)

			if err != nil {
				return err
			}

			rt := &route.Route{
				Name:   routeName,
				Match:  routeMatcher,
				Action: routeRoute,
			}
			if err := e.AddRoute(vhosts, rt); err != nil {
				return fmt.Errorf("failure adding route: %w", err)
			}
		}
	}

	return nil
}

// extract Params returns a map mapping the name of a paths parameter to its schema
// where the schema elements we care about are its type and enum if its defined
func extractParams(parameters openapi3.Parameters) map[string]ParamSchema {
	params := map[string]ParamSchema{}

	for _, parameter := range parameters {
		// Prevent populating map with empty parameter names
		if parameter.Value != nil && parameter.Value.Name != "" {
			params[parameter.Value.Name] = ParamSchema{}

			// Extract the schema if it's not nil and assign the map value
			if parameter.Value.Schema != nil && parameter.Value.Schema.Value != nil {
				schemaValue := parameter.Value.Schema.Value

				// It is acceptable for Type and / or Enum to have their zero value
				// It means the user has not defined it, and we will construct the regex path accordingly
				params[fmt.Sprintf("{%s}", parameter.Value.Name)] = ParamSchema{
					Type: schemaValue.Type,
					Enum: schemaValue.Enum,
				}
			}
		}
	}

	return params
}

// UpdateConfigFromOpts updates Envoy configuration from Options only
func (e *envoyConfiguration) UpdateConfigFromOpts(opts *options.StaticOptions) error {
	var vhosts []string
	for _, vhost := range opts.Hosts {
		vhosts = append(vhosts, string(vhost))
	}

	// Iterate on all paths and build routes
	for path, methods := range opts.Paths {
		for method, methodOpts := range methods {
			routePath := generateRoutePath("", path)
			routeName := generateRouteName(routePath, string(method))

			corsPolicy, err := generateCORSPolicy(methodOpts.CORS)
			if err != nil {
				return err
			}

			routeMatcher := generateRouteMatch(routePath, string(method), nil, corsPolicy)
			if methodOpts.Redirect != nil {
				// Generating Redirect
				routeRedirect, err := generateRedirect(methodOpts.Redirect)
				if err != nil {
					return err
				}

				rt := &route.Route{
					Name:   routeName,
					Match:  routeMatcher,
					Action: routeRedirect,
				}
				if err := e.AddRoute(vhosts, rt); err != nil {
					return fmt.Errorf("failure adding redirect route: %w", err)
				}
				continue
			}

			// Generating Route
			var routeRoute *route.Route_Route

			upstreamHostname, upstreamPort := getUpstreamHost(methodOpts.Upstream)
			clusterName := generateClusterName(upstreamHostname, upstreamPort)
			if !e.ClusterExist(clusterName) {
				e.AddCluster(clusterName, upstreamHostname, upstreamPort)
			}

			var rewritePathRegex *envoytypematcher.RegexMatchAndSubstitute
			if methodOpts.Path != nil {
				if methodOpts.Path.Rewrite.Pattern != "" {
					rewritePathRegex = GenerateRewriteRegex(methodOpts.Path.Rewrite.Pattern, methodOpts.Path.Rewrite.Substitution)
				}
			}

			var (
				requestTimeout, requestIdleTimeout int64  = 0, 0
				retries                            uint32 = 0
			)
			if methodOpts.QoS != nil {
				retries = methodOpts.QoS.Retries
				requestTimeout = int64(methodOpts.QoS.RequestTimeout)
				requestIdleTimeout = int64(methodOpts.QoS.IdleTimeout)

			}
			if routeRoute, err = generateRoute(
				clusterName,
				corsPolicy,
				rewritePathRegex,
				requestTimeout,
				requestIdleTimeout,
				retries); err != nil {
				return err
			}

			rt := &route.Route{
				Name:   routeName,
				Match:  routeMatcher,
				Action: routeRoute,
			}
			if err := e.AddRoute(vhosts, rt); err != nil {
				return fmt.Errorf("failure adding route: %w", err)
			}
		}
	}

	return nil
}

func generateRouteMatch(path string, method string, pathParameters map[string]ParamSchema, corsPolicy *route.CorsPolicy) *route.RouteMatch {
	headerMatcherConfig := []*route.HeaderMatcher{
		GetHeaderMatcherConfig([]string{strings.ToUpper(method)}, corsPolicy != nil),
	}

	routeMatcherBuilder := NewRouteMatcherBuilder(path, pathParameters)
	return routeMatcherBuilder.GetRouteMatcher(headerMatcherConfig)
}

func generateRedirect(redirectOpts *options.RedirectOptions) (*route.Route_Redirect, error) {
	if redirectOpts == nil {
		return nil, nil
	}

	redirect, err := NewRouteRedirectBuilder().
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

	return GenerateCORSPolicy(
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

func generateRoute(
	clusterName string,
	corsPolicy *route.CorsPolicy,
	rewritePathRegex *envoytypematcher.RegexMatchAndSubstitute,
	timeout int64,
	idleTimeout int64,
	retries uint32) (*route.Route_Route, error) {

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

	if timeout != 0 {
		routeRoute.Route.Timeout = &durationpb.Duration{Seconds: timeout}
	}
	if idleTimeout != 0 {
		routeRoute.Route.IdleTimeout = &durationpb.Duration{Seconds: idleTimeout}
	}

	if retries != 0 {
		routeRoute.Route.RetryPolicy = &route.RetryPolicy{
			RetryOn:    "5xx",
			NumRetries: &wrappers.UInt32Value{Value: retries},
		}
	}
	if err := routeRoute.Route.Validate(); err != nil {
		return nil, fmt.Errorf("incorrect Route Action: %w", err)
	}
	return routeRoute, nil
}
