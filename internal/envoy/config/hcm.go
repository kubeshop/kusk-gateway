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
package config

import (
	"errors"
	"fmt"
	"time"

	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	envoy_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	simplecache "github.com/envoyproxy/go-control-plane/envoy/extensions/cache/simple_http_cache/v3"
	cachev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/cache/v3"
	cors_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/cors/v3"
	extproc "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_proc/v3"
	ratelimit "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoy_extensions_filters_http_router_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	router_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
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

	proc := &extproc.ExternalProcessor{
		FailureModeAllow: false,
		GrpcService: &envoy_core_v3.GrpcService{
			TargetSpecifier: &envoy_core_v3.GrpcService_GoogleGrpc_{
				GoogleGrpc: &envoy_core_v3.GrpcService_GoogleGrpc{
					TargetUri:  "kusk-gateway-validator-service.kusk-system.svc.cluster.local:17000",
					StatPrefix: "external_proc",
				},
			},
			Timeout: nil,
		},
		ProcessingMode: &extproc.ProcessingMode{
			RequestHeaderMode:   extproc.ProcessingMode_SKIP,
			ResponseHeaderMode:  extproc.ProcessingMode_SKIP,
			RequestBodyMode:     extproc.ProcessingMode_NONE,
			ResponseBodyMode:    extproc.ProcessingMode_NONE,
			RequestTrailerMode:  extproc.ProcessingMode_SKIP,
			ResponseTrailerMode: extproc.ProcessingMode_SKIP,
		},
		AsyncMode:      false,
		MessageTimeout: nil,
		StatPrefix:     "",
	}

	anyExternal, err := anypb.New(proc)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal External Processor configuration: %w", err)
	}

	return &HCMBuilder{
		HTTPConnectionManager: &hcm.HttpConnectionManager{
			CodecType:  hcm.HttpConnectionManager_AUTO,
			StatPrefix: "http",
			RouteSpecifier: &hcm.HttpConnectionManager_Rds{
				Rds: &hcm.Rds{
					ConfigSource:    ConfigSource("xds_cluster"),
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
					Name: "envoy.filters.http.ext_proc",
					ConfigType: &hcm.HttpFilter_TypedConfig{
						TypedConfig: anyExternal,
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
	}, nil

}

func (h *HCMBuilder) ValidateAll() error {
	return h.HTTPConnectionManager.ValidateAll()
}

func (h *HCMBuilder) AddAccessLog(al *accesslog.AccessLog) *HCMBuilder {
	h.HTTPConnectionManager.AccessLog = append(h.HTTPConnectionManager.AccessLog, al)
	return h
}

func (h *HCMBuilder) GetHTTPConnectionManager() *hcm.HttpConnectionManager {
	return h.HTTPConnectionManager
}

// AddFilter appends f to the list of filters for this HTTPConnectionManager.
// f may be nil, in which case it is ignored and an error will be returned.
// Note that Router filters (filters with TypeUrl `type.googleapis.com/envoy.extensions.filters.http.router.v3.Router`)
// are specially treated. There may only be one of these filters, and it must be the last.
// AddFilter will ensure that the router filter, if present, is last, and will panic
// if a second Router is added when one is already present.
func (h *HCMBuilder) AddFilter(newFilter *hcm.HttpFilter) error {
	// Taken from <https://github.com/projectcontour/contour/blob/3fc56275f51db51c64d33f50e052efd4de3349de/internal/envoy/v3/listener.go#L338> and adjusted.
	if newFilter == nil {
		return errors.New("config.HCMBuilder.AddFilter: cannot append nil filter")
	}

	h.HTTPConnectionManager.HttpFilters = append(h.HTTPConnectionManager.HttpFilters, newFilter)

	if len(h.HTTPConnectionManager.HttpFilters) == 1 {
		return nil
	}

	lastIndex := len(h.HTTPConnectionManager.HttpFilters) - 1
	routerIndex := -1
	for i, filter := range h.HTTPConnectionManager.HttpFilters {
		if IsRouterFilter(filter) {
			routerIndex = i
			break
		}
	}

	// We can't add more than one router entry, and there should be no way to do it.
	// If this happens, it has to be programmer error, so we return an error to tell them
	// it needs to be fixed. Note that in hitting this case, it doesn't matter we added
	// the second one earlier, because we're return an erroring anyway.
	if routerIndex != -1 && IsRouterFilter(newFilter) {
		return errors.New("config.HCMBuilder.AddFilter: cannot add more than one router to a filter chain")
	}

	if routerIndex != lastIndex {
		// Move the router to the end of the filters array.
		routerFilter := h.HTTPConnectionManager.HttpFilters[routerIndex]
		h.HTTPConnectionManager.HttpFilters = append(h.HTTPConnectionManager.HttpFilters[:routerIndex], h.HTTPConnectionManager.HttpFilters[routerIndex+1])
		h.HTTPConnectionManager.HttpFilters = append(h.HTTPConnectionManager.HttpFilters, routerFilter)
	}

	return nil
}

func IsRouterFilter(filter *hcm.HttpFilter) bool {
	return filter.GetTypedConfig().MessageIs(&envoy_extensions_filters_http_router_v3.Router{}) || filter.Name == wellknown.Router
}

// ConfigSource returns a *envoy_core_v3.ConfigSource for cluster.
func ConfigSource(cluster string) *envoy_core_v3.ConfigSource {
	const (
		defaultRequestTimeout  = time.Minute * 4
		defaultResponseTimeout = time.Minute * 4
	)

	source := &envoy_core_v3.ConfigSource{
		ResourceApiVersion: envoy_core_v3.ApiVersion_V3,
		ConfigSourceSpecifier: &envoy_core_v3.ConfigSource_ApiConfigSource{
			ApiConfigSource: &envoy_core_v3.ApiConfigSource{
				TransportApiVersion:       envoy_core_v3.ApiVersion_V3,
				ApiType:                   envoy_core_v3.ApiConfigSource_GRPC,
				RequestTimeout:            durationpb.New(defaultRequestTimeout),
				SetNodeOnFirstMessageOnly: true,
				GrpcServices: []*envoy_core_v3.GrpcService{
					makeGrpcService(cluster, "", defaultResponseTimeout),
				},
			},
		},
	}

	return source
}

// GRPCService returns a envoy_core_v3.makeGrpcService for the given parameters.
func makeGrpcService(clusterName, sni string, timeout time.Duration) *envoy_core_v3.GrpcService {
	return &envoy_core_v3.GrpcService{
		TargetSpecifier: &envoy_core_v3.GrpcService_EnvoyGrpc_{
			EnvoyGrpc: &envoy_core_v3.GrpcService_EnvoyGrpc{
				ClusterName: clusterName,
			},
		},
		Timeout: durationpb.New(timeout),
	}
}
