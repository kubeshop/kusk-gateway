package options

import (
	"fmt"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
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

func (s SubOptions) Validate() error {
	return v.ValidateStruct(&s,
		v.Field(&s.Service),
		v.Field(&s.Path),
		v.Field(&s.Redirect),
		v.Field(&s.CORS),
		v.Field(&s.RateLimits),
		v.Field(&s.Timeouts),
	)
}

// RewriteRegex is used during the redirects and paths mangling
type RewriteRegex struct {
	Pattern      string `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Substitution string `yaml:"substitution,omitempty" json:"substitution,omitempty"`
}

func (r RewriteRegex) Validate() error {
	if r.Substitution != "" && r.Pattern == "" {
		return fmt.Errorf("rewrite regex pattern must be specified if the substitution is defined")
	}
	return nil
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

func (o Options) Validate() error {
	return v.ValidateStruct(&o,
		v.Field(&o.Hosts, v.Each(v.By(validateHost))),
		v.Field(&o.SubOptions),
		v.Field(&o.PathSubOptions, v.Each()),
		v.Field(&o.OperationSubOptions, v.Each()),
	)
}

// validation helper to check vHost
// similarly to Envoy we support preffix and suffix wildcard (*), single wildcard and usual hostnames
func validateHost(value interface{}) error {
	host, ok := value.(string)
	if !ok {
		return fmt.Errorf("validatable object must be the string")
	}
	// We're looking for *, either single, in the preffix and in the suffix.
	// For preffix and suffix * is replaced with "w" to form the DNS name that we can validate (e.g. "*-site.com", "site.*").
	switch {
	case host == "*":
		return nil
	case strings.HasPrefix(host, "*") && strings.HasSuffix(host, "*"):
		return fmt.Errorf("wildcards are not supported on both the start AND the end of hostname")
	case strings.HasPrefix(host, "*"):
		host = strings.TrimPrefix(host, "*")
		host = "w" + host
	case strings.HasSuffix(host, "*"):
		host = strings.TrimSuffix(host, "*")
		host = host + "w"
	}
	// Finally check if this is DNS name
	return v.Validate(host, is.DNSName)
}

func (o *Options) FillDefaultsAndValidate() error {
	o.fillDefaults()
	return o.Validate()
}
