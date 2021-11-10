/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"github.com/kubeshop/kusk-gateway/options"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StaticRouteSpec defines the desired state of StaticRoute
type StaticRouteSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Hosts is a collection of vhosts the rules apply to
	Hosts []options.Host `json:"hosts"`

	// Paths is a multidimensional map of path / method to the routing rules
	Paths map[Path]Methods `json:"paths"`
}

// Path is a URL path without a query
// Must start with /, could be exact (/index.html), prefix (/front/, / in the end defines prefix), regex (/images/(\d+))
type Path string

// Methods maps Method (GET, POST) to Action
type Methods map[options.HTTPMethod]*Action

// Action is either a route to the backend or a redirect, they're mutually exclusive.
type Action struct {
	Route    *Route                   `json:"route,omitempty"`
	Redirect *options.RedirectOptions `json:"redirect,omitempty"`
}

// Route defines a routing rule that proxies to backend
type Route struct {
	Backend  *options.BackendOptions `json:"backend"`
	CORS     *options.CORSOptions    `json:"cors,omitempty"`
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
