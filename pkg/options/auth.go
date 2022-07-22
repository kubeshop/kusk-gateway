// MIT License
//
// Copyright (c) 2022 Kubeshop
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package options

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// AuthOptions example:
//
// x-kusk:
//   ...
//   auth:
//     scheme: basic
//     path_prefix: /login #optional
//     auth-upstream:
//       host:
//         hostname: example.com
//         port: 80
type AuthOptions struct {
	Scheme       string       `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	PathPrefix   *string      `json:"path_prefix,omitempty" yaml:"path_prefix,omitempty"`
	AuthUpstream AuthUpstream `json:"auth-upstream,omitempty" yaml:"auth-upstream,omitempty"`
}

func (o AuthOptions) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.Scheme, validation.Required, validation.In("basic", "cloudentity")),
		validation.Field(&o.AuthUpstream, validation.Required),
	)
}

type AuthUpstream struct {
	Host AuthUpstreamHost `json:"host,omitempty" yaml:"host,omitempty"`
}

func (o AuthUpstream) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.Host, validation.Required),
	)
}

type AuthUpstreamHost struct {
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	Port     uint32 `json:"port,omitempty" yaml:"port,omitempty"`
}

func (o AuthUpstreamHost) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.Hostname, validation.Required, is.Host),
		validation.Field(&o.Port, validation.Required), // Do not attempt to validate using `is.Port`, otherwise this error, `port: must be either a string or byte slice` occurs.
	)
}
