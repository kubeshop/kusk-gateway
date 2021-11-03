package options

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// BackendOptions describes backend to proxy to
type BackendOptions struct {

	// Hostname is the upstream hostname, without port.
	Hostname string `yaml:"hostname,omitempty" json:"hostname,omitempty"`

	// Port is the upstream port.
	Port uint32 `yaml:"port,omitempty" json:"port,omitempty"`

	// Rewrite is the pattern (regex) and a substitution string that will change URL when request is being forwarded
	// to the upstream service.
	// e.g. given that URL path is "/petstore/api/v3",
	// with Rewrite.Pattern "^/petstore", Rewrite.Substitution ""
	// backend will receive "/api/v3/pets".
	Rewrite RewriteRegex `yaml:"rewrite,omitempty" json:"rewrite,omitempty"`

	// Retries define how many times to retry calling the backend
	Retries uint32 `yaml:"retries,omitempty" json:"retries,omitempty"`
}

func (o BackendOptions) Validate() error {
	return v.ValidateStruct(&o,
		v.Field(&o.Hostname, is.DNSName),
		v.Field(&o.Port, v.Min(uint32(1)), v.Max(uint32(65356)), v.When(o.Hostname != "", v.Required)),
		v.Field(&o.Retries, v.Min(uint32(0)), v.Max(MaxRetries)),
		v.Field(&o.Rewrite),
	)
}
