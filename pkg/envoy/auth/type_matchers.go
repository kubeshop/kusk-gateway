// MIT License
//
// Copyright (c) 2022 Kubeshop
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package auth

import (
	envoy_type_matcher_v3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
)

func PathMatcherExact(exact string, ignoreCase bool) *envoy_type_matcher_v3.PathMatcher {
	return &envoy_type_matcher_v3.PathMatcher{
		Rule: &envoy_type_matcher_v3.PathMatcher_Path{
			Path: StringMatcherExact(exact, ignoreCase),
		},
	}
}

func StringMatcherContains(contains string, ignoreCase bool) *envoy_type_matcher_v3.StringMatcher {
	return &envoy_type_matcher_v3.StringMatcher{
		MatchPattern: &envoy_type_matcher_v3.StringMatcher_Contains{
			Contains: contains,
		},
		IgnoreCase: ignoreCase,
	}
}

func StringMatcherExact(exact string, ignoreCase bool) *envoy_type_matcher_v3.StringMatcher {
	return &envoy_type_matcher_v3.StringMatcher{
		MatchPattern: &envoy_type_matcher_v3.StringMatcher_Exact{
			Exact: exact,
		},
		IgnoreCase: ignoreCase,
	}
}
