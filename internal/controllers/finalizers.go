package controllers

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type resourceUpdater interface {
	Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error
}

func handleFinalizers(ctx context.Context, updater resourceUpdater, obj client.Object, finalizer string) error {
	// The object is not being deleted, so if it does not have our finalizer,
	// then lets add the finalizer and update the object. This is equivalent
	// registering our finalizer.
	if obj.GetDeletionTimestamp().IsZero() && !containsString(obj.GetFinalizers(), finalizer) {
		controllerutil.AddFinalizer(obj, finalizer)
		if err := updater.Update(ctx, obj); err != nil {
			return err
		}
		return nil
	}

	// The object is being deleted
	if !obj.GetDeletionTimestamp().IsZero() && containsString(obj.GetFinalizers(), finalizer) {
		// our finalizer is present
		// remove our finalizer from the list and update it.
		controllerutil.RemoveFinalizer(obj, finalizer)
		if err := updater.Update(ctx, obj); err != nil {
			return err
		}
	}

	return nil
}
