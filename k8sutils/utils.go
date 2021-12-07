package k8sutils

import (
	"context"

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
