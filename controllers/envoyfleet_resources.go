package controllers

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gatewayv1alpha1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
)

const (
	envoyHTTPListenerPort  int32 = 8080
	envoyHTTPSListenerPort int32 = 8443
	envoyAdminListenerPort int32 = 19000
)

// EnvoyFleetResources is a collection of related Envoy Fleet K8s resources
type EnvoyFleetResources struct {
	fleetName    string
	configMap    *corev1.ConfigMap
	deployment   *appsv1.Deployment
	service      *corev1.Service
	commonLabels map[string]string
}

func NewEnvoyFleetResources(ef *gatewayv1alpha1.EnvoyFleet) (*EnvoyFleetResources, error) {
	f := &EnvoyFleetResources{
		fleetName: ef.Name,
		commonLabels: map[string]string{
			"app":   "kusk-gateway",
			"fleet": ef.Name,
		},
	}

	if err := f.CreateConfigMap(ef); err != nil {
		return nil, err
	}
	// Depends on the ConfigMap
	if err := f.CreateDeployment(ef); err != nil {
		return nil, err
	}
	// Depends on the Service
	if err := f.CreateService(ef); err != nil {
		return nil, err
	}
	return f, nil
}

func (e *EnvoyFleetResources) CreateConfigMap(ef *gatewayv1alpha1.EnvoyFleet) error {
	// future object labels
	labels := map[string]string{
		"component": "envoy-config",
	}
	// Copy over shared labels map
	for key, value := range e.commonLabels {
		labels[key] = value
	}

	configMapName := "kusk-envoy-config-" + e.fleetName

	e.configMap = &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            configMapName,
			Namespace:       ef.Namespace,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{envoyFleetAsOwner(ef)},
		},
		Data: map[string]string{
			"envoy-config.yaml": fmt.Sprintf(envoyConfigTemplate, e.fleetName),
		},
	}

	return nil
}

func (e *EnvoyFleetResources) CreateDeployment(ef *gatewayv1alpha1.EnvoyFleet) error {
	// future object labels
	labels := map[string]string{
		"component": "envoy",
	}
	// Copy over shared labels map
	for key, value := range e.commonLabels {
		labels[key] = value
	}

	deploymentName := "kusk-envoy-" + e.fleetName

	configMapName := e.configMap.Name

	envoyContainer := corev1.Container{
		Name:            "envoy",
		Image:           ef.Spec.Image,
		ImagePullPolicy: "IfNotPresent",
		Command:         []string{"/bin/sh", "-c"},
		Args: []string{
			"envoy -c /etc/envoy/envoy.yaml --service-node $POD_NAME",
		},
		Env: []corev1.EnvVar{
			{
				Name: "POD_NAME",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "envoy-config",
				MountPath: "/etc/envoy/envoy.yaml",
				SubPath:   "envoy-config.yaml",
			},
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: envoyHTTPListenerPort,
			},
			{
				Name:          "https",
				ContainerPort: envoyHTTPSListenerPort,
			},
			{
				Name:          "admin",
				ContainerPort: envoyAdminListenerPort,
			},
		},
	}
	// Set Enovy Pod Resources if specified
	if ef.Spec.Resources != nil {
		if ef.Spec.Resources.Limits != nil {
			envoyContainer.Resources.Limits = *&ef.Spec.Resources.Limits
		}
		if ef.Spec.Resources.Requests != nil {
			envoyContainer.Resources.Requests = *&ef.Spec.Resources.Requests
		}
	}
	e.deployment = &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            deploymentName,
			Namespace:       ef.Namespace,
			Labels:          labels,
			Annotations:     ef.Spec.Annotations,
			OwnerReferences: []metav1.OwnerReference{envoyFleetAsOwner(ef)},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ef.Spec.Size,
			Selector: labelSelectors(labels),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						envoyContainer,
					},
					Volumes: []corev1.Volume{
						{
							Name: "envoy-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMapName,
									},
								},
							},
						},
					},
					NodeSelector:                  ef.Spec.NodeSelector,
					Affinity:                      ef.Spec.Affinity,
					Tolerations:                   ef.Spec.Tolerations,
					TerminationGracePeriodSeconds: ef.Spec.TerminationGracePeriodSeconds,
				},
			},
		},
	}
	return nil
}

func (f *EnvoyFleetResources) CreateService(ef *gatewayv1alpha1.EnvoyFleet) error {
	// future object labels
	labels := map[string]string{
		"component": "envoy-svc",
	}
	// Copy over shared labels map
	for key, value := range f.commonLabels {
		labels[key] = value
	}
	serviceName := "kusk-envoy-svc-" + ef.Name

	f.service = &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            serviceName,
			Namespace:       ef.Namespace,
			Labels:          labels,
			Annotations:     ef.Spec.Service.Annotations,
			OwnerReferences: []metav1.OwnerReference{envoyFleetAsOwner(ef)},
		},
		Spec: corev1.ServiceSpec{
			Ports:    ef.Spec.Service.Ports,
			Selector: f.deployment.Spec.Selector.MatchLabels,
			Type:     ef.Spec.Service.Type,
		},
	}
	// Static IP address for the LoadBalancer
	if ef.Spec.Service.Type == corev1.ServiceTypeLoadBalancer && ef.Spec.Service.LoadBalancerIP != "" {
		f.service.Spec.LoadBalancerIP = ef.Spec.Service.LoadBalancerIP
	}

	return nil
}
func envoyFleetAsOwner(cr *gatewayv1alpha1.EnvoyFleet) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: cr.APIVersion,
		Kind:       cr.Kind,
		Name:       cr.Name,
		UID:        cr.UID,
		Controller: &trueVar,
	}
}

func labelSelectors(labels map[string]string) *metav1.LabelSelector {
	return &metav1.LabelSelector{MatchLabels: labels}
}

var envoyConfigTemplate = `
node:
  cluster: %s

dynamic_resources:
  ads_config:
    api_type: GRPC
    transport_api_version: V3
    grpc_services:
    - envoy_grpc:
        cluster_name: xds_cluster
  cds_config:
    resource_api_version: V3
    ads: {}
  lds_config:
    resource_api_version: V3
    ads: {}

static_resources:
  clusters:
  - type: STRICT_DNS
    typed_extension_protocol_options:
      envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
        "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
        explicit_http_config:
          http2_protocol_options: {}
    name: xds_cluster
    load_assignment:
      cluster_name: xds_cluster
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: kusk-xds-service.kusk-system.svc.cluster.local
                port_value: 18000

admin:
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 19000

`
