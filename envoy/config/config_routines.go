// package config provides structures to create and update routing configuration for Envoy Fleet
// it is not used for Fleet creation, only for configuration snapshot creation.

package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoytypematcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/getkin/kin-openapi/openapi3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/kubeshop/kusk-gateway/options"
)

var (
	// compiles e.g. /pets/{petID}/legs/{leg1}
	rePathParams = regexp.MustCompile(`{[A-z0-9]+}`)

	redirectResponseCodeName = map[uint32]string{
		301: "MOVED_PERMANENTLY",
		302: "FOUND",
		303: "SEE_OTHER",
		307: "TEMPORARY_REDIRECT",
		308: "PERMANENT_REDIRECT",
	}

	// regexes for path that has OpenAPI parameters names ({petID})
	// OpenAPI supports:
	// * strings
	// * number (double and float)
	// * integer (int32, int64)
	// we don't use OpenAPI "format" or "pattern" right now
	parameterTypeRegexReplacements = map[string]string{
		"string":  `([.a-zA-Z0-9-]+)`,
		"integer": `([0-9]+)`,
		"number":  `([0-9]*[.])?[0-9]+`,
	}
)

type ParamSchema struct {
	Type string
	Enum []interface{}
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

func generateRouteMatch(path string, method string, pathParameters map[string]ParamSchema, corsPolicy *route.CorsPolicy) *route.RouteMatch {
	// headerMatcher allows as to match route by the method (:method header) or any other header
	var headerMatcher []*route.HeaderMatcher
	method = strings.ToUpper(method)
	if corsPolicy != nil {
		// If CORS specified, we add OPTIONS method to match the route
		headerMatcher = append(headerMatcher, generateMethodHeaderMatcher([]string{method, "OPTIONS"}))
	} else {
		headerMatcher = append(headerMatcher, generateMethodHeaderMatcher([]string{method}))
	}

	var routeMatcher *route.RouteMatch
	// Create Path matcher - either regex if there are parameters, prefix or simple path match
	switch {
	// if has regex - regex matcher
	case rePathParams.MatchString(path):
		// if regex in the path - matcher is using RouteMatch_Regex with /{pattern} replaced by related regex
		routePath := path
		for _, match := range rePathParams.FindAllString(routePath, -1) {
			param := pathParameters[match]

			// default replacement regex
			replacementRegex := ""
			// if type = enum, construct enum regex capture grouup
			if len(param.Enum) > 0 {
				enumStrings := convertToStringSlice(param.Enum)
				replacementRegex = fmt.Sprintf("(%s)", strings.Join(enumStrings, "|"))
			} else if regex, ok := parameterTypeRegexReplacements[param.Type]; ok {
				replacementRegex = regex
			} else {
				// If param type didn't match, we use string, basically - anything valid for URL path
				replacementRegex = parameterTypeRegexReplacements["string"]
			}
			routePath = strings.ReplaceAll(routePath, match, replacementRegex)
		}
		routeMatcher = &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_SafeRegex{
				SafeRegex: &envoytypematcher.RegexMatcher{
					EngineType: &envoytypematcher.RegexMatcher_GoogleRe2{
						GoogleRe2: &envoytypematcher.RegexMatcher_GoogleRE2{},
					},
					Regex: routePath,
				},
			},
			Headers: headerMatcher,
		}
	case strings.HasSuffix(path, "/"):
		// if path ends in / - path prefix match
		routeMatcher = &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: path,
			},
			Headers: headerMatcher,
		}
	default:
		// default - exact path matching
		routeMatcher = &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Path{
				Path: path,
			},
			Headers: headerMatcher,
		}
	}
	return routeMatcher
}

