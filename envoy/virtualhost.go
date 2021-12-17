package envoy

import (
	"fmt"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type VirtualHosts map[string]*VirtualHost

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
