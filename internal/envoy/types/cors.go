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
