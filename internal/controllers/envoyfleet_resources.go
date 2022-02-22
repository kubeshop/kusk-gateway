package controllers

import (
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gateway "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/internal/k8sutils"
)

const (
	envoyHTTPListenerPort       int32  = 8080
	envoyAdminListenerPort      int32  = 19000
	helperImageName             string = "kusk-gateway-helper"
	kuskGatewayManagerImageName        = "kusk-gateway"
)

// EnvoyFleetResources is a collection of related Envoy Fleet K8s resources
type EnvoyFleetResources struct {
	client       client.Client
	fleet        *gateway.EnvoyFleet
	fleetID      string
	configMap    *corev1.ConfigMap
	deployment   *appsv1.Deployment
	service      *corev1.Service
	sharedLabels map[string]string
}

func NewEnvoyFleetResources(ctx context.Context, client client.Client, ef *gateway.EnvoyFleet) (*EnvoyFleetResources, error) {
	fleetID := gateway.EnvoyFleetID{Name: ef.Name, Namespace: ef.Namespace}.String()
	f := &EnvoyFleetResources{
		client:  client,
		fleet:   ef,
		fleetID: fleetID,
		sharedLabels: map[string]string{
			"app.kubernetes.io/name":       "kusk-gateway-envoy-fleet",
			"app.kubernetes.io/managed-by": "kusk-gateway-manager",
			"app.kubernetes.io/created-by": "kusk-gateway-manager",
			"app.kubernetes.io/part-of":    "kusk-gateway",
			"app.kubernetes.io/instance":   ef.Name,
			"fleet":                        fleetID,
		},
	}

	if err := f.generateConfigMap(ctx); err != nil {
		return nil, err
	}
	// Depends on the ConfigMap
	if err := f.generateDeployment(ctx); err != nil {
		return nil, err
	}
	// Depends on the Service
	f.generateService()

	return f, nil
}

func (e *EnvoyFleetResources) CreateOrUpdate(ctx context.Context) error {
	if err := k8sutils.CreateOrReplace(ctx, e.client, e.configMap); err != nil {
		return fmt.Errorf("failed to deploy Envoy Fleet config map: %w", err)
	}
	if err := k8sutils.CreateOrReplace(ctx, e.client, e.deployment); err != nil {
		return fmt.Errorf("failed to  deploy Envoy Fleet deployment: %w", err)
	}
	if err := k8sutils.CreateOrReplace(ctx, e.client, e.service); err != nil {
		return fmt.Errorf("failed to deploy Envoy Fleet service: %w", err)
	}
	return nil
}

func (e *EnvoyFleetResources) generateConfigMap(ctx context.Context) error {
	// future object labels
	labels := map[string]string{
		"app.kubernetes.io/component": "envoy-config",
	}
	// Copy over shared labels map
	for key, value := range e.sharedLabels {
		labels[key] = value
	}

	configMapName := gateway.EnvoyResourceNamePrefix + e.fleet.Name

	xdsLabels := map[string]string{"app.kubernetes.io/name": "kusk-gateway", "app.kubernetes.io/component": "xds-service"}
	xdsServices, err := k8sutils.GetServicesByLabels(ctx, e.client, xdsLabels)
	if err != nil {
		return fmt.Errorf("cannot create Envoy Fleet %s config map: %w", e.fleet.Name, err)
	}
	switch svcs := len(xdsServices); {
	case svcs == 0:
		return fmt.Errorf("cannot create Envoy Fleet %s config map: no xds services detected in the cluster when searching with the labels %s", e.fleet.Name, xdsLabels)
	case svcs > 1:
		return fmt.Errorf("cannot create Envoy Fleet %s config map: multiple xds services detected in the cluster when searching with the labels %s", e.fleet.Name, xdsLabels)
	}
	// At this point - we have exactly one service with (we ASSUME!) one port
	xdsServiceHostname := fmt.Sprintf("%s.%s.svc.cluster.local.", xdsServices[0].Name, xdsServices[0].Namespace)
	xdsServicePort := xdsServices[0].Spec.Ports[0].Port

	e.configMap = &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            configMapName,
			Namespace:       e.fleet.Namespace,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{envoyFleetAsOwner(e.fleet)},
		},
		Data: map[string]string{
			"envoy-config.yaml": fmt.Sprintf(envoyConfigTemplate, e.fleetID, xdsServiceHostname, xdsServicePort),
		},
	}

	return nil
}

