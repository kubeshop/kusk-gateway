package envoy

import (
	"fmt"
	"strconv"
	"strings"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoytypematcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func GenerateCORSPolicy(
	origins,
	methods,
	headers,
	exposeHeaders []string,
	maxAge int,
	credentials *bool,
) (*route.CorsPolicy, error) {
	corsPolicy := &route.CorsPolicy{
		AllowOriginStringMatch: getAllowOriginsMatchers(origins),
		AllowMethods:           strings.Join(methods, ","),
		AllowHeaders:           strings.Join(headers, ","),
		ExposeHeaders:          strings.Join(exposeHeaders, ","),
		MaxAge:                 strconv.Itoa(maxAge),
	}
	if credentials != nil {
		corsPolicy.AllowCredentials = &wrapperspb.BoolValue{
			Value: *credentials,
		}
	}

	if err := corsPolicy.Validate(); err != nil {
		return nil, fmt.Errorf("unable to validate cors policy: %w", err)
	}

	return corsPolicy, nil
}

func getAllowOriginsMatchers(origins []string) []*envoytypematcher.StringMatcher {
	var allowOriginsMatcher []*envoytypematcher.StringMatcher
	for _, origin := range origins {
		entry := &envoytypematcher.StringMatcher{
			// TODO: We support only exact strings, no regexp - fix this if applicable
			MatchPattern: &envoytypematcher.StringMatcher_Exact{
				Exact: origin,
			},
			IgnoreCase: false,
		}
		allowOriginsMatcher = append(allowOriginsMatcher, entry)
	}

	return allowOriginsMatcher
}
