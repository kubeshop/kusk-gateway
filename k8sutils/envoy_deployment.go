package k8sutils

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1 "github.com/kubeshop/kusk-gateway/api/v1"
)

func CreateEnvoyDeployment(ctx context.Context, client client.Client, cr *gatewayv1.EnvoyFleet) error {
	labels := map[string]string{
		"app":       "kusk-gateway",
		"component": "envoy",
		"fleet":     cr.Name,
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            cr.Name,
			Namespace:       cr.Namespace,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{envoyFleetAsOwner(cr)},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: cr.Spec.Size,
			Selector: labelSelectors(labels),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "envoy",
							Image:   "envoyproxy/envoy-dev:latest",
							Command: []string{"envoy"},
							Args:    []string{"-c /etc/envoy/envoy.yaml"},
							// VolumeMounts: []corev1.VolumeMount{
							// 	{
							// 		Name:      "envoy-config",
							// 		MountPath: "/etc/envoy/envoy.yaml",
							// 		SubPath:   "envoy-config.yaml",
							// 	},
							// },
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "envoy-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: nil, // TODO:
							},
						},
					},
				},
			},
		},
	}

	return client.Create(ctx, deployment)
}

func envoyFleetAsOwner(cr *gatewayv1.EnvoyFleet) metav1.OwnerReference {
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
