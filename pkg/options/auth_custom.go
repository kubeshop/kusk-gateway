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

const (
	SchemeCloudEntity = "cloudentity"
)

// +kubebuilder:object:generate=true
type Custom struct {
	// REQUIRED.
	Host AuthUpstreamHost `json:"host,omitempty" yaml:"host,omitempty"`
	// OPTIONAL.
	PathPrefix *string `json:"path_prefix,omitempty" yaml:"path_prefix,omitempty"`
}

func (o Custom) String() string {
	return ToCompactJSON(o)
}

func (o Custom) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.Host, validation.Required),
	)
}

// +kubebuilder:object:generate=true
type Cloudentity struct {
	// REQUIRED.
	Host AuthUpstreamHost `json:"host,omitempty" yaml:"host,omitempty"`
	// OPTIONAL.
	PathPrefix *string `json:"path_prefix,omitempty" yaml:"path_prefix,omitempty"`
}

func (o Cloudentity) String() string {
	return ToCompactJSON(o)
}

func (o Cloudentity) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.Host, validation.Required),
	)
}

// +kubebuilder:object:generate=true
type AuthUpstreamHost struct {
	// REQUIRED.
	// +required
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	// REQUIRED.
	// +required
	Port uint32 `json:"port,omitempty" yaml:"port,omitempty"`
	// OPTIONAL.
	// +optional
	Path *string `json:"path,omitempty" yaml:"path,omitempty"`
}

func (o AuthUpstreamHost) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.Hostname, validation.Required, is.Host),
		// Do not attempt to validate with `is.Port`, otherwise `port: must be either a string or byte slice` error is returned.
		validation.Field(&o.Port, validation.Required),
	)
}
