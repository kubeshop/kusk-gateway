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
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// +kubebuilder:object:generate=true
type AuthOptions struct {
	// OPTIONAL
	// +optional
	OAuth2 *OAuth2 `json:"oauth2,omitempty" yaml:"oauth2,omitempty"`
	// OPTIONAL
	// +optional
	Custom *Custom `json:"custom,omitempty" yaml:"custom,omitempty"`
	// OPTIONAL
	// +optional
	Cloudentity *Cloudentity `json:"cloudentity,omitempty" yaml:"cloudentity,omitempty"`
	// OPTIONAL
	// +optional
	JWT *JWT `json:"jwt,omitempty" yaml:"jwt,omitempty"`
}

func (o AuthOptions) String() string {
	return ToCompactJSON(o)
}

func (o AuthOptions) Validate() error {
	if o.OAuth2 == nil && o.Custom == nil && o.Cloudentity == nil && o.JWT == nil {
		return fmt.Errorf("`auth` must have one of the following defined `oauth2`, `custom`, `cloudentity`, Cloudentity or `jwt`")
	}

	if o.OAuth2 != nil && o.Custom != nil {
		return fmt.Errorf("`auth` cannot have `custom` and `oauth2` enabled at the same time")
	}

	if o.OAuth2 != nil {
		return validation.ValidateStruct(&o, validation.Field(&o.OAuth2, validation.Required))
	}
	if o.Custom != nil {
		return validation.ValidateStruct(&o, validation.Field(&o.Custom, validation.Required))
	}
	if o.Cloudentity != nil {
		return validation.ValidateStruct(&o, validation.Field(&o.Cloudentity, validation.Required))
	}
	if o.JWT != nil {
		return validation.ValidateStruct(&o, validation.Field(&o.JWT, validation.Required))
	}

	return nil
}
