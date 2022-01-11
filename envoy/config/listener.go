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

type TLS struct {
	CipherSuite               string
	TlsMinimumProtocolVersion string
	TlsMaximumProtocolVersion string
	Certificates              []Certificate
}

type Certificate struct {
	Cert string
	Key  string
}

// AddHTTPManagerFilterChains inserts HTTP Manager as the listener filter chain(s)
// If certificates are present an additional TLS-enabled filter chain is added and protocol type detection is enabled with TLS Inspector Listener filter.
func (l *listenerBuilder) AddHTTPManagerFilterChains(httpConnectionManager *hcm.HttpConnectionManager, tlsConfig TLS) error {
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

	if len(tlsConfig.Certificates) == 0 {
		return nil
	}

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

	tlsCerts := make([]*tls.TlsCertificate, len(tlsConfig.Certificates))
	for _, cert := range tlsConfig.Certificates {

		tlsCert := &tls.TlsCertificate{
			CertificateChain: &core.DataSource{
				Specifier: &core.DataSource_InlineString{InlineString: cert.Cert},
			},
			PrivateKey: &core.DataSource{
				Specifier: &core.DataSource_InlineString{InlineString: cert.Key},
			},
		}

		if err := tlsCert.Validate(); err != nil {
			return fmt.Errorf("invalid tls certificate: %w", err)
		}

		tlsCerts = append(tlsCerts, tlsCert)
	}

	tlsParams := &tls.TlsParameters{}

	if tlsConfig.CipherSuite != "" {
		tlsParams.CipherSuites = []string{tlsConfig.CipherSuite}
	}

	if tlsConfig.TlsMinimumProtocolVersion != "" {
		tlsProtocolValue, ok := tls.TlsParameters_TlsProtocol_value[tlsConfig.TlsMinimumProtocolVersion]
		if !ok {
			return fmt.Errorf("unsupported tls protocol version %s", tlsConfig.TlsMinimumProtocolVersion)
		}
		tlsParams.TlsMinimumProtocolVersion = tls.TlsParameters_TlsProtocol(tlsProtocolValue)
	}

	if tlsConfig.TlsMaximumProtocolVersion != "" {
		tlsProtocolValue, ok := tls.TlsParameters_TlsProtocol_value[tlsConfig.TlsMaximumProtocolVersion]
		if !ok {
			return fmt.Errorf("unsupported tls protocol version %s", tlsConfig.TlsMaximumProtocolVersion)
		}
		tlsParams.TlsMinimumProtocolVersion = tls.TlsParameters_TlsProtocol(tlsProtocolValue)
	}

	if err := tlsParams.Validate(); err != nil {
		return fmt.Errorf("invalid tls parameters: %w", err)
	}

	tlsDownstreamContext := &tls.DownstreamTlsContext{
		CommonTlsContext: &tls.CommonTlsContext{
			TlsCertificates: tlsCerts,
			TlsParams:       tlsParams,
		},
	}

	if err := tlsDownstreamContext.Validate(); err != nil {
		return fmt.Errorf("invalid tls downstream context: %w", err)
	}

	anyTls, err := anypb.New(tlsDownstreamContext)
	if err != nil {
		return fmt.Errorf("unable to marshal TLS config to typed struct: %w", err)
	}

	hcmSecureChain.TransportSocket = &core.TransportSocket{
		Name:       wellknown.TransportSocketTLS,
		ConfigType: &core.TransportSocket_TypedConfig{TypedConfig: anyTls},
	}

	if err := hcmPlainChain.Validate(); err != nil {
		return fmt.Errorf("invalid secure listener chain: %w", err)
	}

	l.addListenerFilterChain(hcmSecureChain)

	return nil
}

func (l *listenerBuilder) GetListener() *listener.Listener {
	return l.lr
}
