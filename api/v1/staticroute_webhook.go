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

package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var staticroutelog = logf.Log.WithName("staticroute-resource")

func (r *StaticRoute) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-gateway-kusk-io-v1-staticroute,mutating=true,failurePolicy=fail,sideEffects=None,groups=gateway.kusk.io,resources=staticroutes,verbs=create;update,versions=v1,name=mstaticroute.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &StaticRoute{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *StaticRoute) Default() {
	staticroutelog.Info("default", "name", r.Name)
	if len(r.Spec.Hosts) == 0 {
		staticroutelog.Info("missing Hosts entry, setting Hosts to default *")
		r.Spec.Hosts = append(r.Spec.Hosts, "*")
	}
	staticroutelog.Info("spec", "r", r.Spec)
}

//+kubebuilder:webhook:path=/validate-gateway-kusk-io-v1-staticroute,mutating=false,failurePolicy=fail,sideEffects=None,groups=gateway.kusk.io,resources=staticroutes,verbs=create;update,versions=v1,name=vstaticroute.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &StaticRoute{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *StaticRoute) ValidateCreate() error {
	staticroutelog.Info("validate create", "name", r.Name)
	// This is a simple validation of options generation from spec
	// TODO: we need to ensure that this spec can be merged with other configurations.
	spec, _ := r.Spec.GetOptionsFromSpec()
	return spec.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *StaticRoute) ValidateUpdate(old runtime.Object) error {
	staticroutelog.Info("validate update", "name", r.Name)

	// TODO: we need to ensure that this spec can be merged with other configurations.
	spec, _ := r.Spec.GetOptionsFromSpec()
	return spec.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *StaticRoute) ValidateDelete() error {
	staticroutelog.Info("validate delete", "name", r.Name)
	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
