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

package options

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// See https://github.com/projectcontour/contour/blob/main/internal/dag/dag.go#L673.

// +kubebuilder:object:generate=true
type JWT struct {
	// Providers to use for verifying JSON Web Tokens (JWTs) on the virtual host.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	JWTProviders []JWTProvider `json:"jwtProviders" yaml:"jwtProviders"`
}

func (o JWT) String() string {
	return ToCompactJSON(o)
}

func (o JWT) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.JWTProviders, validation.Required),
		validation.Field(&o.JWTProviders, validation.Each()),
	)
}

// JWTProvider defines how to verify JWTs on requests.
// +kubebuilder:object:generate=true
type JWTProvider struct {
	// Unique name for the provider.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name" yaml:"name"`
	// Whether the provider should apply to all
	// routes in the HTTPProxy/its includes by
	// default. At most one provider can be marked
	// as the default. If no provider is marked
	// as the default, individual routes must explicitly
	// identify the provider they require.
	// +optional
	Default bool `json:"default,omitempty" yaml:"default,omitempty"`
	// Issuer that JWTs are required to have in the "iss" field.
	// If not provided, JWT issuers are not checked.
	// +optional
	Issuer string `json:"issuer,omitempty" yaml:"issuer,omitempty"`
	// Audiences that JWTs are allowed to have in the "aud" field.
	// If not provided, JWT audiences are not checked.
	// +optional
	Audiences []string `json:"audiences,omitempty" yaml:"audiences,omitempty"`
	// Remote JWKS to use for verifying JWT signatures.
	// +kubebuilder:validation:Required
	RemoteJWKS RemoteJWKS `json:"remoteJWKS" yaml:"remoteJWKS"`
	// Whether the JWT should be forwarded to the backend
	// service after successful verification. By default,
	// the JWT is not forwarded.
	// +optional
	ForwardJWT bool `json:"forwardJWT,omitempty" yaml:"forwardJWT,omitempty"`
}

func (o JWTProvider) String() string {
	return ToCompactJSON(o)
}

func (o JWTProvider) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.Name, validation.Required),
		validation.Field(&o.Audiences, validation.Each()),
		validation.Field(&o.RemoteJWKS, validation.Required),
	)
}

// RemoteJWKS defines how to fetch a JWKS from an HTTP endpoint.
// +kubebuilder:object:generate=true
type RemoteJWKS struct {
	// The URI for the JWKS.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	URI string `json:"uri" yaml:"uri"`
	// UpstreamValidation defines how to verify the JWKS's TLS certificate.
	// +optional
	UpstreamValidation *UpstreamValidation `json:"validation,omitempty" yaml:"validation,omitempty"`
	// How long to wait for a response from the URI.
	// If not specified, a default of 1s applies.
	// +optional
	//// +kubebuilder:validation:Pattern=`^(((\d*(\.\d*)?h)|(\d*(\.\d*)?m)|(\d*(\.\d*)?s)|(\d*(\.\d*)?ms)|(\d*(\.\d*)?us)|(\d*(\.\d*)?µs)|(\d*(\.\d*)?ns))+)$`
	// Timeout string `json:"timeout,omitempty"`
	Timeout *time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	// How long to cache the JWKS locally. If not specified,
	// Envoy's default of 5m applies.
	// +optional
	//// +kubebuilder:validation:Pattern=`^(((\d*(\.\d*)?h)|(\d*(\.\d*)?m)|(\d*(\.\d*)?s)|(\d*(\.\d*)?ms)|(\d*(\.\d*)?us)|(\d*(\.\d*)?µs)|(\d*(\.\d*)?ns))+)$`
	// CacheDuration string `json:"cacheDuration,omitempty"`
	CacheDuration *time.Duration `json:"cacheDuration,omitempty" yaml:"cacheDuration,omitempty"`
	// The DNS IP address resolution policy for the JWKS URI.
	// When configured as "v4", the DNS resolver will only perform a lookup
	// for addresses in the IPv4 family. If "v6" is configured, the DNS resolver
	// will only perform a lookup for addresses in the IPv6 family.
	// If "auto" is configured, the DNS resolver will first perform a lookup
	// for addresses in the IPv6 family and fallback to a lookup for addresses
	// in the IPv4 family. If not specified, the Contour-wide setting defined
	// in the config file or ContourConfiguration applies (defaults to "auto").
	//
	// See https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto.html#envoy-v3-api-enum-config-cluster-v3-cluster-dnslookupfamily
	// for more information.
	// +optional
	// +kubebuilder:validation:Enum=auto;v4;v6
	DNSLookupFamily string `json:"dnsLookupFamily,omitempty" yaml:"dnsLookupFamily,omitempty"`
}

func (o RemoteJWKS) String() string {
	return ToCompactJSON(o)
}

func (o RemoteJWKS) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.URI, validation.Required),
	)
}

// UpstreamValidation defines how to verify the backend service's certificate
// +kubebuilder:object:generate=true
type UpstreamValidation struct {
	// Name or namespaced name of the Kubernetes secret used to validate the certificate presented by the backend.
	// The secret must contain key named ca.crt.
	CACertificate string `json:"caSecret" yaml:"caSecret"`
	// Key which is expected to be present in the 'subjectAltName' of the presented certificate.
	SubjectName string `json:"subjectName" yaml:"subjectName"`
}

func (o UpstreamValidation) String() string {
	return ToCompactJSON(o)
}

func (o UpstreamValidation) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.CACertificate, validation.Required),
		validation.Field(&o.SubjectName, validation.Required),
	)
}
