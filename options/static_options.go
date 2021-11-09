package options

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
)

// StaticSubOptions allow user to specify proxy routing options to a backend
type StaticSubOptions struct {
	// Backend defines where trafic is proxied to
	Backend *BackendOptions `yaml:"backend,omitempty" json:"backend,omitempty"`
	// Redirect specifies redirect option, mutually exlusive with the backend option
	Redirect *RedirectOptions `yaml:"redirect,omitempty" json:"redirect,omitempty"`
	CORS     *CORSOptions     `yaml:"cors,omitempty" json:"cors,omitempty"`
	Timeouts *TimeoutOptions  `yaml:"timeouts,omitempty" json:"timeouts,omitempty"`
}

func (s StaticSubOptions) Validate() error {
	return v.ValidateStruct(&s,
		v.Field(&s.Backend),
		v.Field(&s.Redirect),
		v.Field(&s.CORS),
		v.Field(&s.Timeouts),
	)
}

// StaticOperationSubOptions maps method (get, post) to related static subopts
type StaticOperationSubOptions map[HTTPMethod]*StaticSubOptions

// HTTPMethod defines HTTP Method like GET, POST...
type HTTPMethod string

// StaticOptions define options for single routing config - domain names to use
// and path to setup routes for.
type StaticOptions struct {
	// Host (Host header) to serve for.
	// Multiple are possible. Each Host creates Envoy's Virtual Host with Domains set to only that Host and routes specified in config.
	// I.e. multiple hosts - multiple vHosts with Domains=Host and with the same routes.
	// Note on Host header matching in Envoy:
	/* courtesy of @hzxuzhonghu (https://github.com/istio/istio/issues/35826#issuecomment-957184380)
	The virtual hosts order does not influence the domain matching order
	It is the domain matters
	Domain search order:
	1. Exact domain names: www.foo.com.
	2. Suffix domain wildcards: *.foo.com or *-bar.foo.com.
	3. Prefix domain wildcards: foo.* or foo-*.
	4. Special wildcard * matching any domain. (edited)
	*/
	Hosts []Host

	// Paths allow to specify a specific set of option for a given path and a method.
	// This is a 2-dimensional map[path][method].
	// The map key is the path and the next map key is a HTTP method (operation).
	// This closely follows OpenAPI structure, but was chosen only due to the only way to specify different routing action for
	// different methods in one YAML document.
	// E.g. if GET / goes to frontend, and POST / goes to API, you cannot specify path as a key with the different methods twice in one YAML file.
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
