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

package v1alpha1

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

//+kubebuilder:webhook:path=/mutate-gateway-kusk-io-v1alpha1-api,mutating=true,failurePolicy=fail,sideEffects=None,groups=gateway.kusk.io,resources=apis,verbs=create;update,versions=v1alpha1,name=mapi.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &API{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *API) Default() {
	apilog.Info("default", "name", r.Name)
}

// change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-gateway-kusk-io-v1alpha1-api,mutating=false,failurePolicy=fail,sideEffects=None,groups=gateway.kusk.io,resources=apis,verbs=create;update,versions=v1alpha1,name=vapi.kb.io,admissionReviewVersions={v1,v1beta1}

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
	if len(apiSpec.Paths) == 0 {
		return fmt.Errorf("spec: should be a valid OpenAPI spec, no paths found")
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
