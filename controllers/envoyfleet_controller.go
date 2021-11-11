/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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

	gatewayv1 "github.com/kubeshop/kusk-gateway/api/v1"
	"github.com/kubeshop/kusk-gateway/k8sutils"
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

	ef := &gatewayv1.EnvoyFleet{}

	err := r.Client.Get(context.TODO(), req.NamespacedName, ef)
	if err != nil {
		if errors.IsNotFound(err) {
			// EnvoyFleet was deleted - deployment and config deletion is handled by the API server itself
			// thanks to OwnerReference
			return ctrl.Result{}, nil
		}

		return ctrl.Result{Requeue: true}, err
	}

	if err := controllerutil.SetControllerReference(ef, ef, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	if err := k8sutils.CreateEnvoyConfig(ctx, r.Client, ef); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create envoy config: %w", err)
	}

	if err := k8sutils.CreateEnvoyDeployment(ctx, r.Client, ef); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create envoy deployment: %w", err)
	}

	if err := k8sutils.CreateEnvoyService(ctx, r.Client, ef); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create envoy service: %w", err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EnvoyFleetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1.EnvoyFleet{}).
		Complete(r)
}
