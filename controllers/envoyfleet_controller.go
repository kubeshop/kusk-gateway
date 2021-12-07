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

package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	gatewayv1alpha1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/k8sutils"
)

// EnvoyFleetReconciler reconciles a EnvoyFleet object
type EnvoyFleetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=gateway.kusk.io,resources=envoyfleet,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gateway.kusk.io,resources=envoyfleet/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=gateway.kusk.io,resources=envoyfleet/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *EnvoyFleetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := ctrl.LoggerFrom(ctx)
	l.Info("EnvoyFleet changed", "changed", req.NamespacedName)

	ef := &gatewayv1alpha1.EnvoyFleet{}

	err := r.Client.Get(ctx, req.NamespacedName, ef)
	if err != nil {
		if errors.IsNotFound(err) {
			// EnvoyFleet was deleted - deployment and config deletion is handled by the API server itself
			// thanks to OwnerReference
			l.Info("No objects found, looks like EnvoyFleet was deleted")
			return ctrl.Result{}, nil
		}
		l.Error(err, "Failed to retrieve EnvoyFleet object with cluster API")
		return ctrl.Result{Requeue: true}, err
	}

	if err := controllerutil.SetControllerReference(ef, ef, r.Scheme); err != nil {
		l.Error(err, "Failed setting controller owner reference")
		return ctrl.Result{}, err
	}
	efResources, err := NewEnvoyFleetResources(ef)
	if err != nil {
		l.Error(err, "Failed to create envoy fleet configuration")
		return ctrl.Result{}, fmt.Errorf("failed to create envoy fleet configuration: %w", err)
	}

	if err := k8sutils.CreateOrReplace(ctx, r.Client, efResources.configMap); err != nil {
		l.Error(err, "Failed to create envoy envoy config map")
		return ctrl.Result{}, fmt.Errorf("failed to create envoy fleet config map: %w", err)
	}
	if err := k8sutils.CreateOrReplace(ctx, r.Client, efResources.deployment); err != nil {
		l.Error(err, "Failed to create envoy deployment")
		return ctrl.Result{}, fmt.Errorf("failed to create envoy fleet deployment: %w", err)
	}
	if err := k8sutils.CreateOrReplace(ctx, r.Client, efResources.service); err != nil {
		l.Error(err, "Failed to create envoy fleet service")
		return ctrl.Result{}, fmt.Errorf("failed to create envoy fleetconfig map: %w", err)
	}
	l.Info(fmt.Sprintf("Reconciled Envoy Fleet '%s' resources", efResources.fleetName))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EnvoyFleetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1alpha1.EnvoyFleet{}).
		Complete(r)
}
