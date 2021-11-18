package options

import (
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
)

// SubOptions allow user to overwrite certain options at path/operation level using x-kusk extension
type SubOptions struct {
	Disabled *bool `yaml:"disabled,omitempty" json:"disabled,omitempty"`
	// Upstream is a set of options of a target service to receive traffic.
	Upstream *UpstreamOptions `yaml:"upstream,omitempty" json:"upstream,omitempty"`
	// Redirect specifies thre redirect optins, mutually exclusive with Upstream
	Redirect *RedirectOptions `yaml:"redirect,omitempty" json:"redirect,omitempty"`
	// Path is a set of options to configure service endpoints paths.
	Path PathOptions `yaml:"path,omitempty" json:"path,omitempty"`
	QoS  *QoSOptions `yaml:"qos,omitempty" json:"qos,omitempty"`
	CORS CORSOptions `yaml:"cors,omitempty" json:"cors,omitempty"`
}

func (o SubOptions) Validate() error {
	if o.Upstream != nil && o.Redirect != nil {
		return fmt.Errorf("Upstream and Service are mutually exclusive")
	}
	return v.ValidateStruct(&o,
		v.Field(&o.Upstream),
		v.Field(&o.Redirect),
		v.Field(&o.Path),
		v.Field(&o.QoS),
		v.Field(&o.CORS),
	)
}

type Options struct {
	// Top root of the configuration (top x-kusk settings) provides configuration for all paths/methods, overridable
	SubOptions
	// Host (Host header) to serve for
	// Multiple are possible
	// Hosts are not overridable per path/method intentionally since
	// there is no valid use case for such override per path in the same OpenAPI config
	Hosts []Host `yaml:"hosts,omitempty" json:"hosts,omitempty"`

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
		v.Field(&o.Hosts, v.Each()),
		v.Field(&o.SubOptions),
		v.Field(&o.PathSubOptions, v.Each()),
		v.Field(&o.OperationSubOptions, v.Each()),
	)
}

func (o *Options) FillDefaultsAndValidate() error {
	o.fillDefaults()
	return o.Validate()
}
