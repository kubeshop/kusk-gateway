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
package config

import (
	"fmt"

	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	simplecache "github.com/envoyproxy/go-control-plane/envoy/extensions/cache/simple_http_cache/v3"
	cachev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/cache/v3"
	cors_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/cors/v3"
	ratelimit "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	router_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"

	"google.golang.org/protobuf/types/known/anypb"
)

const (
	RouteName string = "local_route"
)

type HCMBuilder struct {
	HTTPConnectionManager *hcm.HttpConnectionManager
}

func NewHCMBuilder() (*HCMBuilder, error) {
	rl := &ratelimit.LocalRateLimit{
		StatPrefix: "http_local_rate_limiter",
	}

	anyRateLimit, err := anypb.New(rl)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal ratelimit configuration: %w", err)
	}

	sc := &simplecache.SimpleHttpCacheConfig{}
	simpleCacheConfig, err := anypb.New(sc)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal SimpleHttpCache configuration: %w", err)
	}
	cc := &cachev3.CacheConfig{
		TypedConfig:        simpleCacheConfig,
		AllowedVaryHeaders: nil,
		KeyCreatorParams:   nil,
		MaxBodyBytes:       0,
	}
	cacheConfig, err := anypb.New(cc)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal cacheconfig configuration: %w", err)
	}

	cors := &cors_v3.Cors{}
	anyCORS, err := anypb.New(cors)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal Cors configuration: %w", err)
	}

	router := &router_v3.Router{}
	anyRouter, err := anypb.New(router)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal Router configuration: %w", err)
	}

	hcmBuilder := &HCMBuilder{
		HTTPConnectionManager: &hcm.HttpConnectionManager{
			CodecType:  hcm.HttpConnectionManager_AUTO,
			StatPrefix: "http",
			RouteSpecifier: &hcm.HttpConnectionManager_Rds{
				Rds: &hcm.Rds{
					ConfigSource:    makeConfigSource(),
					RouteConfigName: RouteName,
				},
			},
			HttpFilters: []*hcm.HttpFilter{
				{
					Name: "envoy.filters.http.cache",
					ConfigType: &hcm.HttpFilter_TypedConfig{
						TypedConfig: cacheConfig,
					},
				},
				{
					Name: "envoy.filters.http.local_ratelimit",
					ConfigType: &hcm.HttpFilter_TypedConfig{
						TypedConfig: anyRateLimit,
					},
				},
				{
					Name: wellknown.CORS,
					ConfigType: &hcm.HttpFilter_TypedConfig{
						TypedConfig: anyCORS,
					},
				},
				{
					Name: wellknown.Router,
					ConfigType: &hcm.HttpFilter_TypedConfig{
						TypedConfig: anyRouter,
					},
				},
			},
		},
	}

	return hcmBuilder, nil
}

// AppendFilterHTTPExternalAuthorizationFilterToStart - `HTTPExternalAuthorization` needs to come before `CORS` and `Router`,
// so append it to the start of the list.
func (h *HCMBuilder) AppendFilterHTTPExternalAuthorizationFilterToStart(anyAuthorization *anypb.Any) {
	httpExternalAuthorizationFilter := &hcm.HttpFilter{
		Name: wellknown.HTTPExternalAuthorization,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: anyAuthorization,
		},
	}

	h.HTTPConnectionManager.HttpFilters = append(
		[]*hcm.HttpFilter{
			httpExternalAuthorizationFilter,
		},
		h.HTTPConnectionManager.HttpFilters...,
	)
}

func (h *HCMBuilder) Validate() error {
	return h.HTTPConnectionManager.Validate()
}

func (h *HCMBuilder) AddAccessLog(al *accesslog.AccessLog) *HCMBuilder {
	h.HTTPConnectionManager.AccessLog = append(h.HTTPConnectionManager.AccessLog, al)
	return h
}

func (h *HCMBuilder) GetHTTPConnectionManager() *hcm.HttpConnectionManager {
	return h.HTTPConnectionManager
}

func makeConfigSource() *core.ConfigSource {
	source := &core.ConfigSource{}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	source.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
		ApiConfigSource: &core.ApiConfigSource{
			TransportApiVersion:       resource.DefaultAPIVersion,
			ApiType:                   core.ApiConfigSource_GRPC,
			SetNodeOnFirstMessageOnly: true,
			GrpcServices: []*core.GrpcService{{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"},
				},
			}},
		},
	}
	return source
}
