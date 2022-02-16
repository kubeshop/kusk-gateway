package options

import (
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// UpstreamOptions defines upstream that we proxy to
// Host and Service are mutually exclusive
type UpstreamOptions struct {
	Host    *UpstreamHost    `yaml:"host,omitempty" json:"host,omitempty"`
	Service *UpstreamService `yaml:"service,omitempty" json:"service,omitempty"`

	// Rewrite is the pattern (regex) and a substitution string that will change URL when request is being forwarded
	// to the upstream service.
	// e.g. given that Prefix is set to "/petstore/api/v3", and with
	// Rewrite.Pattern is set to "^/petstore", Rewrite.Substitution is set to ""
	// path that would be generated is "/petstore/api/v3/pets", URL that the upstream service would receive
	// is "/api/v3/pets".
	Rewrite RewriteRegex `yaml:"rewrite,omitempty" json:"rewrite,omitempty"`
}

// UpstreamHost defines any DNS hostname with port that we can proxy to, even outside of the cluster
type UpstreamHost struct {
	// Hostname is the upstream hostname, without port.
	Hostname string `yaml:"hostname" json:"hostname"`

	// Port is the upstream port.
	Port uint32 `yaml:"port" json:"port"`
}

// UpstreamService defines K8s Service in the cluster
type UpstreamService struct {
	// Name is the upstream K8s Service's name.
	Name string `yaml:"name" json:"name,omitempty"`

	// Namespace where service is located
	Namespace string `yaml:"namespace" json:"namespace"`

	// Port is the upstream K8s Service's port.
	Port uint32 `yaml:"port" json:"port"`
}

func (o UpstreamHost) Validate() error {
	return v.ValidateStruct(&o,
		v.Field(&o.Hostname, is.DNSName, v.Required),
		v.Field(&o.Port, v.Min(uint32(1)), v.Max(uint32(65356)), v.Required),
	)
}
func (o UpstreamService) Validate() error {
	return v.ValidateStruct(&o,
		v.Field(&o.Name, is.DNSName, v.Required),
		v.Field(&o.Namespace, is.DNSName, v.Required),
		v.Field(&o.Port, v.Min(uint32(1)), v.Max(uint32(65356)), v.Required),
	)
}

func (o UpstreamOptions) Validate() error {
	if o.Host != nil && o.Service != nil {
		return fmt.Errorf("Host and Service are mutually exclusive")
	}
	if o.Host == nil && o.Service == nil {
		return fmt.Errorf("at least one of Host or Service must be specified")
	}
	return v.ValidateStruct(&o,
		v.Field(&o.Host),
		v.Field(&o.Service),
		v.Field(&o.Rewrite),
	)
}

// DeepCopy creates a copy of an object
func (in *UpstreamOptions) DeepCopy() *UpstreamOptions {
	out := new(UpstreamOptions)
	*out = *in
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(UpstreamService)
		**out = **in
	}
	if in.Host != nil {
		in, out := &in.Host, &out.Host
		*out = new(UpstreamHost)
		**out = **in
	}
	return out
}
