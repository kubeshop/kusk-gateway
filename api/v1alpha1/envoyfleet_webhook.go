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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/kubeshop/kusk-gateway/internal/cert"
)

// log is for logging in this package.
var (
	envoyfleetlog = logf.Log.WithName("envoyfleet-resource")
)

const (
	EnvoyFleetMutatingWebhookPath   string = "/mutate-gateway-kusk-io-v1alpha1-envoyfleet"
	EnvoyFleetValidatingWebhookPath        = "/validate-gateway-kusk-io-v1alpha1-envoyfleet"
)

//+kubebuilder:webhook:path=/mutate-gateway-kusk-io-v1alpha1-envoyfleet,mutating=true,failurePolicy=fail,sideEffects=None,groups=gateway.kusk.io,resources=envoyfleet,verbs=create;update,versions=v1alpha1,name=menvoyfleet.kb.io,admissionReviewVersions=v1

// EnvoyFleetMutator handles EnvoyFleet objects defaulting and any additional mutation
//+kubebuilder:object:generate:=false
type EnvoyFleetMutator struct {
	decoder *admission.Decoder
}

func (e *EnvoyFleetMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	efObject := &EnvoyFleet{}

	if err := e.decoder.Decode(req, efObject); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if efObject.Spec.TLS.TlsSecrets == nil {
		efObject.Spec.TLS.TlsSecrets = []TLSSecrets{}
	}

	marshaledObj, err := json.Marshal(efObject)
	if err != nil {
		admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledObj)
}

// EnvoyFleetMutator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (e *EnvoyFleetMutator) InjectDecoder(d *admission.Decoder) error {
	e.decoder = d
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-gateway-kusk-io-v1alpha1-envoyfleet,mutating=false,failurePolicy=fail,sideEffects=None,groups=gateway.kusk.io,resources=envoyfleet,verbs=create;update,versions=v1alpha1,name=venvoyfleet.kb.io,admissionReviewVersions=v1

// EnvoyFleetValidator handles EnvoyFleet objects validation
//+kubebuilder:object:generate:=false
type EnvoyFleetValidator struct {
	Client  client.Client
	decoder *admission.Decoder
}

// EnvoyFleetValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (e *EnvoyFleetValidator) InjectDecoder(d *admission.Decoder) error {
	e.decoder = d
	return nil
}

func (e *EnvoyFleetValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	envoyFleetObj := &EnvoyFleet{}

	err := e.decoder.Decode(req, envoyFleetObj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	return e.validate(ctx, envoyFleetObj)
}

func (e *EnvoyFleetValidator) validate(ctx context.Context, envoyFleet *EnvoyFleet) admission.Response {

	if resp := e.validateNameWithinSizeBound(envoyFleet.Name); !resp.Allowed {
		return resp
	}

	if resp := e.validateNoOverlappingSANSInTLS(ctx, envoyFleet.Spec.TLS.TlsSecrets); !resp.Allowed {
		return resp
	}

	return admission.Allowed("")
}

func (e *EnvoyFleetValidator) validateNameWithinSizeBound(name string) admission.Response {
	if kubernetesMaxNameLength := 64; len(name) > kubernetesMaxNameLength {
		err := fmt.Errorf(
			"resulting name of envoy resources (%s) is larger than the kubernetes max allowed name of %d",
			name,
			kubernetesMaxNameLength,
		)
		return admission.Errored(http.StatusBadRequest, err)
	}

	return admission.Allowed("")
}

func (e *EnvoyFleetValidator) validateNoOverlappingSANSInTLS(ctx context.Context, secrets []TLSSecrets) admission.Response {
	getSecret := func(tlsSecret TLSSecrets) (*v1.Secret, error) {
		var secret v1.Secret
		if err := e.Client.Get(
			ctx,
			types.NamespacedName{
				Name:      tlsSecret.SecretRef,
				Namespace: tlsSecret.Namespace,
			},
			&secret,
		); err != nil {
			return nil, err
		}

		return &secret, nil
	}

	getDNSNamesFromCert := func(crt []byte) ([]string, error) {
		certChain, err := cert.DecodeCertificates(crt)
		if err != nil {
			return nil, fmt.Errorf("unable to decode certificates: %w", err)
		}

		if len(certChain) == 0 {
			return nil, fmt.Errorf("resulting cert chain length was 0")
		}

		leafCert := certChain[0]
		if len(leafCert.DNSNames) == 0 {
			return nil, fmt.Errorf("found certificate without SAN. All provided certificates must have at least one SAN")
		}

		return leafCert.DNSNames, nil
	}

	// map of sans pointing to the secret they already associate with
	sanSet := map[string]*v1.Secret{}

	for _, tlsSecret := range secrets {
		envoyfleetlog.Info("processing secret", "secret", tlsSecret.SecretRef, "namespace", tlsSecret.Namespace)
		secret, err := getSecret(tlsSecret)
		if err != nil {
			wrappedErr := fmt.Errorf(
				"unable to query secret %s in namespace %s: %w",
				tlsSecret.SecretRef,
				tlsSecret.Namespace,
				err,
			)

			var statusCode int32 = http.StatusInternalServerError
			if errors.IsNotFound(err) {
				statusCode = http.StatusNotFound
			} else if errors.IsBadRequest(err) {
				statusCode = http.StatusBadRequest
			}

			return admission.Errored(statusCode, wrappedErr)
		}

		crt, ok := secret.Data["tls.crt"]
		if !ok {
			return admission.Errored(
				http.StatusBadRequest,
				fmt.Errorf("tls.crt field not found in secret %s in namespace %s", secret.Name, secret.Namespace),
			)
		}

		dnsNames, err := getDNSNamesFromCert(crt)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		for _, dnsName := range dnsNames {
			if associatedSecret, ok := sanSet[dnsName]; ok {
				return admission.Errored(
					http.StatusBadRequest,
					fmt.Errorf(
						"conflicting SAN %s found in secret %s in namespace %s. "+
							"%s already associated with secret %s in %s",
						dnsName,
						secret.Name,
						secret.Namespace,
						dnsName,
						associatedSecret.Name,
						associatedSecret.Namespace,
					),
				)
			}

			sanSet[dnsName] = secret
		}

	}

	return admission.Allowed("")
}
