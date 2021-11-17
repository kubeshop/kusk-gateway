package k8sutils

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientPkg "sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1alpha1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
)

func CreateEnvoyConfig(ctx context.Context, client clientPkg.Client, ef *gatewayv1alpha1.EnvoyFleet) error {
	labels := map[string]string{
		"app":       "kusk-gateway",
		"component": "envoy-config",
		"fleet":     ef.Name,
	}

	configMapName := "kusk-envoy-config-" + ef.Name

	configMap := &corev1.ConfigMap{
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
			"envoy-config.yaml": fmt.Sprintf(envoyConfigTemplate, ef.Name),
		},
	}

	return createOrReplace(ctx, client, configMap.GroupVersionKind(), configMap)
}

func CreateEnvoyService(ctx context.Context, client clientPkg.Client, ef *gatewayv1alpha1.EnvoyFleet) error {
	labels := map[string]string{
		"app":       "kusk-gateway",
		"component": "envoy-svc",
		"fleet":     ef.Name,
	}

	envoyLabels := map[string]string{
		"app":       "kusk-gateway",
		"component": "envoy",
		"fleet":     ef.Name,
	}

	serviceName := "kusk-envoy-svc-" + ef.Name

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            serviceName,
			Namespace:       ef.Namespace,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{envoyFleetAsOwner(ef)},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
			},
			Selector: envoyLabels,
			Type:     corev1.ServiceTypeLoadBalancer,
		},
	}

	return createOrReplace(ctx, client, service.GroupVersionKind(), service)
}

func CreateEnvoyDeployment(ctx context.Context, client clientPkg.Client, ef *gatewayv1alpha1.EnvoyFleet) error {
	labels := map[string]string{
		"app":       "kusk-gateway",
		"component": "envoy",
		"fleet":     ef.Name,
	}

	deploymentName := "kusk-envoy-" + ef.Name
	configMapName := "kusk-envoy-config-" + ef.Name

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            deploymentName,
			Namespace:       ef.Namespace,
			Labels:          labels,
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
						{
							Name:    "envoy",
							Image:   "envoyproxy/envoy-alpine:v1.20.0",
							Command: []string{"/bin/sh", "-c"},
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
						},
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
				},
			},
		},
	}

	return createOrReplace(ctx, client, deployment.GroupVersionKind(), deployment)
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

func checkIfExists(ctx context.Context, client clientPkg.Client, gvk schema.GroupVersionKind, key clientPkg.ObjectKey) (resourceVersion string, ok bool, err error) {
	var obj unstructured.Unstructured

	obj.SetGroupVersionKind(gvk)

	err = client.Get(ctx, key, &obj)
	if err != nil {
		if errors.IsNotFound(err) {
			return "", false, nil
		}

		return "", false, err
	}

	return obj.GetResourceVersion(), true, nil
}

func createOrReplace(ctx context.Context, client clientPkg.Client, gvk schema.GroupVersionKind, obj clientPkg.Object) error {
	resourceVersion, ok, err := checkIfExists(ctx, client, gvk, clientPkg.ObjectKeyFromObject(obj))
	if err != nil {
		return err
	}

	if ok {
		obj.SetResourceVersion(resourceVersion)
		return client.Update(ctx, obj)
	}

	return client.Create(ctx, obj)
}
