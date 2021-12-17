package types

import (
	"fmt"
	"strings"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
)

const headerMatcherName = ":method"

func GetHeaderMatcherConfig(methods []string, cors bool) *route.HeaderMatcher {
	if len(methods) == 0 {
		return nil
	}

	if len(methods) == 1 && !cors {
		return &route.HeaderMatcher{
			Name: headerMatcherName,
			HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
				StringMatch: &matcher.StringMatcher{
					MatchPattern: &matcher.StringMatcher_Exact{Exact: methods[0]},
				},
			},
		}
	}

	// creates regex "^OPTIONS$|^GET$"
	for i := range methods {
		methods[i] = fmt.Sprintf("^%s$", methods[i])
	}

	if cors {
		methods = append(methods, "^OPTIONS$")
	}

	return &route.HeaderMatcher{
		Name: headerMatcherName,
		HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
			StringMatch: &matcher.StringMatcher{
				MatchPattern: &matcher.StringMatcher_SafeRegex{
					SafeRegex: &matcher.RegexMatcher{
						EngineType: &matcher.RegexMatcher_GoogleRe2{
							GoogleRe2: &matcher.RegexMatcher_GoogleRE2{},
						},
						Regex: strings.Join(methods, "|"),
					},
				},
			},
		},
	}

}