func (e *EnvoyFleetResources) generateDeployment(ctx context.Context) error {
	// future object labels
	labels := map[string]string{
		"app.kubernetes.io/component": "envoy",
	}
	// Copy over shared labels map
	for key, value := range e.sharedLabels {
		labels[key] = value
	}

	deploymentName := gateway.EnvoyResourceNamePrefix + e.fleet.Name

	configMapName := e.configMap.Name

	// Create container template first
	envoyContainer := corev1.Container{
		Name:            "envoy",
		Image:           e.fleet.Spec.Image,
		ImagePullPolicy: corev1.PullIfNotPresent,
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
				Name:          "admin",
				ContainerPort: envoyAdminListenerPort,
			},
		},
	}
	// Set Enovy Pod Resources if specified
	if e.fleet.Spec.Resources != nil {
		if e.fleet.Spec.Resources.Limits != nil {
			envoyContainer.Resources.Limits = e.fleet.Spec.Resources.Limits
		}
		if e.fleet.Spec.Resources.Requests != nil {
			envoyContainer.Resources.Requests = e.fleet.Spec.Resources.Requests
		}
	}
	// Creation of the helper sidecar requires passing the parameter with Kusk Gateway Helper management service.
	// We do the service detection dynamically.
	helperServiceLabels := map[string]string{"app.kubernetes.io/name": "kusk-gateway", "app.kubernetes.io/component": "helper-service"}
	helperServices, err := k8sutils.GetServicesByLabels(ctx, e.client, helperServiceLabels)
	if err != nil {
		return fmt.Errorf("cannot create Envoy Fleet %s: %w", e.fleet.Name, err)
	}
	switch svcs := len(helperServices); {
	case svcs == 0:
		return fmt.Errorf("cannot create Envoy Fleet %s: no Helper management services were detected in the cluster when searching with the labels %s", e.fleet.Name, helperServiceLabels)
	case svcs > 1:
		return fmt.Errorf("cannot create Envoy Fleet %s: multiple Helper management services were detected in the cluster when searching with the labels %s", e.fleet.Name, helperServiceLabels)
	}
	// At this point - we have exactly one service with (we ASSUME!) one port
	helperServiceHostname := fmt.Sprintf("%s.%s.svc.cluster.local.", helperServices[0].Name, helperServices[0].Namespace)
	helperServicePort := helperServices[0].Spec.Ports[0].Port

	// Helper container (sidecar)
	helperContainer := corev1.Container{
		Name:            "helper",
		ImagePullPolicy: corev1.PullIfNotPresent,
		// Command:         []string{"/bin/sh", "-c"},
		Args: []string{
			"-fleetID",
			e.fleetID,
			"-helper-config-manager-service-address",
			fmt.Sprintf("%s:%d", helperServiceHostname, helperServicePort),
		},
	}
	// Additional parameters for the helper service
	if e.fleet.Spec.Helper != nil {
		if e.fleet.Spec.Helper.Image != "" {
			helperContainer.Image = e.fleet.Spec.Helper.Image
		}
		if e.fleet.Spec.Helper.Resources != nil {
			if e.fleet.Spec.Helper.Resources.Limits != nil {
				helperContainer.Resources.Limits = e.fleet.Spec.Helper.Resources.Limits
			}
			if e.fleet.Spec.Helper.Resources.Requests != nil {
				helperContainer.Resources.Requests = e.fleet.Spec.Helper.Resources.Requests
			}
		}
	}
	// Image for the helper container was not set, do the autodetection based on Kusk Gateway Manager Image
	if helperContainer.Image == "" {

		kuskGatewayManagerLabels := map[string]string{"app.kubernetes.io/name": "kusk-gateway", "app.kubernetes.io/component": "kusk-gateway-manager"}
		kuskGatewayManagerDeployments, err := k8sutils.GetDeploymentsByLabels(ctx, e.client, kuskGatewayManagerLabels)
		if err != nil {
			return fmt.Errorf("cannot create Envoy Fleet %s: %w", e.fleet.Name, err)
		}
		switch deploys := len(kuskGatewayManagerDeployments); {
		case deploys == 0:
			return fmt.Errorf("cannot create Envoy Fleet %s: no Deployments of Kusk Gateway Manager were found were detected in the cluster when searching with the labels %s, where we're running from?", e.fleet.Name, kuskGatewayManagerLabels)
		case deploys > 1:
			return fmt.Errorf("cannot create Envoy Fleet %s: multiple Deployments of Kusk Gateway Manager were detected in the cluster when searching with the labels %s", e.fleet.Name, kuskGatewayManagerLabels)
		}
		deployment := kuskGatewayManagerDeployments[0]
		for _, container := range deployment.Spec.Template.Spec.Containers {
			// Skip if not the right container
			if container.Name != "manager" {
				continue
			}
			managerImageTag := strings.Split(container.Image, ":")
			if len(managerImageTag) != 2 {
				return fmt.Errorf("cannot create Envoy Fleet %s: failed Kusk Gateway Manager's version autodetection - container image tag %s doesn't match the imageName:version pattern", e.fleet.Name, container.Image)
			}
			containerRepositoryURL := strings.TrimSuffix(managerImageTag[0], kuskGatewayManagerImageName)
			// Form and set helper server container image tag
			helperContainer.Image = fmt.Sprintf("%s%s:%s", containerRepositoryURL, helperImageName, managerImageTag[1])
			break
		}
	}
	// Create the deployment
	e.deployment = &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            deploymentName,
			Namespace:       e.fleet.Namespace,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{envoyFleetAsOwner(e.fleet)},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: e.fleet.Spec.Size,
			Selector: labelSelectors(labels),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: e.fleet.Spec.Annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						envoyContainer,
						helperContainer,
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
					NodeSelector:                  e.fleet.Spec.NodeSelector,
					Affinity:                      e.fleet.Spec.Affinity,
					Tolerations:                   e.fleet.Spec.Tolerations,
					TerminationGracePeriodSeconds: e.fleet.Spec.TerminationGracePeriodSeconds,
				},
			},
		},
	}
	return nil
}

