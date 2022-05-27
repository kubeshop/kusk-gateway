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

type Auth struct {
	Scheme       string       `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	PathPrefix   *string      `json:"path_prefix,omitempty" yaml:"path_prefix,omitempty"`
	AuthUpstream AuthUpstream `json:"auth-upstream,omitempty" yaml:"auth-upstream,omitempty"`
}

type AuthUpstream struct {
	Host *AuthUpstreamHost `json:"host,omitempty" yaml:"host,omitempty"`
}

type AuthUpstreamHost struct {
	Hostname Host   `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	Port     uint32 `json:"port,omitempty" yaml:"port,omitempty"`
}

func (a *Auth) Validate() error {
	return nil

	// TODO(MBana): Double-check if we should do validation here.
	// return v.ValidateStruct(&a,
	// 	v.Field(&a.Scheme, v.In("http", "https")),
	// 	v.Field(&a.Scheme, v.In("basic")),
	// 	v.Field(&a.PathPrefix, is.),
	// 	v.Field(&a.AuthUpstream.Host.Hostname, is.Host),
	// 	v.Field(&a.AuthUpstream.Host.Port, is.Port),
	// 	v.Field(&o.PathRedirect, v.By(o.MutuallyExlusivePathRedirectCheck)),
	// )
}