// Matching on method means matching on its header.
// Additional headers can be ammended.
func generateMethodHeaderMatcher(methods []string) *route.HeaderMatcher {
	switch len(methods) {
	case 0:
		return nil
	case 1:
		// creates exact match for method
		return &route.HeaderMatcher{
			Name: ":method",
			HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
				StringMatch: &envoytypematcher.StringMatcher{
					MatchPattern: &envoytypematcher.StringMatcher_Exact{Exact: methods[0]},
				},
			},
		}
	default:
		// creates regex "^OPTIONS$|^GET$"
		for i := range methods {
			methods[i] = fmt.Sprintf("^%s$", methods[i])
		}
		regex := strings.Join(methods, "|")
		return &route.HeaderMatcher{
			Name: ":method",
			HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
				StringMatch: &envoytypematcher.StringMatcher{
					MatchPattern: &envoytypematcher.StringMatcher_SafeRegex{
						SafeRegex: &envoytypematcher.RegexMatcher{
							EngineType: &envoytypematcher.RegexMatcher_GoogleRe2{
								GoogleRe2: &envoytypematcher.RegexMatcher_GoogleRE2{},
							},
							Regex: regex,
						},
					},
				},
			},
		}
	}
}

func convertToStringSlice(in []interface{}) []string {
	s := make([]string, len(in))
	for i := range in {
		s[i] = fmt.Sprint(in[i])
	}

	return s
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
			NumRetries: &wrapperspb.UInt32Value{Value: retries},
		}
	}
	if err := routeRoute.Route.Validate(); err != nil {
		return nil, fmt.Errorf("incorrect Route Action: %w", err)
	}
	return routeRoute, nil
}

func generateRedirect(redirectOpts *options.RedirectOptions) (*route.Route_Redirect, error) {
	if redirectOpts == nil {
		return nil, nil
	}
	redirectAction := &route.RedirectAction{
		HostRedirect: redirectOpts.HostRedirect,
		PortRedirect: redirectOpts.PortRedirect,
	}
	if redirectOpts.SchemeRedirect != "" {
		redirectAction.SchemeRewriteSpecifier = &route.RedirectAction_SchemeRedirect{SchemeRedirect: redirectOpts.SchemeRedirect}
	}
	// PathRedirect and RewriteRegex are mutually exlusive
	// Path rewrite with regex
	if redirectOpts.RewriteRegex != nil && redirectOpts.RewriteRegex.Pattern != "" {
		redirectAction.PathRewriteSpecifier = &route.RedirectAction_RegexRewrite{
			RegexRewrite: generateRewriteRegex(redirectOpts.RewriteRegex.Pattern, redirectOpts.RewriteRegex.Substitution),
		}
	}
	// Or path rewrite with path redirect
	if redirectOpts.PathRedirect != "" {
		redirectAction.PathRewriteSpecifier = &route.RedirectAction_PathRedirect{
			PathRedirect: redirectOpts.PathRedirect,
		}
	}
	// if the code is undefined, it is set to 301 by default in Envoy
	// otherwise we need to convert HTTP code (e.g. 308) to internal go-control-plane enumeration
	if redirectOpts.ResponseCode != 0 {
		// go-control-plane uses internal map of response code names that we need to translate from real HTTP code
		redirectActionResponseCodeName, ok := redirectResponseCodeName[redirectOpts.ResponseCode]
		if !ok {
			return nil, fmt.Errorf("missing redirect code name for HTTP code %d", redirectOpts.ResponseCode)
		}
		code := route.RedirectAction_RedirectResponseCode_value[redirectActionResponseCodeName]
		redirectAction.ResponseCode = route.RedirectAction_RedirectResponseCode(code)
	}
	if redirectOpts.StripQuery != nil {
		redirectAction.StripQuery = *redirectOpts.StripQuery
	}
	if err := redirectAction.Validate(); err != nil {
		return nil, fmt.Errorf("incorrect Redirect Action: %w", err)
	}
	return &route.Route_Redirect{Redirect: redirectAction}, nil
}

func generateCORSPolicy(corsOpts *options.CORSOptions) (*route.CorsPolicy, error) {
	if corsOpts == nil {
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
		return nil, fmt.Errorf("incorrect CORS configuration: %w", err)
	}
	return corsPolicy, nil
}

func getUpstreamHost(upstreamOpts *options.UpstreamOptions) (hostname string, port uint32) {
	if upstreamOpts.Service != nil {
		return fmt.Sprintf("%s.%s.svc.cluster.local.", upstreamOpts.Service.Name, upstreamOpts.Service.Namespace), upstreamOpts.Service.Port
	}
	return upstreamOpts.Host.Hostname, upstreamOpts.Host.Port
}
