package k8sutils

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientPkg "sigs.k8s.io/controller-runtime/pkg/client"
)

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

func CreateOrReplace(ctx context.Context, client clientPkg.Client, obj clientPkg.Object) error {
	resourceVersion, ok, err := checkIfExists(ctx, client, obj.GetObjectKind().GroupVersionKind(), clientPkg.ObjectKeyFromObject(obj))
	if err != nil {
		return err
	}

	if ok {
		obj.SetResourceVersion(resourceVersion)
		return client.Update(ctx, obj)
	}

	return client.Create(ctx, obj)
}

func GetServicesByLabels(ctx context.Context, client clientPkg.Client, labels map[string]string) ([]corev1.Service, error) {
	labelSelector := clientPkg.MatchingLabels(labels)

	servicesList := &corev1.ServiceList{}
	if err := client.List(ctx, servicesList, labelSelector); err != nil {
		return []corev1.Service{}, fmt.Errorf("failed getting services from the cluster: %w", err)
	}

	return servicesList.Items, nil
}

func GetDeploymentsByLabels(ctx context.Context, client clientPkg.Client, labels map[string]string) ([]appsv1.Deployment, error) {
	labelSelector := clientPkg.MatchingLabels(labels)

	deployList := &appsv1.DeploymentList{}
	if err := client.List(ctx, deployList, labelSelector); err != nil {
		return []appsv1.Deployment{}, fmt.Errorf("failed getting deployments from the cluster: %w", err)
	}

	return deployList.Items, nil
}
