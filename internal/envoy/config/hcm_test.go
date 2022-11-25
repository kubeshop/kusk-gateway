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

package config

import (
	"fmt"
	"testing"

	envoy_cors_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/cors/v3"
	envoy_config_filter_http_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoy_router_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	http "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/stretchr/testify/assert"
)

func TestHTTPConnectionManagerIsRouterFilter(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    *http.HttpFilter
		expected bool
	}{
		"name_is_router": {
			input: &http.HttpFilter{
				Name: "router",
				ConfigType: &http.HttpFilter_TypedConfig{
					TypedConfig: MustMarshalAny(t, &envoy_router_v3.Router{}),
				},
			},
			expected: true,
		},
		"type_is_router": {
			input: &http.HttpFilter{
				ConfigType: &http.HttpFilter_TypedConfig{
					TypedConfig: MustMarshalAny(t, &envoy_router_v3.Router{}),
				},
			},
			expected: true,
		},
		"is_not_router_cors": {
			input: &http.HttpFilter{
				Name: "cors",
				ConfigType: &http.HttpFilter_TypedConfig{
					TypedConfig: MustMarshalAny(t, &envoy_cors_v3.Cors{}),
				},
			},
			expected: false,
		},
		"is_not_router_local_ratelimit": {
			input: &http.HttpFilter{
				Name: "local_ratelimit",
				ConfigType: &http.HttpFilter_TypedConfig{
					TypedConfig: MustMarshalAny(t, &envoy_config_filter_http_local_ratelimit_v3.LocalRateLimit{
						StatPrefix: "http",
					}),
				},
			},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(fmt.Sprintf("%s_%s", t.Name(), name), func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)

			actual := IsRouterFilter(test.input)
			assert.Equal(test.expected, actual)
		})
	}
}

func MustMarshalAny(t *testing.T, pb proto.Message) *any.Any {
	t.Helper()
	assert := assert.New(t)

	a, err := anypb.New(proto.Message(pb))
	if err != nil {
		assert.Fail(err.Error())
	}

	return a
}
