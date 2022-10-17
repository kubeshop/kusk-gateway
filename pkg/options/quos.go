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
	v "github.com/go-ozzo/ozzo-validation/v4"
)

type QoSOptions struct {
	// Retries define how many times to retry calling the backend
	Retries uint32 `yaml:"retries,omitempty" json:"retries,omitempty"`
	// RequestTimeout is total request timeout
	RequestTimeout uint32 `yaml:"request_timeout,omitempty" json:"request_timeout,omitempty"`
	// IdleTimeout is timeout for idle connection
	IdleTimeout uint32 `yaml:"idle_timeout,omitempty" json:"idle_timeout,omitempty"`
}

func (o QoSOptions) Validate() error {
	return v.ValidateStruct(&o,
		v.Field(&o.Retries, v.Min(uint32(0)), v.Max(MaxRetries)),
		v.Field(&o.RequestTimeout, v.Min(uint32(0))),
		v.Field(&o.IdleTimeout, v.Min(uint32(0))),
	)
}
