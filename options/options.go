package options

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
)

// SubOptions allow user to overwrite certain options at path/operation level
// using x-kusk extension
type SubOptions struct {
	Disabled *bool `yaml:"disabled,omitempty" json:"disabled,omitempty"`

	Host       string           `yaml:"host,omitempty" json:"host,omitempty"`
	CORS       CORSOptions      `yaml:"cors,omitempty" json:"cors,omitempty"`
	RateLimits RateLimitOptions `yaml:"rate_limits,omitempty" json:"rate_limits,omitempty"`
	Timeouts   TimeoutOptions   `yaml:"timeouts,omitempty" json:"timeouts,omitempty"`
}

type Options struct {
	Disabled bool `yaml:"disabled,omitempty" json:"disabled,omitempty"`

	// Namespace for the generated resource. Default value is "default".
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`

	// Service is a set of options of a target service to receive traffic.
	Service ServiceOptions `yaml:"service,omitempty" json:"service,omitempty"`

	// Path is a set of options to configure service endpoints paths.
	Path PathOptions `yaml:"path,omitempty" json:"path,omitempty"`

	// Cluster is a set of cluster-wide options.
	Cluster ClusterOptions `yaml:"cluster,omitempty" json:"cluster,omitempty"`

	// Host is an ingress host rule.
	// See https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-rules for additional documentation.
	Host string `yaml:"host,omitempty" json:"host,omitempty"`

	CORS CORSOptions `yaml:"cors,omitempty" json:"cors,omitempty"`

	// NGINXIngress is a set of custom nginx-ingress options.
	NGINXIngress NGINXIngressOptions `yaml:"nginx_ingress,omitempty" json:"nginx_ingress,omitempty"`

	// PathSubOptions allow to overwrite specific subset of Options for a given path.
	// They are filled during extension parsing, the map key is path.
	PathSubOptions map[string]SubOptions `yaml:"-" json:"-"`

	// OperationSubOptions allow to overwrite specific subset of Options for a given operation.
	// They are filled during extension parsing, the map key is method+path.
	OperationSubOptions map[string]SubOptions `yaml:"-" json:"-"`

	RateLimits RateLimitOptions `yaml:"rate_limits,omitempty" json:"rate_limits,omitempty"`

	Timeouts TimeoutOptions `yaml:"timeouts,omitempty" json:"timeouts,omitempty"`
}

func (o *Options) fillDefaults() {
	if o.Namespace == "" {
		o.Namespace = "default"
	}

	if o.Path.Base == "" {
		o.Path.Base = "/"
	}

	if o.Cluster.ClusterDomain == "" {
		o.Cluster.ClusterDomain = "cluster.local"
	}

	if o.Service.Port == 0 {
		o.Service.Port = 80
	}
}

func (o *Options) Validate() error {
	err := v.ValidateStruct(o,
		v.Field(&o.Namespace, v.Required.Error("Target namespace is required")),
	)

	if err != nil {
		return err
	}

	return nil
}

func (o *Options) FillDefaultsAndValidate() error {
	o.fillDefaults()

	return v.Validate([]v.Validatable{
		o,
		&o.Service,
		&o.Path,
		&o.Cluster,
		&o.CORS,
		&o.NGINXIngress,
		&o.RateLimits,
		&o.Timeouts,
	})

}

func (o *Options) IsOperationDisabled(path, method string) bool {
	opSubOptions, ok := o.OperationSubOptions[method+path]

	// If the operation has an explicit value set, return that (takes precedence over the path level setting)
	if ok && opSubOptions.Disabled != nil {
		return *opSubOptions.Disabled
	}

	// No explicit value set for `Disabled` at the operation level, check the path level
	return o.IsPathDisabled(path)
}

func (o *Options) IsPathDisabled(path string) bool {
	pathSubOptions, ok := o.PathSubOptions[path]

	// If the path has an explicit value set, return that (takes precedence over the global level setting)
	if ok && pathSubOptions.Disabled != nil {
		return *pathSubOptions.Disabled
	}

	return o.Disabled
}
