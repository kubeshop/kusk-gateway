/*
MIT License

Copyright (c) 2022 Kubeshop

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
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
)

// Options is the container for all parsed from the supplied configuration OpenAPI x-kusk options and any additional options
type Options struct {
	// Top root of the configuration (top x-kusk settings) provides configuration for all paths/methods
	// Overrides per method+path are in OperationFinalSubOptions map
	SubOptions
	// Host (Host header) to serve for
	// Multiple are possible
	// Hosts are not overridable per path/method intentionally since
	// there is no valid use case for such override per path in the same OpenAPI config
	Hosts []Host `yaml:"hosts,omitempty" json:"hosts,omitempty"`

	// OperationFinalSubOptions is the map of method+path key with value - merged root suboptions + path suboptions + operation suboptions
	// They are filled during the parsing of OpenAPI with extension
	OperationFinalSubOptions map[string]SubOptions `yaml:"-" json:"-"`
}

func (o *Options) FillDefaults() {
	if len(o.Hosts) == 0 {
		o.Hosts = append(o.Hosts, "*")
	}

	if o.Upstream != nil {
		o.Upstream.FillDefaults()
	}
}

func (o Options) Validate() error {
	if err := v.ValidateStruct(&o,
		v.Field(&o.Hosts, v.Each()),
		v.Field(&o.OperationFinalSubOptions, v.Each()),
	); err != nil {
		return err
	}

	// check if global options contains either mocking or an upstream service that covers all endpoints
	if o.Mocking != nil {
		err := o.Mocking.Validate()
		return err
	}

	if o.Upstream != nil {
		err := o.Upstream.Validate()
		return err
	}

	if o.Redirect != nil {
		err := o.Redirect.Validate()
		return err
	}

	// if we get here, it means there aren't global options contains either mocking or an upstream service that covers all endpoints
	// therefore we need to iterate over all operations and check if they have either mocking or an upstream service

	for pathAndMethod, op := range o.OperationFinalSubOptions {
		if op.Mocking != nil {
			if err := op.Mocking.Validate(); err != nil {
				return err
			}
		}

		if op.Upstream != nil {
			if err := op.Upstream.Validate(); err != nil {
				return err
			}
		}

		if op.Redirect != nil {
			if err := op.Redirect.Validate(); err != nil {
				return err
			}
		}

		// if we reach here then this path that doesn't have either mocking or an upstream service and is not covered by a
		// global upstream service or mocking config, so return an error
		return fmt.Errorf("no upstream or mocking configuration found for path %s", pathAndMethod)
	}

	return nil
}

// SubOptions allow user to overwrite certain options at path/operation level using x-kusk extension
type SubOptions struct {
	Disabled *bool `yaml:"disabled,omitempty" json:"disabled,omitempty"`
	// Upstream is a set of options of a target service to receive traffic.
	Upstream *UpstreamOptions `yaml:"upstream,omitempty" json:"upstream,omitempty"`
	// Redirect specifies thre redirect optins, mutually exclusive with Upstream
	Redirect *RedirectOptions `yaml:"redirect,omitempty" json:"redirect,omitempty"`
	// Path is a set of options to configure service endpoints paths.
	Path        *PathOptions       `yaml:"path,omitempty" json:"path,omitempty"`
	QoS         *QoSOptions        `yaml:"qos,omitempty" json:"qos,omitempty"`
	CORS        *CORSOptions       `yaml:"cors,omitempty" json:"cors,omitempty"`
	Websocket   *bool              `json:"websocket,omitempty" yaml:"websocket,omitempty"`
	Validation  *ValidationOptions `json:"validation,omitempty" yaml:"validation,omitempty"`
	Mocking     *MockingOptions    `json:"mocking,omitempty" yaml:"mocking,omitempty"`
	RateLimit   *RateLimitOptions  `json:"rate_limit,omitempty" yaml:"rate_limit,omitempty"`
	Cache       *CacheOptions      `json:"cache,omitempty" yaml:"cache,omitempty"`
	OpenAPIPath string             `json:"openapi-path,omitempty" yaml:"openapi-path,omitempty"`
	Auth        *AuthOptions       `json:"auth,omitempty" yaml:"auth,omitempty"`
}

func (o SubOptions) Validate() error {
	if o.Upstream != nil && o.Redirect != nil {
		return fmt.Errorf("upstream and service are mutually exclusive")
	}
	// fail if doesn't have upstream or redirect and is "enabled"
	if o.Upstream == nil && o.Redirect == nil {
		if o.Disabled != nil && !*o.Disabled {
			return fmt.Errorf("either Upstream or Service must be specified")
		}
	}

	return v.ValidateStruct(&o,
		v.Field(&o.Upstream),
		v.Field(&o.Redirect),
		v.Field(&o.Path),
		v.Field(&o.QoS),
		v.Field(&o.CORS),
		v.Field(&o.Mocking),
		v.Field(&o.Auth),
	)
}

// MergeInSubOptions handles merging other SubOptions (usually - upper level in root/path/method hierarchy)
func (o *SubOptions) MergeInSubOptions(in *SubOptions) {
	if o.Disabled == nil && in.Disabled != nil {
		o.Disabled = in.Disabled
	}
	if o.Upstream == nil && o.Redirect == nil {
		if in.Upstream != nil {
			o.Upstream = in.Upstream
		}
		if in.Redirect != nil {
			o.Redirect = in.Redirect
		}
	}
	// Path params merging
	switch {
	case o.Path == nil && in.Path != nil:
		o.Path = in.Path
	case o.Path != nil && in.Path != nil:
		if o.Path.Prefix == "" && in.Path.Prefix != "" {
			o.Path.Prefix = in.Path.Prefix
		}
	}
	// QoS params merging
	switch {
	case o.QoS == nil && in.QoS != nil:
		o.QoS = in.QoS
	case o.QoS != nil && in.QoS != nil:
		if o.QoS.IdleTimeout == 0 && in.QoS.IdleTimeout != 0 {
			o.QoS.IdleTimeout = in.QoS.IdleTimeout
		}
		if o.QoS.RequestTimeout == 0 && in.QoS.RequestTimeout != 0 {
			o.QoS.RequestTimeout = in.QoS.RequestTimeout
		}
		if o.QoS.Retries == 0 && in.QoS.Retries != 0 {
			o.QoS.Retries = in.QoS.Retries
		}
	}
	// CORS - we don't merge CORS params, we override them completely since CORS must be treated as complete entity
	if o.CORS == nil && in.CORS != nil {
		o.CORS = in.CORS
	}

	// Websockets
	if o.Websocket == nil && in.Websocket != nil {
		o.Websocket = in.Websocket
	}

	// Validation
	if o.Validation == nil && in.Validation != nil {
		o.Validation = in.Validation
	}

	// Mocking
	if o.Mocking == nil && in.Mocking != nil {
		o.Mocking = in.Mocking
	}
	// RateLimit
	if o.RateLimit == nil && in.RateLimit != nil {
		o.RateLimit = in.RateLimit
	}
	// Cache
	if o.Cache == nil && in.Cache != nil {
		o.Cache = in.Cache
	}
	// Auth
	if o.Auth == nil && in.Auth != nil {
		o.Auth = in.Auth
	}
}
