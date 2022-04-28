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
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestGenerateCORSPolicy(t *testing.T) {
	trueValue := true

	type args struct {
		origins       []string
		methods       []string
		headers       []string
		exposeHeaders []string
		maxAge        int
		credentials   *bool
	}
	tests := []struct {
		name    string
		args    args
		want    *route.CorsPolicy
		wantErr bool
	}{
		{
			name: "credentials nil",
			args: args{
				origins:       []string{"*"},
				methods:       []string{"GET", "POST"},
				headers:       []string{"X-CUSTOM-HEADER"},
				exposeHeaders: []string{"X-CUSTOM-EXPOSE-HEADER"},
				maxAge:        10,
				credentials:   nil,
			},
			want: &route.CorsPolicy{
				AllowOriginStringMatch: []*envoytypematcher.StringMatcher{
					{
						MatchPattern: &envoytypematcher.StringMatcher_Exact{
							Exact: "*",
						},
						IgnoreCase: false,
					},
				},
				AllowMethods:     "GET,POST",
				AllowHeaders:     "X-CUSTOM-HEADER",
				ExposeHeaders:    "X-CUSTOM-EXPOSE-HEADER",
				MaxAge:           "10",
				AllowCredentials: nil,
			},
		},
		{
			name: "credentials not nil",
			args: args{
				origins:       []string{"*"},
				methods:       []string{"GET", "POST"},
				headers:       []string{"X-CUSTOM-HEADER"},
				exposeHeaders: []string{"X-CUSTOM-EXPOSE-HEADER"},
				maxAge:        10,
				credentials:   &trueValue,
			},
			want: &route.CorsPolicy{
				AllowOriginStringMatch: []*envoytypematcher.StringMatcher{
					{
						MatchPattern: &envoytypematcher.StringMatcher_Exact{
							Exact: "*",
						},
						IgnoreCase: false,
					},
				},
				AllowMethods:  "GET,POST",
				AllowHeaders:  "X-CUSTOM-HEADER",
				ExposeHeaders: "X-CUSTOM-EXPOSE-HEADER",
				MaxAge:        "10",
				AllowCredentials: &wrapperspb.BoolValue{
					Value: trueValue,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateCORSPolicy(tt.args.origins, tt.args.methods, tt.args.headers, tt.args.exposeHeaders, tt.args.maxAge, tt.args.credentials)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateCORSPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateCORSPolicy() got = %v, want %v", got, tt.want)
			}
		})
	}
}
