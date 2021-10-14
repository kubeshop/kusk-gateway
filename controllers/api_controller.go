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
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gateway "github.com/kubeshop/kusk-gateway/api/v1"
	"github.com/kubeshop/kusk-gateway/envoy/config"
	"github.com/kubeshop/kusk-gateway/envoy/manager"
	"github.com/kubeshop/kusk-gateway/spec"
)

// APIReconciler reconciles a API object
type APIReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	EnvoyManager *manager.EnvoyConfigManager

	m sync.Mutex
}

//+kubebuilder:rbac:groups=gateway.kusk.io,resources=apis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gateway.kusk.io,resources=apis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gateway.kusk.io,resources=apis/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the API object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *APIReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	parser := spec.NewParser(nil)

	// acquiring this lock is required so that no potentially conflicting updates would happen at the same time
	// this probably should be done on a per-envoy basis but as we have a static config for now this will do
	r.m.Lock()
	defer r.m.Unlock()

	// fetch all APIs to rebuild Envoy configuration
	// we have a static envoy configuration so we don't need any filters here
	var apis gateway.APIList
	if err := r.Client.List(ctx, &apis); err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	envoyConfig := config.New()

	for _, api := range apis.Items {
		apiSpec, err := parser.ParseFromReader(strings.NewReader(api.Spec.Spec))
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
		}

		opts, err := spec.GetOptions(apiSpec)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to parse options: %w", err)
		}

		err = opts.FillDefaultsAndValidate()
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to validate options: %w", err)
		}

		_, err = envoyConfig.GenerateConfigSnapshotFromOpts(opts, apiSpec)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to generate config: %w", err)
		}
	}

	snapshot, err := envoyConfig.GenerateSnapshot()
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to generate snapshot: %w", err)
	}

	if err := r.EnvoyManager.ApplyNewFleetSnapshot(manager.DefaultFleetName, snapshot); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to apply snapshot: %w", err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gateway.API{}).
		Complete(r)
}
