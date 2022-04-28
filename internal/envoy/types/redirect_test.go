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
	"reflect"
	"testing"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoytypematcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
)

func TestGenerateRewriteRegex(t *testing.T) {
	type args struct {
		pattern      string
		substitution string
	}
	tests := []struct {
		name string
		args args
		want *envoytypematcher.RegexMatchAndSubstitute
	}{
		{
			name: "empty pattern",
			args: args{
				pattern: "",
			},
			want: nil,
		},
		{
			name: "non-empty pattern",
			args: args{
				pattern:      "pattern",
				substitution: "substitution",
			},
			want: &envoytypematcher.RegexMatchAndSubstitute{
				Pattern: &envoytypematcher.RegexMatcher{
					EngineType: &envoytypematcher.RegexMatcher_GoogleRe2{
						GoogleRe2: &envoytypematcher.RegexMatcher_GoogleRE2{},
					},
					Regex: "pattern",
				},
				Substitution: "substitution",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateRewriteRegex(tt.args.pattern, tt.args.substitution); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateRewriteRegex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRouteRedirectBuilder(t *testing.T) {
	tests := []struct {
		name string
		want *RouteRedirectBuilder
	}{
		{
			name: "get new route redirect builder",
			want: &RouteRedirectBuilder{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRouteRedirectBuilder(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRouteRedirectBuilder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouteRedirectBuilder_HostRedirect(t *testing.T) {
	type fields struct {
		redirect *route.Route_Redirect
	}
	type args struct {
		host string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *RouteRedirectBuilder
	}{
		{
			name: "redirect is nil",
			fields: fields{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{},
				},
			},
			args: args{
				host: "",
			},
			want: &RouteRedirectBuilder{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{},
				},
			},
		},
		{
			name: "non-nil redirect and empty host",
			fields: fields{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{},
				},
			},
			args: args{
				host: "",
			},
			want: &RouteRedirectBuilder{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{},
				},
			},
		},
		{
			name: "non-nil redirect and non-empty host",
			fields: fields{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{},
				},
			},
			args: args{
				host: "example.com",
			},
			want: &RouteRedirectBuilder{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{
						HostRedirect: "example.com",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RouteRedirectBuilder{
				redirect: tt.fields.redirect,
			}
			if got := r.HostRedirect(tt.args.host); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HostRedirect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouteRedirectBuilder_PathRedirect(t *testing.T) {
	type fields struct {
		redirect *route.Route_Redirect
	}
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *RouteRedirectBuilder
	}{
		{
			name: "empty path",
			fields: fields{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{},
				},
			},
			args: args{
				path: "",
			},
			want: &RouteRedirectBuilder{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{},
				},
			},
		},
		{
			name: "non-empty path",
			fields: fields{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{},
				},
			},
			args: args{
				path: "/path",
			},
			want: &RouteRedirectBuilder{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{
						PathRewriteSpecifier: &route.RedirectAction_PathRedirect{
							PathRedirect: "/path",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RouteRedirectBuilder{
				redirect: tt.fields.redirect,
			}
			if got := r.PathRedirect(tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PathRedirect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouteRedirectBuilder_PortRedirect(t *testing.T) {
	type fields struct {
		redirect *route.Route_Redirect
	}
	type args struct {
		port uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *RouteRedirectBuilder
	}{
		{
			name: "port set",
			fields: fields{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{},
				},
			},
			args: args{port: 80},
			want: &RouteRedirectBuilder{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{
						PortRedirect: 80,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RouteRedirectBuilder{
				redirect: tt.fields.redirect,
			}
			if got := r.PortRedirect(tt.args.port); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PortRedirect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouteRedirectBuilder_RegexRedirect(t *testing.T) {
	type fields struct {
		redirect *route.Route_Redirect
	}
	type args struct {
		pattern      string
		substitution string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *RouteRedirectBuilder
	}{
		{
			name: "empty pattern",
			fields: fields{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{},
				},
			},
			args: args{
				pattern: "",
			},
			want: &RouteRedirectBuilder{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{},
				},
			},
		},
		{
			name: "non-empty pattern",
			fields: fields{
				redirect: &route.Route_Redirect{},
			},
			args: args{
				pattern:      "pattern",
				substitution: "substitution",
			},
			want: &RouteRedirectBuilder{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{
						PathRewriteSpecifier: &route.RedirectAction_RegexRewrite{
							RegexRewrite: &envoytypematcher.RegexMatchAndSubstitute{
								Pattern: &envoytypematcher.RegexMatcher{
									EngineType: &envoytypematcher.RegexMatcher_GoogleRe2{
										GoogleRe2: &envoytypematcher.RegexMatcher_GoogleRE2{},
									},
									Regex: "pattern",
								},
								Substitution: "substitution",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RouteRedirectBuilder{
				redirect: tt.fields.redirect,
			}
			if got := r.RegexRedirect(tt.args.pattern, tt.args.substitution); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RegexRedirect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouteRedirectBuilder_ResponseCode(t *testing.T) {
	type fields struct {
		redirect *route.Route_Redirect
	}
	type args struct {
		code uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *RouteRedirectBuilder
	}{
		{
			name: "zero response code",
			fields: fields{
				redirect: &route.Route_Redirect{},
			},
			args: args{
				code: 0,
			},
			want: &RouteRedirectBuilder{
				redirect: &route.Route_Redirect{},
			},
		},
		{
			name: "non zero response code and code in map",
			fields: fields{
				redirect: &route.Route_Redirect{},
			},
			args: args{
				code: 301,
			},
			want: &RouteRedirectBuilder{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{
						ResponseCode: route.RedirectAction_RedirectResponseCode(
							0,
						),
					},
				},
			},
		},
		{
			name: "non zero response code and code not in map",
			fields: fields{
				redirect: &route.Route_Redirect{},
			},
			args: args{
				code: 200,
			},
			want: &RouteRedirectBuilder{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{
						ResponseCode: 301,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RouteRedirectBuilder{
				redirect: tt.fields.redirect,
			}
			if got := r.ResponseCode(tt.args.code); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResponseCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouteRedirectBuilder_SchemeRedirect(t *testing.T) {
	type fields struct {
		redirect *route.Route_Redirect
	}
	type args struct {
		scheme string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *RouteRedirectBuilder
	}{
		{
			name: "empty scheme",
			fields: fields{
				redirect: &route.Route_Redirect{},
			},
			args: args{scheme: ""},
			want: &RouteRedirectBuilder{
				redirect: &route.Route_Redirect{},
			},
		},
		{
			name: "non-empty scheme",
			fields: fields{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{},
				},
			},
			args: args{scheme: "https"},
			want: &RouteRedirectBuilder{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{
						SchemeRewriteSpecifier: &route.RedirectAction_SchemeRedirect{
							SchemeRedirect: "https",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RouteRedirectBuilder{
				redirect: tt.fields.redirect,
			}
			if got := r.SchemeRedirect(tt.args.scheme); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SchemeRedirect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouteRedirectBuilder_StripQuery(t *testing.T) {
	trueValue := true

	type fields struct {
		redirect *route.Route_Redirect
	}
	type args struct {
		stripQuery *bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *RouteRedirectBuilder
	}{
		{
			name: "nil strip query",
			fields: fields{
				redirect: &route.Route_Redirect{},
			},
			args: args{stripQuery: nil},
			want: &RouteRedirectBuilder{
				redirect: &route.Route_Redirect{},
			},
		},
		{
			name: "non-nil strip query",
			fields: fields{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{},
				},
			},
			args: args{stripQuery: &trueValue},
			want: &RouteRedirectBuilder{
				redirect: &route.Route_Redirect{
					Redirect: &route.RedirectAction{
						StripQuery: trueValue,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RouteRedirectBuilder{
				redirect: tt.fields.redirect,
			}
			if got := r.StripQuery(tt.args.stripQuery); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StripQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
