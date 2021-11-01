package config

import (
	"fmt"
	"strings"

	envoytypematcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/jinzhu/copier"

	"github.com/kubeshop/kusk-gateway/options"
)

const httpPathSeparator string = "/"

// GenerateConfigSnapshotFromOpts creates Snapshot from OpenAPI spec and x-kusk options
func (e *envoyConfiguration) GenerateConfigSnapshotFromOpts(opts *options.Options, spec *openapi3.T) (*cache.Snapshot, error) {
	e.vhosts = opts.Hosts
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
			return nil, err
		}

		if pathSubOpts, ok := opts.PathSubOptions[path]; ok {
			if err := copier.CopyWithOption(&pathOpts, &pathSubOpts, copier.Option{IgnoreEmpty: true, DeepCopy: false}); err != nil {
				return nil, err
			}
		}

		// x-kusk options per operation (http method)
		for method := range pathItem.Operations() {
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
				return nil, err
			}

			// finally override with opSubOpts, if there are any
			if ok {
				if err := copier.CopyWithOption(&finalOpts, &opSubOpts, copier.Option{IgnoreEmpty: true, DeepCopy: false}); err != nil {
					return nil, err
				}
			}

			// Once we have final merged Options, skip if disabled either on top, path or method level.
			if finalOpts.Disabled != nil && *finalOpts.Disabled {
				continue
			}

			routePath := generateRoutePath(finalOpts.Path.Base, path)
			routeName := generateRouteName(routePath, method)

			routeConfig := &RouteConfiguration{
				name:   routeName,
				method: method,
				path:   routePath,
			}
			// This block creates redirect route
			redirectAction, err := generateRedirectAction(&finalOpts.Redirect)
			if err != nil {
				return nil, err
			}
			// We either create the redirect or the route with proxy to backend
			if redirectAction != nil {
				routeConfig.redirectAction = redirectAction
				e.AddRouteRedirect(routeConfig)
			} else {
				// This block create usual route with backend service
				routeConfig.clusterName = generateClusterName(finalOpts.Service)
				if !e.ClusterExist(routeConfig.clusterName) {
					e.AddCluster(routeConfig.clusterName, finalOpts.Service.Name, finalOpts.Service.Port)
				}

				routeConfig.rewritePathRegex = generateRewriteRegex(finalOpts.Path.Rewrite.Pattern, finalOpts.Path.Rewrite.Substitution)
				routeConfig.idleTimeout = int64(finalOpts.Timeouts.IdleTimeout)
				routeConfig.timeout = int64(finalOpts.Timeouts.RequestTimeout)
				routeConfig.retries = finalOpts.Path.Retries
				routeConfig.corsPolicy, err = generateCORSPolicy(&finalOpts.CORS)
				if err != nil {
					return nil, err
				}

				e.AddRoute(routeConfig)
			}
		}
	}

	return e.GenerateSnapshot()
}

// each cluster can be uniquely identified by dns name + port (i.e. canonical Host, which is hostname:port)
func generateClusterName(service options.ServiceOptions) string {
	return fmt.Sprintf("%s-%d", service.Name, service.Port)
}

// Can be moved to operationID, but generally we just need unique string
func generateRouteName(path string, method string) string { return fmt.Sprintf("%s-%s", path, method) }

func generateRoutePath(base string, path string) string {
	if base == "" {
		return path
	}
	// Avoids path joins (removes // in e.g. /path//subpath, or //subpath)
	return fmt.Sprintf(`%s/%s`, strings.TrimSuffix(base, httpPathSeparator), strings.TrimPrefix(path, httpPathSeparator))
}

func generateRewriteRegex(pattern string, substitution string) *envoytypematcher.RegexMatchAndSubstitute {
	if pattern == "" {
		return nil
	}
	return &envoytypematcher.RegexMatchAndSubstitute{
		Pattern: &envoytypematcher.RegexMatcher{
			EngineType: &envoytypematcher.RegexMatcher_GoogleRe2{
				GoogleRe2: &envoytypematcher.RegexMatcher_GoogleRE2{},
			},
			Regex: pattern,
		},
		Substitution: substitution,
	}
}
