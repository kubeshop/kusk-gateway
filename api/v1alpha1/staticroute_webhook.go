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

package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var srlog = logf.Log.WithName("staticroute-resource")

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-gateway-kusk-io-v1alpha1-staticroute,mutating=true,failurePolicy=fail,sideEffects=None,groups=gateway.kusk.io,resources=staticroutes,verbs=create;update,versions=v1alpha1,name=mstaticroute.kb.io,admissionReviewVersions=v1

const (
	StaticRouteMutatingWebhookPath   string = "/mutate-gateway-kusk-io-v1alpha1-staticroute"
	StaticRouteValidatingWebhookPath string = "/validate-gateway-kusk-io-v1alpha1-staticroute"
)

// StaticRouteMutator handles StaticRoute objects defaulting and any additional mutation.
// +kubebuilder:object:generate:=false
type StaticRouteMutator struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (s *StaticRouteMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	srObj := &StaticRoute{}

	err := s.decoder.Decode(req, srObj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if len(srObj.Spec.Hosts) == 0 {
		srlog.Info("missing Hosts entry, setting Hosts to default *")
		srObj.Spec.Hosts = append(srObj.Spec.Hosts, "*")
	}
	// If the spec.fleet is not set, find the deployed Envoy Fleet in the cluster and update with it.
	// If there are multiple fleets in the cluster, make user update the the resource spec.fleet with the desired fleet.
	if srObj.Spec.Fleet == nil {
		srlog.Info("spec.fleet is not defined in the StaticRoute resource, defaulting it to the present in the cluster Envoy Fleet")

		var fleets EnvoyFleetList
		if err := s.Client.List(ctx, &fleets); err != nil {
			srlog.Error(err, "Failed to get the deployed Envoy Fleets")
			return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to get the deployed Envoy Fleets: %w", err))
		}
		switch l := len(fleets.Items); {
		case l == 0:
			srlog.Error(err, "cannot update StaticRoute spec.fleet to the default fleet in the cluster - we found no deployed Envoy Fleets")
			return admission.Errored(http.StatusConflict, fmt.Errorf("StaticRoute spec.fleet is not set and there is no deployed Envoy Fleets in the cluster to set as the default, deploy at least one to the cluster before trying to submit the StaticRoute resource."))
		case l > 1:
			srlog.Error(err, "cannot update StaticRoute spec.fleet to the default fleet in the cluster - found more than one deployed Envoy Fleets")
			return admission.Errored(http.StatusConflict, fmt.Errorf("StaticRoute spec.fleet is not set and there are multiple deployed Envoy Fleets, set spec.fleet to the desired one."))
		default:
			fl := fleets.Items[0]
			srlog.Info("StaticRoute spec.fleet is not set, defaulting to the deployed %s.%s Envoy Fleet in the cluster", fl.Name, fl.Namespace)
			srObj.Spec.Fleet = &EnvoyFleetID{Name: fl.Name, Namespace: fl.Namespace}
		}
	}

	marshaledObj, err := json.Marshal(srObj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledObj)
}

// StaticRouteMutator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (s *StaticRouteMutator) InjectDecoder(d *admission.Decoder) error {
	s.decoder = d
	return nil
}

// change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-gateway-kusk-io-v1alpha1-staticroute,mutating=false,failurePolicy=fail,sideEffects=None,groups=gateway.kusk.io,resources=staticroutes,verbs=create;update,versions=v1alpha1,name=vstaticroute.kb.io,admissionReviewVersions=v1

// StaticRouteValidator handles StaticRoute objects validation
// +kubebuilder:object:generate:=false
type StaticRouteValidator struct {
	decoder *admission.Decoder
}

func (s *StaticRouteValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	srObj := &StaticRoute{}

	err := s.decoder.Decode(req, srObj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	spec, err := srObj.Spec.GetOptionsFromSpec()
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if err := spec.Validate(); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.Allowed("")
}

// StaticRouteValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (s *StaticRouteValidator) InjectDecoder(d *admission.Decoder) error {
	s.decoder = d
	return nil
}
