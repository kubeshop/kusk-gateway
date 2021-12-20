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
