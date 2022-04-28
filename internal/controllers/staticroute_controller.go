/*
MIT License

Copyright (c) 2022 Kubeshop

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

package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gateway "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/pkg/analytics"
)

const (
	StaticRouteFinalizer = "gateway.kusk.io/srfinalizer"
)

// StaticRouteReconciler reconciles a StaticRoute object
type StaticRouteReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	ConfigManager *KubeEnvoyConfigManager
}

//+kubebuilder:rbac:groups=gateway.kusk.io,resources=staticroutes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gateway.kusk.io,resources=staticroutes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gateway.kusk.io,resources=staticroutes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *StaticRouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithName("static-route-controller")
	analytics.SendAnonymousInfo(fmt.Sprintf("reconciling Static route  %s", req.Name))

	l.Info("Reconciling changed StaticRoute resource", "changed", req.NamespacedName)
	defer l.Info("Finished reconciling changed StaticRoute resource", "changed", req.NamespacedName)

	var srObj gateway.StaticRoute
	// In order to get fleet ID we MUST find the object.
	// If it is missing, that means it was deleted without the finalizer, we don't do anything.
	// If it is in the state of deletion - we get the object and remove the finalizer to allow K8s to finally delete it.
	// If it is present and without the finalizer - we add it.
	if err := r.Client.Get(ctx, req.NamespacedName, &srObj); err != nil {
		// Object not found, log the error but do not retry (not returning the error to the caller)
		if client.IgnoreNotFound(err) == nil {
			l.Error(err, fmt.Sprintf("the StaticRoute object %s.%s was not found, it was likely deleted previously, skipping the processing", req.Name, req.Namespace))
			return ctrl.Result{}, nil
		}
		// Other errors, fail with retry
		l.Error(err, fmt.Sprintf("Failed to reconcile StaticRoute %s, will retry in %d seconds", req.NamespacedName, reconcilerFastRetrySeconds))
		return ctrl.Result{RequeueAfter: time.Duration(time.Second * time.Duration(reconcilerFastRetrySeconds))}, err
	}
	// Handle finalisers
	if srObj.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(srObj.GetFinalizers(), StaticRouteFinalizer) {
			controllerutil.AddFinalizer(&srObj, StaticRouteFinalizer)
			if err := r.Update(ctx, &srObj); err != nil {
				l.Error(err, fmt.Sprintf("Failed to reconcile StaticRoute %s, will retry in %d seconds", req.NamespacedName, reconcilerFastRetrySeconds))
				return ctrl.Result{RequeueAfter: time.Duration(time.Second * time.Duration(reconcilerFastRetrySeconds))}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(srObj.GetFinalizers(), StaticRouteFinalizer) {
			// our finalizer is present
			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(&srObj, StaticRouteFinalizer)
			if err := r.Update(ctx, &srObj); err != nil {
				l.Error(err, fmt.Sprintf("Failed to reconcile StaticRoute %s during finalizer remove, will retry in %d seconds", req.NamespacedName, reconcilerFastRetrySeconds))
				return ctrl.Result{RequeueAfter: time.Duration(time.Second * time.Duration(reconcilerFastRetrySeconds))}, err
			}
		}
	}

	if srObj.Spec.Fleet == nil {
		err := fmt.Errorf("StaticRoute object %s.%s - fleet field is empty", srObj.Name, srObj.Namespace)
		l.Error(err, "Failed to reconcile StaticRoute", "changed", req.NamespacedName)
		return ctrl.Result{}, err
	}
	// Finally call ConfigManager to update the configuration with this fleet ID
	if err := r.ConfigManager.UpdateConfiguration(ctx, *srObj.Spec.Fleet); err != nil {
		l.Error(err, fmt.Sprintf("Failed to reconcile StaticRoute %s, will retry in %d seconds", req.NamespacedName, reconcilerFastRetrySeconds))
		return ctrl.Result{RequeueAfter: time.Duration(time.Second * time.Duration(reconcilerFastRetrySeconds))}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StaticRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Setup client caching index by StaticRoute objects spec.Fleet field
	if err := mgr.GetFieldIndexer().IndexField(
		context.TODO(),
		&gateway.StaticRoute{},
		"spec.fleet",
		func(rawObj client.Object) []string {
			api := rawObj.(*gateway.StaticRoute)
			return []string{api.Spec.Fleet.String()}
		},
	); err != nil {
		return fmt.Errorf("unable to add StaticRoute field indexer to the cache: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&gateway.StaticRoute{}).
		Complete(r)
}
