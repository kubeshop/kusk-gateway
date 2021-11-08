package options

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
)

// StaticSubOptions allow user to specify certain routing options
type StaticSubOptions struct {
	// Backends defines where trafic is proxied to
	Backends []*BackendOptions `yaml:"backend,omitempty" json:"backend,omitempty"`
	// Redirect specifies redirect option, mutually exlusive with the backend option
	Redirect *RedirectOptions `yaml:"redirect,omitempty" json:"redirect,omitempty"`
	CORS     *CORSOptions     `yaml:"cors,omitempty" json:"cors,omitempty"`
	Timeouts *TimeoutOptions  `yaml:"timeouts,omitempty" json:"timeouts,omitempty"`
}

func (s StaticSubOptions) Validate() error {
	return v.ValidateStruct(&s,
		v.Field(&s.Backends, v.Each()),
		v.Field(&s.Redirect),
		v.Field(&s.CORS),
		v.Field(&s.Timeouts),
	)
}

// StaticOperationSubOptions maps method (get, post) to related static subopts
type StaticOperationSubOptions map[HTTPMethod]*StaticSubOptions

// HTTPMethod defines HTTP Method like get, post...
// a special method "*" means "all methods"
type HTTPMethod string

type StaticOptions struct {
	// Host (Host header) to serve for
	// Multiple are possible
	// Hosts are not overridable per path/method intentionally since
	// there is no valid use case for such override per path in the same OpenAPI config
	Hosts []Host

	// PathSubOptions allow to specify a specific set of option for a given path and method.
	// The map key is path.
	Paths map[string]StaticOperationSubOptions `yaml:"-" json:"-"`
}

func (o *StaticOptions) fillDefaults() {
	if len(o.Hosts) == 0 {
		o.Hosts = append(o.Hosts, "*")
	}
}

func (o StaticOptions) Validate() error {
	return v.ValidateStruct(&o,
		v.Field(&o.Hosts, v.Each()),
		v.Field(&o.Paths, v.Each()),
	)
}

func (o *StaticOptions) FillDefaultsAndValidate() error {
	o.fillDefaults()
	return o.Validate()
}
