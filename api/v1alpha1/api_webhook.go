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
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/kubeshop/kusk-gateway/pkg/spec"
)

// log is for logging in this package.
var apilog = logf.Log.WithName("api-resource")

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-gateway-kusk-io-v1alpha1-api,mutating=true,failurePolicy=fail,sideEffects=None,groups=gateway.kusk.io,resources=apis,verbs=create;update,versions=v1alpha1,name=mapi.kb.io,admissionReviewVersions={v1,v1beta1}

const (
	APIMutatingWebhookPath   string = "/mutate-gateway-kusk-io-v1alpha1-api"
	APIValidatingWebhookPath string = "/validate-gateway-kusk-io-v1alpha1-api"
)

// APIMutator handles API objects defaulting and any additional mutation.
//+kubebuilder:object:generate:=false
type APIMutator struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (a *APIMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	apiObj := &API{}

	err := a.decoder.Decode(req, apiObj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	// If the spec.fleet is not set, find the deployed Envoy Fleet in the cluster and update with it.
	// If there are multiple fleets in the cluster, make user update the resource spec.fleet with the desired fleet.
	if apiObj.Spec.Fleet == nil {
		apilog.Info("spec.fleet is not defined in the API resource, defaulting it to the present in the cluster Envoy Fleet")

		var fleets EnvoyFleetList
		if err := a.Client.List(ctx, &fleets); err != nil {
			apilog.Error(err, "Failed to get the deployed Envoy Fleets")
			return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to get the deployed Envoy Fleets: %w", err))
		}
		switch l := len(fleets.Items); {
		case l == 0:
			apilog.Error(err, "cannot update API spec.fleet to the default fleet in the cluster - we found no deployed Envoy Fleets")
			return admission.Errored(http.StatusConflict, fmt.Errorf("API spec.fleet is not set and there is no deployed Envoy Fleets in the cluster to set as the default, deploy at least one to the cluster before trying to submit the API resource"))
		default:
			var defaultFleet *EnvoyFleet
			for _, fleet := range fleets.Items {
				if fleet.Spec.Default {
					defaultFleet = &fleet
					break
				}
			}

			if defaultFleet == nil {
				apilog.Error(err, "cannot update API spec.fleet to the default fleet in the cluster - there is no default envoyfleet")
				return admission.Errored(http.StatusConflict, fmt.Errorf("cannot update API spec.fleet to the default fleet in the cluster - there is no default envoyfleet"))
			}

			apilog.Info("API spec.fleet is not set, defaulting to the deployed %s.%s Envoy Fleet in the cluster", defaultFleet.Name, defaultFleet.Namespace)
			apiObj.Spec.Fleet = &EnvoyFleetID{Name: defaultFleet.Name, Namespace: defaultFleet.Namespace}
		}
	}

	marshaledObj, err := json.Marshal(apiObj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledObj)
}

// APIMutator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (a *APIMutator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

// change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-gateway-kusk-io-v1alpha1-api,mutating=false,failurePolicy=fail,sideEffects=None,groups=gateway.kusk.io,resources=apis,verbs=create;update,versions=v1alpha1,name=vapi.kb.io,admissionReviewVersions={v1,v1beta1}

// APIValidator handles API objects validation
//+kubebuilder:object:generate:=false
type APIValidator struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (a *APIValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	apiObj := &API{}

	err := a.decoder.Decode(req, apiObj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := apiObj.validate(); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := a.PathAlreadyDeployed(ctx, apiObj.Spec.Fleet, *apiObj); err != nil {
		return admission.Errored(http.StatusConflict, err)
	}
	return admission.Allowed("")
}

// APIValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (a *APIValidator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

func (a *APIValidator) PathAlreadyDeployed(ctx context.Context, fleet *EnvoyFleetID, newApi API) error {
	var apiObjs APIList
	parser := spec.NewParser(nil)
	newApiSpec, err := parser.ParseFromReader(strings.NewReader(newApi.Spec.Spec))
	if err != nil {
		return err
	}

	// Get all API objects with this fleet field set
	if err := a.Client.List(ctx, &apiObjs, &client.ListOptions{}); err != nil {
		return fmt.Errorf("failure querying for the deployed APIs: %w", err)
	}

	// filter out apis are in the process of deletion
	for _, api := range apiObjs.Items {
		if api.Spec.Fleet.Name == fleet.Name && api.Spec.Fleet.Namespace == fleet.Namespace {
			if api.ObjectMeta.DeletionTimestamp.IsZero() {
				apiSpec, err := parser.ParseFromReader(strings.NewReader(api.Spec.Spec))
				if err != nil {
					return err
				}

				for path := range newApiSpec.Paths {
					if _, ok := apiSpec.Paths[path]; ok {
						return fmt.Errorf("path %s already exists with envoyfleet %s", path, fleet)
					}
				}
			}
		}
	}
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
	opts.FillDefaults()
	if err = opts.Validate(); err != nil {
		return fmt.Errorf("spec: x-kusk should be a valid set of options: %w", err)
	}

	return nil
}
