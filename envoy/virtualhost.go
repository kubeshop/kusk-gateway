package envoy

import (
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type VirtualHosts map[string]*VirtualHost

type VirtualHost struct {
	virtualHost *route.VirtualHost
}

func NewVirtualHost(name string) *VirtualHost {
	return &VirtualHost{
		virtualHost: &route.VirtualHost{
			Name:    name,
			Domains: []string{},
		},
	}
}

func (v *VirtualHost) AddDomain(domain string) {
	v.virtualHost.Domains = append(v.virtualHost.Domains, domain)
}

//func (v *VirtualHost) AddRoute(r *Route) error {
//	routeExists := func(name string, routes []*route.Route) bool {
//		for _, rt := range routes {
//			if rt.Name == name {
//				return true
//			}
//		}
//		return false
//	}
//
//	if routeExists(r.route.Name, v.virtualHost.Routes) {
//		return fmt.Errorf("route %s already exists for vhost %s", r.route.Name, v.virtualHost.Name)
//	}
//
//	v.virtualHost.Routes = append(v.virtualHost.Routes, r.route)
//
//	return nil
//}
