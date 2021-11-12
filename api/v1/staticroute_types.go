/*
MIT License

Copyright (c) 2021 Kubeshop

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

package v1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubeshop/kusk-gateway/options"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StaticRouteSpec defines the desired state of StaticRoute
type StaticRouteSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Hosts is a collection of vhosts the rules apply to.
	// Default - "*" - vhost that matches all domain names.
	Hosts []options.Host `json:"hosts"`

	// Paths is a multidimensional map of path / method to the routing rules
	Paths map[Path]Methods `json:"paths"`
}

// GetOptionsFromSpec is a converter to generate Options object from StaticRoutes spec
func (spec *StaticRouteSpec) GetOptionsFromSpec() (*options.StaticOptions, error) {
	// 2 dimensional map["path"]["method"]SubOptions
	paths := make(map[string]options.StaticOperationSubOptions)
	opts := &options.StaticOptions{
		Paths: paths,
		Hosts: spec.Hosts,
	}
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate options: %w", err)
	}
	for specPath, specMethods := range spec.Paths {
		path := string(specPath)
		opts.Paths[path] = make(options.StaticOperationSubOptions)
		pathMethods := opts.Paths[path]
		for specMethod, specRouteAction := range specMethods {
			methodOpts := &options.StaticSubOptions{}
			pathMethods[specMethod] = methodOpts
			if specRouteAction.Redirect != nil {
				methodOpts.Redirect = specRouteAction.Redirect
				continue
			}
			if specRouteAction.Route != nil {
				methodOpts.Backend = *&specRouteAction.Route.Backend
				if specRouteAction.Route.CORS != nil {
					methodOpts.CORS = specRouteAction.Route.CORS.DeepCopy()
				}
				if specRouteAction.Route.Timeouts != nil {
					methodOpts.Timeouts = specRouteAction.Route.Timeouts
				}
			}
		}
	}
	return opts, opts.Validate()
}

// Path is a URL path without a query
// Must start with /, could be exact (/index.html), prefix (/front/, / in the end defines prefix), regex (/images/(\d+))
type Path string

// Methods maps Method (GET, POST) to Action
type Methods map[options.HTTPMethod]*Action

// Action is either a route to the backend or a redirect, they're mutually exclusive.
type Action struct {
	// +optional
	Route *Route `json:"route,omitempty"`
	// +optional
	Redirect *options.RedirectOptions `json:"redirect,omitempty"`
}

// Route defines a routing rule that proxies to backend
type Route struct {
	Backend *options.BackendOptions `json:"backend"`
	// +optional
	CORS *options.CORSOptions `json:"cors,omitempty"`
	// +optional
	Timeouts *options.TimeoutOptions `json:"timeouts,omitempty"`
}

// StaticRouteStatus defines the observed state of StaticRoute
type StaticRouteStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// StaticRoute is the Schema for the staticroutes API
type StaticRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StaticRouteSpec   `json:"spec,omitempty"`
	Status StaticRouteStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StaticRouteList contains a list of StaticRoute
type StaticRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StaticRoute `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StaticRoute{}, &StaticRouteList{})
}
