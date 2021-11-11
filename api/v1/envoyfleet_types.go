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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EnvoyFleetSpec defines the desired state of EnvoyFleet
type EnvoyFleetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Size field specifies the number of Envoy Pods being deployed. Optional, default value is 1.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default:=1
	Size *int32 `json:"size,omitempty"`
}

// EnvoyFleetStatus defines the observed state of EnvoyFleet
type EnvoyFleetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="size",type="integer",JSONPath=".spec.size"

// EnvoyFleet is the Schema for the envoyfleet API
type EnvoyFleet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvoyFleetSpec   `json:"spec"`
	Status EnvoyFleetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EnvoyFleetList contains a list of EnvoyFleet
type EnvoyFleetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EnvoyFleet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EnvoyFleet{}, &EnvoyFleetList{})
}
