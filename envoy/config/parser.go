package config

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/jinzhu/copier"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/kubeshop/kusk-gateway/options"
)

const httpPathSeparator string = "/"

// UpdateConfigFromOpts creates Snapshot from OpenAPI spec and x-kusk options
func (e *envoyConfiguration) UpdateConfigFromOpts(opts *options.Options, spec *openapi3.T) error {
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
			// routeConfig := &RouteConfiguration{
			// 	name:       routeName,
			// 	method:     method,
			// 	path:       routePath,
			// 	vhosts:     vhosts,
			// 	parameters: params,
			// }

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
			var routeRoute *route.Route_Route
			if &finalOpts.Redirect != nil {
				if routeRedirect, err = generateRedirect(&finalOpts.Redirect); err != nil {
					return err
				}
			} else {
				clusterName := generateClusterName(finalOpts.Service)
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

			}
			e.AddRoute(routeName, vhosts, routeMatcher, routeRoute, routeRedirect)
		}
	}

	return nil
}