func (e *EnvoyFleetResources) generateService() {
	// future object labels
	labels := map[string]string{
		"app.kubernetes.io/component": "envoy-svc",
	}
	// Copy over shared labels map
	for key, value := range e.sharedLabels {
		labels[key] = value
	}
	serviceName := gateway.EnvoyResourceNamePrefix + e.fleet.Name

	e.service = &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            serviceName,
			Namespace:       e.fleet.Namespace,
			Labels:          labels,
			Annotations:     e.fleet.Spec.Service.Annotations,
			OwnerReferences: []metav1.OwnerReference{envoyFleetAsOwner(e.fleet)},
		},
		Spec: corev1.ServiceSpec{
			Ports:    e.fleet.Spec.Service.Ports,
			Selector: e.deployment.Spec.Selector.MatchLabels,
			Type:     e.fleet.Spec.Service.Type,
		},
	}
	// Static IP address for the LoadBalancer
	if e.fleet.Spec.Service.Type == corev1.ServiceTypeLoadBalancer && e.fleet.Spec.Service.LoadBalancerIP != "" {
		e.service.Spec.LoadBalancerIP = e.fleet.Spec.Service.LoadBalancerIP
	}
	if e.fleet.Spec.Service.ExternalTrafficPolicy != "" {
		e.service.Spec.ExternalTrafficPolicy = e.fleet.Spec.Service.ExternalTrafficPolicy
	}
}

func envoyFleetAsOwner(cr *gateway.EnvoyFleet) metav1.OwnerReference {
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
                address: %s
                port_value: %d

admin:
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 19000

`
