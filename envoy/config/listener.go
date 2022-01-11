// package config provides structures to create and update routing configuration for Envoy Fleet
// it is not used for Fleet creation, only for configuration snapshot creation.

package config

import (
	"fmt"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
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

func (l *listenerBuilder) addListenerFilter(f *listener.ListenerFilter) *listenerBuilder {
	l.lr.ListenerFilters = append(l.lr.ListenerFilters, f)
	return l
}

type Certificate struct {
	Cert string
	Key  string
}

// AddHTTPManagerFilterChains inserts HTTP Manager as the listener filter chain(s)
// If certificates are present an additional TLS-enabled filter chain is added and protocol type detection is enabled with TLS Inspector Listener filter.
func (l *listenerBuilder) AddHTTPManagerFilterChains(httpConnectionManager *hcm.HttpConnectionManager, certs []Certificate) error {
	anyHTTPManagerConfig, err := anypb.New(httpConnectionManager)
	if err != nil {
		return fmt.Errorf("failed to add http manager to the filter chain: cannot convert to Any message type: %w", err)
	}
	hcmFilter := &listener.Filter{
		Name:       wellknown.HTTPConnectionManager,
		ConfigType: &listener.Filter_TypedConfig{TypedConfig: anyHTTPManagerConfig},
	}
	// Plain HTTP manager filter chain
	hcmPlainChain := &listener.FilterChain{
		Filters: []*listener.Filter{hcmFilter},
	}
	l.addListenerFilterChain(hcmPlainChain)

	if len(certs) > 0 {
		// When certificates are present, we add an additional Listener filter chain that is selected when the connection protocol type is tls.
		// HTTP Manager configuration is the same.
		// Enable TLS Inspector in the Listener to detect plain http or tls requests.
		l.addListenerFilter(&listener.ListenerFilter{Name: wellknown.TLSInspector})

		// Make sure plain http manager filter chain is selected when protocol type is raw_buffer (not tls).
		hcmPlainChain.FilterChainMatch = &listener.FilterChainMatch{TransportProtocol: "raw_buffer"}

		// Secure (TLS) HTTP manager filter chain.
		// Selected when the connection type is tls.
		hcmSecureChain := &listener.FilterChain{
			FilterChainMatch: &listener.FilterChainMatch{TransportProtocol: "tls"},
			Filters:          []*listener.Filter{hcmFilter},
		}

		tlsCerts := make([]*tls.TlsCertificate, len(certs))
		for _, cert := range certs {
			tlsCerts = append(tlsCerts, &tls.TlsCertificate{
				CertificateChain: &core.DataSource{
					Specifier: &core.DataSource_InlineString{InlineString: cert.Cert},
				},
				PrivateKey: &core.DataSource{
					Specifier: &core.DataSource_InlineString{InlineString: cert.Key},
				},
			})
		}

		tlsDownstreamContext := &tls.DownstreamTlsContext{
			CommonTlsContext: &tls.CommonTlsContext{
				TlsCertificates: tlsCerts,
				// TODO: add cipher suites if specified
				// TlsParams:
			},
		}

		anyTls, err := anypb.New(tlsDownstreamContext)
		if err != nil {
			return fmt.Errorf("unable to marshal TLS config to typed struct: %w", err)
		}

		hcmSecureChain.TransportSocket = &core.TransportSocket{
			Name:       wellknown.TransportSocketTLS,
			ConfigType: &core.TransportSocket_TypedConfig{TypedConfig: anyTls},
		}

		l.addListenerFilterChain(hcmSecureChain)
	}

	return nil
}

func (l *listenerBuilder) GetListener() *listener.Listener {
	return l.lr
}
