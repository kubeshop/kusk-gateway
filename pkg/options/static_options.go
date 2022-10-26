/*
MIT License

# Copyright (c) 2022 Kubeshop

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
package options

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
)

// StaticOperationSubOptions maps method (get, post) to related static subopts
type StaticOperationSubOptions map[HTTPMethod]*SubOptions

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

	Auth *AuthOptions `json:"auth,omitempty" yaml:"auth,omitempty"`

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
