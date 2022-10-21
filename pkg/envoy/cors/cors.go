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

package cors

import (
	"strings"

	"github.com/davecgh/go-spew/spew"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/go-logr/logr"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	HeaderOrigin                        = "Origin"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"
)

func ConfigureCORSOnRoute(logger logr.Logger, corsPolicy *envoy_config_route_v3.CorsPolicy, route *envoy_config_route_v3.Route, origins []string) {
	if corsPolicy == nil {
		return
	}

	// Initialize if empty.
	if route.RequestHeadersToAdd == nil {
		route.RequestHeadersToAdd = []*envoy_config_core_v3.HeaderValueOption{}
	}
	if route.ResponseHeadersToAdd == nil {
		route.ResponseHeadersToAdd = []*envoy_config_core_v3.HeaderValueOption{}
	}

	// Request headers to add for upstream.
	for _, origin := range origins {
		route.RequestHeadersToAdd = append(route.RequestHeadersToAdd, &envoy_config_core_v3.HeaderValueOption{
			Header: &envoy_config_core_v3.HeaderValue{
				Key:   HeaderOrigin,
				Value: origin,
			},
			Append: wrapperspb.Bool(true),
		})
	}

	// Response headers to add.
	// for _, origin := range origins {
	route.ResponseHeadersToAdd = append(route.ResponseHeadersToAdd, &envoy_config_core_v3.HeaderValueOption{
		Header: &envoy_config_core_v3.HeaderValue{
			Key:   HeaderAccessControlAllowOrigin,
			Value: strings.Join(origins, ","),
			// Value: origin,
		},
		Append: wrapperspb.Bool(false),
	})
	// }

	if corsPolicy.MaxAge != "" {
		route.ResponseHeadersToAdd = append(route.ResponseHeadersToAdd, &envoy_config_core_v3.HeaderValueOption{
			Header: &envoy_config_core_v3.HeaderValue{
				Key:   HeaderAccessControlMaxAge,
				Value: corsPolicy.MaxAge,
			},
			Append: wrapperspb.Bool(false),
		})
	}

	if corsPolicy.AllowCredentials != nil && corsPolicy.AllowCredentials.Value {
		route.ResponseHeadersToAdd = append(route.ResponseHeadersToAdd, &envoy_config_core_v3.HeaderValueOption{
			Header: &envoy_config_core_v3.HeaderValue{
				Key:   HeaderAccessControlAllowCredentials,
				Value: "true",
			},
			Append: wrapperspb.Bool(false),
		})
	}

	logger.Info("ConfigureCORSOnRoute", "origins", origins)
	logger.Info("ConfigureCORSOnRoute", "route.RequestHeadersToAdd", spew.Sprint(route.RequestHeadersToAdd))
	logger.Info("ConfigureCORSOnRoute", "route.ResponseHeadersToAdd", spew.Sprint(route.ResponseHeadersToAdd))
}
