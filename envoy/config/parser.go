package config

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoytypematcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

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
	vhosts := []string{}
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

			var routePath string
			if finalOpts.Path != nil {
				routePath = generateRoutePath(finalOpts.Path.Prefix, path)
			} else {
				routePath = generateRoutePath("", path)
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
				if err := e.AddRoute(routeName, vhosts, routeMatcher, nil, routeRedirect); err != nil {
					return fmt.Errorf("failure adding redirect route: %w", err)
				}
			} else {
				var routeRoute *route.Route_Route
				upstreamHostname, upstreamPort := getUpstreamHost(finalOpts.Upstream)
				clusterName := generateClusterName(upstreamHostname, upstreamPort)
				if !e.ClusterExist(clusterName) {
					e.AddCluster(clusterName, upstreamHostname, upstreamPort)
				}
				var rewritePathRegex *envoytypematcher.RegexMatchAndSubstitute
				if finalOpts.Path != nil {
					rewritePathRegex = generateRewriteRegex(finalOpts.Path.Rewrite.Pattern, finalOpts.Path.Rewrite.Substitution)
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
				if routeRoute, err = generateRoute(
					clusterName,
					corsPolicy,
					rewritePathRegex,
					requestTimeout,
					requestIdleTimeout,
					retries); err != nil {
					return err
				}
				if err := e.AddRoute(routeName, vhosts, routeMatcher, routeRoute, nil); err != nil {
					return fmt.Errorf("failure adding route: %w", err)
				}
			}
		}
	}

	return nil
}

// UpdateConfigFromOpts updates Envoy configuration from Options only
func (e *envoyConfiguration) UpdateConfigFromOpts(opts *options.StaticOptions) error {
	vhosts := []string{}
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
				if err := e.AddRoute(routeName, vhosts, routeMatcher, nil, routeRedirect); err != nil {
					return fmt.Errorf("failure adding redirect route: %w", err)
				}
				continue
			} else {
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
						rewritePathRegex = generateRewriteRegex(methodOpts.Path.Rewrite.Pattern, methodOpts.Path.Rewrite.Substitution)
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
				if err := e.AddRoute(routeName, vhosts, routeMatcher, routeRoute, nil); err != nil {
					return fmt.Errorf("failure adding route: %w", err)
				}

			}
		}
	}

	return nil
}
