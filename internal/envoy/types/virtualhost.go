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
package types

import (
	"fmt"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type VirtualHost struct {
	*route.VirtualHost
}

func NewVirtualHost(name string) *VirtualHost {
	return &VirtualHost{
		&route.VirtualHost{
			Name:    name,
			Domains: []string{},
		},
	}
}

func (v *VirtualHost) AddDomain(domain string) {
	// return early if present in the list
	for _, d := range v.Domains {
		if d == domain {
			return
		}
	}
	v.Domains = append(v.Domains, domain)
}

func (v *VirtualHost) AddRoute(r *route.Route) error {
	routeExists := func(name string, routes []*route.Route) bool {
		for _, rt := range routes {
			if rt.Name == name {
				return true
			}
		}
		return false
	}

	if routeExists(r.Name, v.Routes) {
		return fmt.Errorf("route %s already exists for vhost %s", r.Name, v.Name)
	}

	v.Routes = append(v.Routes, r)

	return nil
}
