// package config provides structures to create and update routing configuration for Envoy Fleet
// it is not used for Fleet creation, only for configuration snapshot creation.

package config

import (
	"fmt"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	listenerName string = "listener_0"
	listenerPort uint32 = 8080
)

type listenerBuilder struct {
	lr *listener.Listener
}

func (l *listenerBuilder) Validate() error {
	return l.lr.Validate()
}

func NewListenerBuilder() *listenerBuilder {
	return &listenerBuilder{
		lr: &listener.Listener{
			Name: listenerName,
			Address: &core.Address{
				Address: &core.Address_SocketAddress{
					SocketAddress: &core.SocketAddress{
						Protocol: core.SocketAddress_TCP,
						Address:  "0.0.0.0",
						PortSpecifier: &core.SocketAddress_PortValue{
							PortValue: listenerPort,
						},
					},
				},
			},
		},
	}
}

func (l *listenerBuilder) addListenerFilterChain(c *listener.FilterChain) *listenerBuilder {
	l.lr.FilterChains = append(l.lr.FilterChains, c)
	return l
}

func (l *listenerBuilder) AddHTTPManagerFilterChain(httpConnectionManager *hcm.HttpConnectionManager) error {
	anyHTTPManagerConfig, err := anypb.New(httpConnectionManager)
	if err != nil {
		return fmt.Errorf("failed to add http manager to the filter chain: cannot convert to Any message type: %w", err)
	}
	hcmchain := &listener.FilterChain{
		Filters: []*listener.Filter{
			{
				Name:       wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{TypedConfig: anyHTTPManagerConfig},
			},
		},
	}
	l.addListenerFilterChain(hcmchain)
	return nil
}

func (l *listenerBuilder) GetListener() *listener.Listener {
	return l.lr
}
