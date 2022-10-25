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

import "fmt"

type RateLimitOptions struct {
	RequestsPerUnit uint32 `json:"requests_per_unit,omitempty" yaml:"requests_per_unit,omitempty"`
	Unit            string `json:"unit,omitempty" yaml:"unit,omitempty"`
	PerConnection   bool   `json:"per_connection,omitempty" yaml:"per_connection,omitempty"`
	ResponseCode    uint32 `yaml:"response_code,omitempty" json:"response_code,omitempty"`
}

func (o RateLimitOptions) Validate() error {
	switch o.Unit {
	case "minute", "second", "hour":
	default:
		return fmt.Errorf("unsupported unit '%s', must be second, minute or hour", o.Unit)
	}
	return nil
}
