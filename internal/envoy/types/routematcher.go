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
package types

import (
	"fmt"
	"regexp"
	"strings"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoytypematcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
)

var (
	// compiles e.g. /pets/{petID}/legs/{leg1}
	rePathParams = regexp.MustCompile(`{[A-z0-9]+}`)

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

type RouteMatcherBuilder struct {
	path           string
	pathParameters map[string]ParamSchema
}

func NewRouteMatcherBuilder(path string, pathParameters map[string]ParamSchema) *RouteMatcherBuilder {
	return &RouteMatcherBuilder{
		path:           path,
		pathParameters: pathParameters,
	}
}

func (r RouteMatcherBuilder) GetRouteMatcher(headers []*route.HeaderMatcher) *route.RouteMatch {
	if rePathParams.MatchString(r.path) {
		routePath := r.convertParamsToRoutePath()

		return &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_SafeRegex{
				SafeRegex: &envoytypematcher.RegexMatcher{
					EngineType: &envoytypematcher.RegexMatcher_GoogleRe2{
						GoogleRe2: &envoytypematcher.RegexMatcher_GoogleRE2{},
					},
					Regex: routePath,
				},
			},
			Headers: headers,
		}
	}

	if usePrefix := strings.HasSuffix(r.path, "/"); usePrefix {
		return &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: r.path,
			},
			Headers: headers,
		}
	}

	return &route.RouteMatch{
		PathSpecifier: &route.RouteMatch_Path{
			Path: r.path,
		},
		Headers: headers,
	}
}

func (r RouteMatcherBuilder) convertParamsToRoutePath() string {
	routePath := r.path
	if r.pathParameters == nil {
		return routePath
	}

	for _, match := range rePathParams.FindAllString(routePath, -1) {
		param := r.pathParameters[match]

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

	return routePath
}

func convertToStringSlice(in []interface{}) []string {
	s := make([]string, len(in))
	for i := range in {
		s[i] = fmt.Sprint(in[i])
	}

	return s
}
