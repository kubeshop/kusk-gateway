package options

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
)

// SubOptions allow user to overwrite certain options at path/operation level using x-kusk extension
type SubOptions struct {
	Disabled *bool `yaml:"disabled,omitempty" json:"disabled,omitempty"`
	// Service is a set of options of a target service to receive traffic.
	Service ServiceOptions `yaml:"service,omitempty" json:"service,omitempty"`
	// Path is a set of options to configure service endpoints paths.
	Path       PathOptions      `yaml:"path,omitempty" json:"path,omitempty"`
	Redirect   RedirectOptions  `yaml:"redirect,omitempty" json:"redirect,omitempty"`
	CORS       CORSOptions      `yaml:"cors,omitempty" json:"cors,omitempty"`
	RateLimits RateLimitOptions `yaml:"rate_limits,omitempty" json:"rate_limits,omitempty"`
	Timeouts   TimeoutOptions   `yaml:"timeouts,omitempty" json:"timeouts,omitempty"`
}

// RewriteRegex is used during the redirects and paths mangling
type RewriteRegex struct {
	Pattern      string `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Substitution string `yaml:"substitution,omitempty" json:"substitution,omitempty"`
}

type Options struct {
	// Top root of the configuration (top x-kusk settings) provides configuration for all paths/methods, overridable
	SubOptions
	// Host (Host header) to serve for
	// Multiple are possible
	// Hosts are not overridable per path/method intentionally since
	// there is no valid use case for such override per path in the same OpenAPI config
	Hosts []string `yaml:"hosts,omitempty" json:"hosts,omitempty"`

	// PathSubOptions allow to overwrite specific subset of Options for a given path.
	// They are filled during extension parsing, the map key is path.
	PathSubOptions map[string]SubOptions `yaml:"-" json:"-"`

	// OperationSubOptions allow to overwrite specific subset of Options for a given operation.
	// They are filled during extension parsing, the map key is method+path.
	OperationSubOptions map[string]SubOptions `yaml:"-" json:"-"`
}

func (o *Options) fillDefaults() {
	if len(o.Hosts) == 0 {
		o.Hosts = append(o.Hosts, "*")
	}
}

func (o *Options) Validate() error {
	return nil
}

func (o *Options) FillDefaultsAndValidate() error {
	o.fillDefaults()
	// TODO: validation for o.Hosts
	return v.Validate([]v.Validatable{
		o,
		&o.Service,
		&o.Path,
		&o.Redirect,
		&o.CORS,
		&o.RateLimits,
		&o.Timeouts,
	})

}
