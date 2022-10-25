/*
MIT License

# Copyright (c) 2022 Kubeshop

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
	"net/http"
	"reflect"
	"testing"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
)

func TestGetHeaderMatcherConfig(t *testing.T) {
	type args struct {
		methods []string
		cors    bool
	}
	tests := []struct {
		name string
		args args
		want *route.HeaderMatcher
	}{
		{
			name: "nil methods",
			args: args{
				methods: nil,
			},
			want: nil,
		},
		{
			name: "no methods",
			args: args{
				methods: []string{},
			},
			want: nil,
		},
		{
			name: "one method and no cors",
			args: args{
				methods: []string{http.MethodGet},
				cors:    false,
			},
			want: &route.HeaderMatcher{
				Name: headerMatcherName,
				HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
					StringMatch: &matcher.StringMatcher{
						MatchPattern: &matcher.StringMatcher_Exact{Exact: http.MethodGet},
					},
				},
			},
		},
		{
			name: "one method and cors",
			args: args{
				methods: []string{http.MethodGet},
				cors:    true,
			},
			want: &route.HeaderMatcher{
				Name: headerMatcherName,
				HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
					StringMatch: &matcher.StringMatcher{
						MatchPattern: &matcher.StringMatcher_SafeRegex{
							SafeRegex: &matcher.RegexMatcher{
								EngineType: &matcher.RegexMatcher_GoogleRe2{
									GoogleRe2: &matcher.RegexMatcher_GoogleRE2{},
								},
								Regex: fmt.Sprintf("^%s$|^%s$", http.MethodGet, http.MethodOptions),
							},
						},
					},
				},
			},
		},
		{
			name: "multiple methods and no cors",
			args: args{
				methods: []string{http.MethodGet, http.MethodPost},
				cors:    false,
			},
			want: &route.HeaderMatcher{
				Name: headerMatcherName,
				HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
					StringMatch: &matcher.StringMatcher{
						MatchPattern: &matcher.StringMatcher_SafeRegex{
							SafeRegex: &matcher.RegexMatcher{
								EngineType: &matcher.RegexMatcher_GoogleRe2{
									GoogleRe2: &matcher.RegexMatcher_GoogleRE2{},
								},
								Regex: fmt.Sprintf("^%s$|^%s$", http.MethodGet, http.MethodPost),
							},
						},
					},
				},
			},
		},
		{
			name: "multiple methods and cors",
			args: args{
				methods: []string{http.MethodGet, http.MethodPost},
				cors:    true,
			},
			want: &route.HeaderMatcher{
				Name: headerMatcherName,
				HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
					StringMatch: &matcher.StringMatcher{
						MatchPattern: &matcher.StringMatcher_SafeRegex{
							SafeRegex: &matcher.RegexMatcher{
								EngineType: &matcher.RegexMatcher_GoogleRe2{
									GoogleRe2: &matcher.RegexMatcher_GoogleRE2{},
								},
								Regex: fmt.Sprintf("^%s$|^%s$|^%s$", http.MethodGet, http.MethodPost, http.MethodOptions),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHeaderMatcherConfig(tt.args.methods, tt.args.cors); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetHeaderMatcherConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
