package config

import (
	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

const (
	RouteName string = "local_route"
)

type hcmBuilder struct {
	httpConnectionManager *hcm.HttpConnectionManager
}

func NewHCMBuilder() *hcmBuilder {
	return &hcmBuilder{
		httpConnectionManager: &hcm.HttpConnectionManager{
			CodecType:  hcm.HttpConnectionManager_AUTO,
			StatPrefix: "http",
			RouteSpecifier: &hcm.HttpConnectionManager_Rds{
				Rds: &hcm.Rds{
					ConfigSource:    makeConfigSource(),
					RouteConfigName: RouteName,
				},
			},
			UpgradeConfigs: []*hcm.HttpConnectionManager_UpgradeConfig{
				{
					UpgradeType: "websocket",
				},
			},
			HttpFilters: []*hcm.HttpFilter{
				{
					Name: wellknown.CORS,
				},
				{
					Name: wellknown.Router,
				},
			},
		},
	}
}

func (h *hcmBuilder) Validate() error {
	return h.httpConnectionManager.Validate()
}

func (h *hcmBuilder) AddAccessLog(al *accesslog.AccessLog) *hcmBuilder {
	h.httpConnectionManager.AccessLog = append(h.httpConnectionManager.AccessLog, al)
	return h
}

func (h *hcmBuilder) GetHTTPConnectionManager() *hcm.HttpConnectionManager {
	return h.httpConnectionManager
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
