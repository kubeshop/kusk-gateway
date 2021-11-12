package config

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/jinzhu/copier"

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

const httpPathSeparator string = "/"

// UpdateConfigFromAPIOpts updates Envoy configuration from OpenAPI spec and x-kusk options
func (e *envoyConfiguration) UpdateConfigFromAPIOpts(opts *options.Options, spec *openapi3.T) error {
	vhosts := []string{}
	for _, vhost := range opts.Hosts {
		vhosts = append(vhosts, string(vhost))
	}
	// Iterate on all paths and build routes
	// The overriding works in the following way:
	// 1. For each path create a copy of top x-kusk SubOpts struct as new pathOpts var. For that path override it with pathSubOpts.
	// 2. For each method create a copy of previously created pathOpts (finalOpts) and override it with opSubOpts.
	// Copier will skip override of undefined (nul) fields with IgnoreEmpty option.
	for path, pathItem := range spec.Paths {
		// x-kusk options per path
		// This var is reused during following methods merges,
		// we do this merge once per path since it is expensive to do it for every method
		var pathOpts options.SubOptions
		if err := copier.CopyWithOption(&pathOpts, &opts.SubOptions, copier.Option{IgnoreEmpty: true, DeepCopy: false}); err != nil {
			return err
		}

		if pathSubOpts, ok := opts.PathSubOptions[path]; ok {
			if err := copier.CopyWithOption(&pathOpts, &pathSubOpts, copier.Option{IgnoreEmpty: true, DeepCopy: false}); err != nil {
				return err
			}
		}

		// x-kusk options per operation (http method)
		for method, operation := range pathItem.Operations() {
			opSubOpts, ok := opts.OperationSubOptions[method+path]
			if ok {
				// Exit early if disabled per method, don't do further copies
				if *opSubOpts.Disabled {
					continue
				}
			}

			var finalOpts options.SubOptions

			// copy into new var already merged path opts
			if err := copier.CopyWithOption(&finalOpts, &pathOpts, copier.Option{IgnoreEmpty: true, DeepCopy: false}); err != nil {
				return err
			}

			// finally override with opSubOpts, if there are any
			if ok {
				if err := copier.CopyWithOption(&finalOpts, &opSubOpts, copier.Option{IgnoreEmpty: true, DeepCopy: false}); err != nil {
					return err
				}
			}

			// Once we have final merged Options, skip if disabled either on top, path or method level.
			if finalOpts.Disabled != nil && *finalOpts.Disabled {
				continue
			}

			routePath := generateRoutePath(finalOpts.Path.Base, path)
			routeName := generateRouteName(routePath, method)

			params := extractParams(operation.Parameters)

			corsPolicy, err := generateCORSPolicy(&finalOpts.CORS)
			if err != nil {
				return err
			}
			// routeMatcher defines how we match a route by the provided path and the headers
			routeMatcher := generateRouteMatch(routePath, method, params, corsPolicy)

			// This block creates redirect route
			// We either create the redirect or the route with proxy to backend
			// Redirect takes a precedence.
			var routeRedirect *route.Route_Redirect
			if routeRedirect, err = generateRedirect(&finalOpts.Redirect); err != nil {
				return err
			}
			if routeRedirect != nil {
				if err := e.AddRoute(routeName, vhosts, routeMatcher, nil, routeRedirect); err != nil {
					return fmt.Errorf("failure adding redirect route: %w", err)
				}
			} else {
				var routeRoute *route.Route_Route
				clusterName := generateClusterName(finalOpts.Service.Name, finalOpts.Service.Port)
				if !e.ClusterExist(clusterName) {
					e.AddCluster(clusterName, finalOpts.Service.Name, finalOpts.Service.Port)
				}
				rewritePathRegex := generateRewriteRegex(finalOpts.Path.Rewrite.Pattern, finalOpts.Path.Rewrite.Substitution)
				if routeRoute, err = generateRoute(
					clusterName,
					corsPolicy,
					rewritePathRegex,
					int64(finalOpts.Timeouts.RequestTimeout),
					int64(finalOpts.Timeouts.IdleTimeout),
					finalOpts.Path.Retries); err != nil {
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

				backend := methodOpts.Backend

				clusterName := generateClusterName(backend.Hostname, backend.Port)
				if !e.ClusterExist(clusterName) {
					e.AddCluster(clusterName, backend.Hostname, backend.Port)
				}

				var rewritePathRegex *envoytypematcher.RegexMatchAndSubstitute
				if backend.Rewrite != nil {
					rewritePathRegex = generateRewriteRegex(backend.Rewrite.Pattern, backend.Rewrite.Substitution)
				}

				var requestTimeout, requestIdleTimeout int64 = 0, 0
				if methodOpts.Timeouts != nil {
					requestTimeout = int64(methodOpts.Timeouts.RequestTimeout)
					requestIdleTimeout = int64(methodOpts.Timeouts.IdleTimeout)

				}
				if routeRoute, err = generateRoute(
					clusterName,
					corsPolicy,
					rewritePathRegex,
					requestTimeout,
					requestIdleTimeout,
					backend.Retries); err != nil {
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
