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

	"github.com/kubeshop/kusk-gateway/internal/cert"
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
	CipherSuites              []string
	TlsMinimumProtocolVersion string
	TlsMaximumProtocolVersion string
	Certificates              []Certificate
}

type Certificate struct {
	Cert string
	Key  string
}

func makeHTTPSFilterChain(
	certificate Certificate,
	hosts []string,
	tlsParams *tls.TlsParameters,
	anyHttpConnectionManager *anypb.Any,
) (*listener.FilterChain, error) {
	tlsCert := &tls.TlsCertificate{
		CertificateChain: &core.DataSource{
			Specifier: &core.DataSource_InlineString{InlineString: certificate.Cert},
		},
		PrivateKey: &core.DataSource{
			Specifier: &core.DataSource_InlineString{InlineString: certificate.Key},
		},
	}

	if err := tlsCert.Validate(); err != nil {
		return nil, fmt.Errorf("invalid tls certificate: %w", err)
	}

	tlsDownstreamContext := &tls.DownstreamTlsContext{
		CommonTlsContext: &tls.CommonTlsContext{
			TlsCertificates: []*tls.TlsCertificate{tlsCert},
			TlsParams:       tlsParams,
		},
	}

	if err := tlsDownstreamContext.Validate(); err != nil {
		return nil, fmt.Errorf("invalid tls downstream context: %w", err)
	}

	anyTls, err := anypb.New(tlsDownstreamContext)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal TLS config to typed struct: %w", err)
	}

	return &listener.FilterChain{
		FilterChainMatch: &listener.FilterChainMatch{
			TransportProtocol: "tls",
			ServerNames:       hosts,
		},
		Filters: []*listener.Filter{
			{
				Name:       wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{TypedConfig: anyHttpConnectionManager},
			},
		},
		TransportSocket: &core.TransportSocket{
			Name:       wellknown.TransportSocketTLS,
			ConfigType: &core.TransportSocket_TypedConfig{TypedConfig: anyTls},
		},
	}, nil
}

func getTLSParameters(tlsConfig TLS) (*tls.TlsParameters, error) {
	tlsParams := &tls.TlsParameters{}

	if len(tlsConfig.CipherSuites) > 0 {
		tlsParams.CipherSuites = tlsConfig.CipherSuites
	}

	if tlsConfig.TlsMinimumProtocolVersion != "" {
		tlsProtocolValue, ok := tls.TlsParameters_TlsProtocol_value[tlsConfig.TlsMinimumProtocolVersion]
		if !ok {
			return nil, fmt.Errorf("unsupported tls protocol version %s", tlsConfig.TlsMinimumProtocolVersion)
		}
		tlsParams.TlsMinimumProtocolVersion = tls.TlsParameters_TlsProtocol(tlsProtocolValue)
	}

	if tlsConfig.TlsMaximumProtocolVersion != "" {
		tlsProtocolValue, ok := tls.TlsParameters_TlsProtocol_value[tlsConfig.TlsMaximumProtocolVersion]
		if !ok {
			return nil, fmt.Errorf("unsupported tls protocol version %s", tlsConfig.TlsMaximumProtocolVersion)
		}
		tlsParams.TlsMaximumProtocolVersion = tls.TlsParameters_TlsProtocol(tlsProtocolValue)
	}

	if err := tlsParams.Validate(); err != nil {
		return nil, fmt.Errorf("invalid tls parameters: %w", err)
	}

	return tlsParams, nil
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

	tlsParams, err := getTLSParameters(tlsConfig)
	if err != nil {
		return fmt.Errorf("unable to get TLS Parameters: %w", err)
	}

	for _, tlsCert := range tlsConfig.Certificates {
		certChain, err := cert.DecodeCertificates([]byte(tlsCert.Cert))
		if err != nil {
			return fmt.Errorf("unable to decode certificates: %w", err)
		}

		if len(certChain) == 0 {
			return fmt.Errorf("resulting cert chain length was 0")
		}

		leafCert := certChain[0]
		if len(leafCert.DNSNames) == 0 {
			return fmt.Errorf("found certificate without SAN. All provided certificates must have at least one SAN")
		}

		filterChain, err := makeHTTPSFilterChain(tlsCert, leafCert.DNSNames, tlsParams, anyHTTPManagerConfig)
		if err != nil {
			return fmt.Errorf("unable to make HTTPS filter chain with hosts: %w", err)
		}
		l.addListenerFilterChain(filterChain)
	}

	return nil
}

func (l *listenerBuilder) GetListener() *listener.Listener {
	return l.lr
}
