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
package mocking

import (
	"fmt"
	"regexp"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoytypematcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	"github.com/kubeshop/kusk-gateway/internal/envoy/types"
)

var (
	jsonPatternStr       = "^application/.*json$"
	jsonMediaTypePattern = regexp.MustCompile(jsonPatternStr)

	xmlPatternStr       = "^application/.*xml$"
	xmlMediaTypePattern = regexp.MustCompile(xmlPatternStr)

	textPatternStr       = "^text/.*$"
	textMediaTypePattern = regexp.MustCompile(textPatternStr)
)

type BuildMockedRouteArgs struct {
	RoutePath           string
	Method              string
	StatusCode          uint32
	ExampleContent      interface{}
	RequireAcceptHeader bool
}

type MockedRouteBuilder interface {
	BuildMockedRoute(args *BuildMockedRouteArgs) (*route.Route, error)
}

// NewRouteBuilder returns a new route builder for building routes that are mocked
// based on the provided mediaType
// Supported mediaTypes are:
// - application/json
// - application/xml
// - text/plain
// if the mediaType is not supported, an error is returned
func NewRouteBuilder(mediaType string) (MockedRouteBuilder, error) {
	baseMockedRouteBuilder := baseMockedRouteBuilder{}

	switch {
	case jsonMediaTypePattern.MatchString(mediaType):
		return mockedJsonRouteBuilder{
			baseMockedRouteBuilder,
		}, nil
	case xmlMediaTypePattern.MatchString(mediaType):
		return mockedXmlRouteBuilder{
			baseMockedRouteBuilder,
		}, nil
	case textMediaTypePattern.MatchString(mediaType):
		return mockedTextRouteBuilder{
			baseMockedRouteBuilder,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported media type: %s", mediaType)
	}
}

type baseMockedRouteBuilder struct{}

func (b baseMockedRouteBuilder) getBaseRoute(routePath, method string) *route.Route {
	return &route.Route{
		Name: types.GenerateRouteName(routePath, method),
		Match: &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_SafeRegex{
				SafeRegex: &envoytypematcher.RegexMatcher{
					EngineType: &envoytypematcher.RegexMatcher_GoogleRe2{
						GoogleRe2: &envoytypematcher.RegexMatcher_GoogleRE2{},
					},
					Regex: routePath,
				},
			},
			Headers: []*route.HeaderMatcher{
				types.GetHeaderMatcherConfig([]string{method}, false),
			},
		},
	}
}

func (b baseMockedRouteBuilder) getAcceptHeaderMatcher(regex string) *route.HeaderMatcher {
	return &route.HeaderMatcher{
		Name: "Accept",
		HeaderMatchSpecifier: &route.HeaderMatcher_SafeRegexMatch{
			SafeRegexMatch: &envoytypematcher.RegexMatcher{
				Regex: regex,
				EngineType: &envoytypematcher.RegexMatcher_GoogleRe2{
					GoogleRe2: &envoytypematcher.RegexMatcher_GoogleRE2{},
				},
			},
		},
	}
}

func (b baseMockedRouteBuilder) getAction(statusCode uint32, content string) *route.Route_DirectResponse {
	return &route.Route_DirectResponse{
		DirectResponse: &route.DirectResponseAction{
			Status: statusCode,
			Body: &envoy_config_core_v3.DataSource{
				Specifier: &envoy_config_core_v3.DataSource_InlineString{
					InlineString: content,
				},
			},
		},
	}
}
