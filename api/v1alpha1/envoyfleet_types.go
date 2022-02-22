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

package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EnvoyFleetSpec defines the desired state of EnvoyFleet
type EnvoyFleetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Service describes Envoy K8s service settings
	Service *ServiceConfig `json:"service"`

	// Envoy image tag
	Image string `json:"image"`
	// Node Selector is used to schedule the Envoy pod(s) to the specificly labeled nodes, optional
	// This is the map of "key: value" labels (e.g. "disktype": "ssd")
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Affinity is used to schedule Envoy pod(s) to specific nodes, optional
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Tolerations allow pod to be scheduled to the nodes that has specific toleration labels, optional
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Optional duration in seconds the pod needs to terminate gracefully. May be decreased in delete request.
	// Value must be non-negative integer. The value zero indicates stop immediately via
	// the kill signal (no opportunity to shut down).
	// If this value is nil, the default grace period will be used instead.
	// The grace period is the duration in seconds after the processes running in the pod are sent
	// a termination signal and the time when the processes are forcibly halted with a kill signal.
	// Set this value longer than the expected cleanup time for your process.
	// Defaults to 30 seconds.
	// +optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`
	// Additional Envoy Deployment annotations, optional
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// Resources allow to set CPU and Memory resource requests and limits, optional
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// Size field specifies the number of Envoy Pods being deployed. Optional, default value is 1.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default:=1
	Size *int32 `json:"size,omitempty"`

	// Access logging settings for the Envoy
	AccessLog *AccessLoggingConfig `json:"accesslog,omitempty"`

	// TLS configuration
	//+optional
	TLS TLS `json:"tls,omitempty"`

	// Helper sidecar configuration
	//+optional
	Helper *HelperSpec `json:"helper,omitempty"`
}

type HelperSpec struct {
	// Helper sidecar image tag.
	// If empty (most of the cases) - will be detected from the Kusk Gateway Manager version and default Kubeshop container repository
	//+optional
	Image string `json:"image,omitempty"`

	// Helper sidecar CPU and Memory resources requests and limits
	//+optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

type ServiceConfig struct {

	// Kubernetes service type: NodePort, ClusterIP or LoadBalancer
	// +kubebuilder:validation:Enum=NodePort;ClusterIP;LoadBalancer
	Type corev1.ServiceType `json:"type"`

	// Kubernetes Service ports
	Ports []corev1.ServicePort `json:"ports"`

	// Service's annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Static ip address for the LoadBalancer type if available
	// +optional
	LoadBalancerIP string `json:"loadBalancerIP,omitempty"`

	// externalTrafficPolicy denotes if this Service desires to route external
	// traffic to node-local or cluster-wide endpoints. "Local" preserves the
	// client source IP and avoids a second hop for LoadBalancer and Nodeport
	// type services, but risks potentially imbalanced traffic spreading.
	// "Cluster" obscures the client source IP and may cause a second hop to
	// another node, but should have good overall load-spreading.
	// For the preservation of the real client ip in access logs chose "Local"
	// +optional
	// +kubebuilder:validation:Enum=Cluster;Local
	ExternalTrafficPolicy corev1.ServiceExternalTrafficPolicyType `json:"externalTrafficPolicy,omitempty"`
}

// AccessLoggingConfig defines the access logs Envoy logging settings
type AccessLoggingConfig struct {
	// Stdout logging format - text for unstructured and json for the structured type of logging
	// +kubebuilder:validation:Enum=json;text
	Format string `json:"format"`

	// Logging format template for the unstructured text type.
	// See https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage for the usage.
	// Uses Kusk Gateway defaults if not specified.
	// +optional
	TextTemplate string `json:"text_template,omitempty"`

	// Logging format template for the structured json type.
	// See https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage for the usage.
	// Uses Kusk Gateway defaults if not specified.
	// +optional
	JsonTemplate map[string]string `json:"json_template,omitempty"`
}

type TLS struct {
	// +optional
	// If specified, the TLS listener will only support the specified cipher list when negotiating TLS 1.0-1.2 (this setting has no effect when negotiating TLS 1.3).
	// If not specified, a default list will be used. Defaults are different for server (downstream) and client (upstream) TLS configurations.
	// For more information see: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/transport_sockets/tls/v3/common.proto
	CipherSuites []string `json:"cipherSuites,omitempty"`

	// +optional
	// Minimum TLS protocol version. By default, it’s TLSv1_2 for clients and TLSv1_0 for servers.
	TlsMinimumProtocolVersion string `json:"tlsMinimumProtocolVersion,omitempty"`

	// +optional
	// Maximum TLS protocol version. By default, it’s TLSv1_2 for clients and TLSv1_3 for servers.
	TlsMaximumProtocolVersion string `json:"tlsMaximumProtocolVersion,omitempty"`

	// SecretName and Namespace combinations for locating TLS secrets containing TLS certificates
	// You can specify more than one
	// For more information on how certificate selection works see: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/security/ssl#certificate-selection
	TlsSecrets []TLSSecrets `json:"tlsSecrets"`

	// +optional
	// TLS requirement for hosts managed by EnvoyFleet
	// NONE - No TLS requirement for the virtual hosts
	// EXTERNAL_ONLY - External requests must use TLS. If a request is external and it is not
	//	using TLS, a 301 redirect will be sent telling the client to use HTTPS.
	// ALL - All requests must use TLS. If a request is not using TLS, a 301 redirect
	//	will be sent telling the client to use HTTPS.
	// Defaults to NONE
	Requirement string `json:"requirement"`
}

type TLSSecrets struct {
	// Name of the Kubernetes secret containing the TLS certificate
	SecretRef string `json:"secretRef"`

	// Namespace where the Kubernetes certificate resides
	Namespace string `json:"namespace"`
}

// EnvoyFleetStatus defines the observed state of EnvoyFleet
type EnvoyFleetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// State indicates Envoy Fleet state
	State string `json:"state,omitempty"`
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

// EnvoyFleetID is used to bind other CR configurations to the deployed Envoy Fleet
// Consists of EnvoyFleet CR name and namespace
type EnvoyFleetID struct {

	//+kubebuilder:validation:Pattern:="^[a-z0-9-]{1,62}$"
	// deployed Envoy Fleet CR name
	Name string `json:"name"`

	//+kubebuilder:validation:Pattern:="^[a-z0-9-]{1,62}$"
	// deployed Envoy Fleet CR namespace
	Namespace string `json:"namespace"`
}

func (e EnvoyFleetID) String() string {
	return fmt.Sprintf("%s.%s", e.Name, e.Namespace)
}
