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

package v1

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/kubeshop/kusk-gateway/spec"
)

// log is for logging in this package.
var apilog = logf.Log.WithName("api-resource")

func (r *API) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-gateway-kusk-io-v1-api,mutating=true,failurePolicy=fail,sideEffects=None,groups=gateway.kusk.io,resources=apis,verbs=create;update,versions=v1,name=mapi.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &API{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *API) Default() {
	apilog.Info("default", "name", r.Name)
}

// change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-gateway-kusk-io-v1-api,mutating=false,failurePolicy=fail,sideEffects=None,groups=gateway.kusk.io,resources=apis,verbs=create;update,versions=v1,name=vapi.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &API{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *API) ValidateCreate() error {
	apilog.Info("validate create", "name", r.Name)

	return r.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *API) ValidateUpdate(old runtime.Object) error {
	apilog.Info("validate update", "name", r.Name)

	return r.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *API) ValidateDelete() error {
	apilog.Info("validate delete", "name", r.Name)
	return nil
}

func (r *API) validate() error {
	parser := spec.NewParser(nil)

	apiSpec, err := parser.ParseFromReader(strings.NewReader(r.Spec.Spec))
	if err != nil {
		return fmt.Errorf("spec: should be a valid OpenAPI spec: %w", err)
	}

	opts, err := spec.GetOptions(apiSpec)
	if err != nil {
		return fmt.Errorf("spec: x-kusk should be a valid set of options: %w", err)
	}

	err = opts.FillDefaultsAndValidate()
	if err != nil {
		return fmt.Errorf("spec: x-kusk should be a valid set of options: %w", err)
	}

	return nil
}
