package types

import (
	"net/http"
	"reflect"
	"testing"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
)

func TestNewRouteMatcherBuilder(t *testing.T) {
	type args struct {
		path           string
		pathParameters map[string]ParamSchema
	}
	tests := []struct {
		name string
		args args
		want *RouteMatcherBuilder
	}{
		{
			name: "new route matcher builder",
			args: args{
				path: "/path",
				pathParameters: map[string]ParamSchema{
					"foo": {
						Type: "bar",
						Enum: []interface{}{},
					},
				},
			},
			want: &RouteMatcherBuilder{
				path: "/path",
				pathParameters: map[string]ParamSchema{
					"foo": {
						Type: "bar",
						Enum: []interface{}{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRouteMatcherBuilder(tt.args.path, tt.args.pathParameters); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRouteMatcherBuilder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouteMatcherBuilder_GetRouteMatcher(t *testing.T) {
	headerMatchers := []*route.HeaderMatcher{
		{
			Name: headerMatcherName,
			HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
				StringMatch: &matcher.StringMatcher{
					MatchPattern: &matcher.StringMatcher_Exact{Exact: http.MethodGet},
				},
			},
		},
	}

	type fields struct {
		path           string
		pathParameters map[string]ParamSchema
	}
	type args struct {
		headers []*route.HeaderMatcher
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *route.RouteMatch
	}{
		{
			name: "path contains no params",
			fields: fields{
				path: "/path",
			},
			args: args{
				headers: headerMatchers,
			},
			want: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Path{
					Path: "/path",
				},
				Headers: headerMatchers,
			},
		},
		{
			name: "path contains string param",
			fields: fields{
				path: "/path/{foo}",
				pathParameters: map[string]ParamSchema{
					"{foo}": {
						Type: "string",
					},
				},
			},
			args: args{
				headers: headerMatchers,
			},
			want: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_SafeRegex{
					SafeRegex: &matcher.RegexMatcher{
						EngineType: &matcher.RegexMatcher_GoogleRe2{
							GoogleRe2: &matcher.RegexMatcher_GoogleRE2{},
						},
						Regex: "/path/([.a-zA-Z0-9-]+)",
					},
				},
				Headers: headerMatchers,
			},
		},
		{
			name: "path contains integer param",
			fields: fields{
				path: "/path/{foo}",
				pathParameters: map[string]ParamSchema{
					"{foo}": {
						Type: "integer",
					},
				},
			},
			args: args{
				headers: headerMatchers,
			},
			want: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_SafeRegex{
					SafeRegex: &matcher.RegexMatcher{
						EngineType: &matcher.RegexMatcher_GoogleRe2{
							GoogleRe2: &matcher.RegexMatcher_GoogleRE2{},
						},
						Regex: "/path/([0-9]+)",
					},
				},
				Headers: headerMatchers,
			},
		},
		{
			name: "path contains enum param",
			fields: fields{
				path: "/path/{foo}",
				pathParameters: map[string]ParamSchema{
					"{foo}": {
						Enum: []interface{}{
							"one",
							"two",
							"three",
						},
					},
				},
			},
			args: args{
				headers: headerMatchers,
			},
			want: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_SafeRegex{
					SafeRegex: &matcher.RegexMatcher{
						EngineType: &matcher.RegexMatcher_GoogleRe2{
							GoogleRe2: &matcher.RegexMatcher_GoogleRE2{},
						},
						Regex: "/path/(one|two|three)",
					},
				},
				Headers: headerMatchers,
			},
		},
		{
			name: "path has trailing slash",
			fields: fields{
				path: "/path/",
			},
			args: args{
				headers: headerMatchers,
			},
			want: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{
					Prefix: "/path/",
				},
				Headers: headerMatchers,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RouteMatcherBuilder{
				path:           tt.fields.path,
				pathParameters: tt.fields.pathParameters,
			}
			if got := r.GetRouteMatcher(tt.args.headers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRouteMatcher() = %v, want %v", got, tt.want)
			}
		})
	}
}
