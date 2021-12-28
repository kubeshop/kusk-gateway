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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	gatewayv1alpha1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
)

const (
	// Used to set the State field in the Status
	envoyFleetStateSuccess string = "Deployed"
	envoyFleetStateFailure string = "Failed"
)

// EnvoyFleetReconciler reconciles a EnvoyFleet object
type EnvoyFleetReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	ConfigManager *KubeEnvoyConfigManager
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
	// Generate Envoy Fleet resources...
	efResources, err := NewEnvoyFleetResources(ctx, r.Client, ef)
	if err != nil {
		l.Error(err, "Failed to create EnvoyFleet configuration")
		ef.Status.State = envoyFleetStateFailure
		if err := r.Client.Status().Update(ctx, ef); err != nil {
			l.Error(err, "Unable to update Envoy Fleet status")
		}
		return ctrl.Result{}, fmt.Errorf("failed to create EnvoyFleet configuration: %w", err)
	}
	// and deploy them
	if err = efResources.CreateOrUpdate(ctx); err != nil {
		l.Error(err, fmt.Sprintf("Failed to reconcile EnvoyFleet, will retry in %d seconds", reconcilerDefaultRetrySeconds))
		ef.Status.State = envoyFleetStateFailure
		if err := r.Client.Status().Update(ctx, ef); err != nil {
			l.Error(err, "Unable to update Envoy Fleet status")
		}

		return ctrl.Result{RequeueAfter: time.Duration(time.Duration(reconcilerDefaultRetrySeconds) * time.Second)},
			fmt.Errorf("failed to create or update EnvoyFleet: %w", err)
	}
	// Call Envoy configuration manager to update Envoy fleet configuration when applicable
	// This could be extended for any field that belongs to EnvoyFleet CRD but is used to configure Envoy proxy.
	if efResources.fleet.Spec.AccessLog != nil {
		l.Info("Calling Config Manager due to change in Envoy Fleet resource", "changed", req.NamespacedName)
		if err := r.ConfigManager.UpdateConfiguration(ctx, gatewayv1alpha1.EnvoyFleetID{Name: req.Name, Namespace: req.Namespace}); err != nil {
			ef.Status.State = envoyFleetStateFailure
			if err := r.Client.Status().Update(ctx, ef); err != nil {
				l.Error(err, "Unable to update Envoy Fleet status")
			}
			l.Error(err, fmt.Sprintf("Failed to reconcile Envoy Fleet, will retry in %d seconds", reconcilerDefaultRetrySeconds))
			return ctrl.Result{RequeueAfter: time.Duration(time.Duration(reconcilerDefaultRetrySeconds) * time.Second)},
				fmt.Errorf("failed to update Envoy Fleet %s configuration: %w", ef.Name, err)
		}
	}
	l.Info(fmt.Sprintf("Reconciled EnvoyFleet '%s' resources", ef.Name))
	ef.Status.State = envoyFleetStateSuccess
	if err := r.Client.Status().Update(ctx, ef); err != nil {
		l.Error(err, "Unable to update Envoy Fleet status")
		return ctrl.Result{RequeueAfter: time.Duration(time.Duration(reconcilerDefaultRetrySeconds) * time.Second)}, fmt.Errorf("unable to update Envoy Fleet status")
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EnvoyFleetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// predicate will prevent triggering the Reconciler on resource Status field changes.
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1alpha1.EnvoyFleet{}).
		WithEventFilter(pred).
		Complete(r)
}
