package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoytypematcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/wrapperspb"

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
		pathSubOpts, _ := opts.PathSubOptions[path]
		if err := copier.CopyWithOption(&pathOpts, &pathSubOpts, copier.Option{IgnoreEmpty: true, DeepCopy: false}); err != nil {
			return nil, err
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
			// finally override with opSubOpts
			if err := copier.CopyWithOption(&finalOpts, &opSubOpts, copier.Option{IgnoreEmpty: true, DeepCopy: false}); err != nil {
				return nil, err
			}
			// Once we have final merged Options, skip if disabled either on top, path or method level.
			if *finalOpts.Disabled {
				continue
			}
			clusterName := generateClusterName(finalOpts.Service)
			if !e.ClusterExist(clusterName) {
				e.AddCluster(clusterName, finalOpts.Service.Name, finalOpts.Service.Port)
			}
			routePath := generateRoutePath(finalOpts.Path.Base, path)
			routeName := generateRouteName(routePath, method)
			trimPrefixRegex := generateTrimPrefixRegex(finalOpts.Path.TrimPrefix)
			corsPolicy, err := generateCORSPolicy(&finalOpts.CORS)
			if err != nil {
				return nil, err
			}
			e.AddRoute(
				routeName,
				routePath,
				method,
				clusterName,
				trimPrefixRegex,
				corsPolicy,
				int64(finalOpts.Timeouts.RequestTimeout),
				int64(finalOpts.Timeouts.IdleTimeout),
				finalOpts.Path.Retries,
			)
		}
	}
	return e.GenerateSnapshot()
}

// each cluster can be uniquely identified by dns name + port (i.e. canonical Host, which is hostname:port)
func generateClusterName(service options.ServiceOptions) string {
	return fmt.Sprintf("%s-%d", service.Name, service.Port)
}

func generateCORSPolicy(corsOpts *options.CORSOptions) (*route.CorsPolicy, error) {
	if reflect.DeepEqual(&options.CORSOptions{}, corsOpts) {
		return nil, nil
	}
	allowOriginsMatcher := []*envoytypematcher.StringMatcher{}
	for _, origin := range corsOpts.Origins {
		entry := &envoytypematcher.StringMatcher{
			// TODO: We support only exact strings, no regexp - fix this if applicable
			MatchPattern: &envoytypematcher.StringMatcher_Exact{
				Exact: origin,
			},
			IgnoreCase: false,
		}
		allowOriginsMatcher = append(allowOriginsMatcher, entry)
	}
	corsPolicy := &route.CorsPolicy{
		AllowOriginStringMatch: allowOriginsMatcher,
		AllowMethods:           strings.Join(corsOpts.Methods, ","),
		AllowHeaders:           strings.Join(corsOpts.Headers, ","),
		ExposeHeaders:          strings.Join(corsOpts.ExposeHeaders, ","),
		MaxAge:                 strconv.Itoa(corsOpts.MaxAge),
	}
	if corsOpts.Credentials != nil {
		corsPolicy.AllowCredentials = &wrapperspb.BoolValue{
			Value: *corsOpts.Credentials,
		}
	}
	if err := corsPolicy.Validate(); err != nil {
		return nil, err
	}
	return corsPolicy, nil
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

func generateTrimPrefixRegex(trimPath string) string {
	if trimPath == "" {
		return ""
	}
	// for e.g. '/path/' or '/path' or 'path' or 'path/' returns '^/path/'
	sanitisedTrimPath := strings.TrimPrefix(strings.TrimSuffix(trimPath, httpPathSeparator), httpPathSeparator)
	return fmt.Sprintf("^%s%s%s", httpPathSeparator, sanitisedTrimPath, httpPathSeparator)
}
