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
package controllers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gateway "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/pkg/analytics"
)

const (
	xKuskAnnotation               = "x-kusk"
	annotationOpenapiUrl          = "openapi-url"
	annotationApiPathPrefix       = "path-prefix"
	annotationApiPathSubstitution = "path-prefix-substitution"
	annotationEnvoyFleet          = "envoy-fleet"
)

// ServiceReconciler reconciles a Pod object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func annotation(a string) string {
	return fmt.Sprintf("kusk-gateway/%s", a)
}

//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithName("service-reconciler")
	analytics.SendAnonymousInfo(ctx, r.Client, "reconciling Service annotations for kusk")

	l.Info("Reconciling changed Service resource", "changed", req.NamespacedName)

	svc := &corev1.Service{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: req.Name}, svc); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	l = l.WithValues(
		"serviceName", svc.Name,
		"serviceNamespace", svc.Namespace,
	)

	openAPIUrlAnnotation := annotation(annotationOpenapiUrl)
	openApiUrl, ok := svc.Annotations[openAPIUrlAnnotation]
	if !ok {
		// if the service doesn't have the kusk-gateway/openapi-url annotation then we dont do anything
		// as this is the minimum requirement for the service reconciler to have an effect
		return ctrl.Result{}, nil
	}

	l.Info(`Detected annotation`, "annotation", openAPIUrlAnnotation, "value", openApiUrl)

	// fetch initial open api spec from url which we will build on
	openApiSpec, err := processOpenAPIURLAnnotation(req, openApiUrl, svc.Spec.Ports[0].Port)
	if err != nil {
		return ctrl.Result{}, err
	}

	processPathPrefixAnnotation(l, openApiSpec, svc.Annotations)
	processSubstitutionAnnotation(l, openApiSpec, svc.Annotations)

	yamlPayload, err := yaml.Marshal(openApiSpec)
	if err != nil {
		return ctrl.Result{}, err
	}

	envoyFleet := getEnvoyFleetFromAnnotations(l, svc.Annotations)

	gatewaySpec := gateway.APISpec{
		Fleet: envoyFleet,
		Spec:  string(yamlPayload),
	}

	api := &gateway.API{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: req.Name}, api); err != nil && errors.IsNotFound(err) {
		//create
		api.Name = req.Name
		api.Namespace = req.Namespace
		api.Spec = gatewaySpec
		if err := r.Client.Create(ctx, api, &client.CreateOptions{}); err != nil {
			l.Error(err, "error occured while creating API")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	api.Spec = gatewaySpec
	if err := r.Client.Update(ctx, api, &client.UpdateOptions{}); err != nil {
		l.Error(err, "error occured while updating API")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Complete(r)
}

func processOpenAPIURLAnnotation(req ctrl.Request, url string, svcPort int32) (map[string]interface{}, error) {
	bOpenAPISpec, err := getOpenAPIfromURL(url)
	if err != nil {
		return nil, err
	}

	var openApiSpec map[string]interface{}
	if err := yaml.Unmarshal(bOpenAPISpec, &openApiSpec); err != nil {
		return nil, err
	}

	service := map[string]interface{}{
		"service": map[string]interface{}{
			"name":      req.Name,
			"namespace": req.Namespace,
			"port":      svcPort,
		},
	}
	upstream := map[string]interface{}{
		"upstream": service,
	}

	if _, ok := openApiSpec[xKuskAnnotation]; !ok {
		openApiSpec[xKuskAnnotation] = upstream
	}

	kusk := openApiSpec[xKuskAnnotation]
	if xkusk, ok := kusk.(map[string]interface{}); ok {
		if _, contains := xkusk["upstream"]; !contains {
			xkusk["upstream"] = service
			openApiSpec[xKuskAnnotation] = xkusk
		}
	}

	return openApiSpec, nil
}

func getOpenAPIfromURL(u string) ([]byte, error) {
	if _, err := url.Parse(u); err != nil {
		return nil, fmt.Errorf("invalid url %s: %w", u, err)
	}

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, err
}

func processPathPrefixAnnotation(l logr.Logger, openApiSpec map[string]interface{}, svcAnnotations map[string]string) {
	pathPrefixAnnotation := annotation(annotationApiPathPrefix)
	pathPrefix, ok := svcAnnotations[pathPrefixAnnotation]
	if !ok {
		// a path is required to properly configure an API. We are making the assumption that
		// they want the api to be hosted at `/` if the user omits this annotation
		pathPrefix = "/"
		l.Info("no path prefix annotation set, defaulting to /")
	} else {
		l.Info(`Detected annotation`, "annotation", pathPrefixAnnotation, "value", pathPrefix)
	}

	xKusk, ok := openApiSpec[xKuskAnnotation].(map[string]interface{})
	if !ok {
		xKusk = map[string]interface{}{}
	}

	if _, ok := xKusk["path"]; !ok {
		xKusk["path"] = map[string]string{
			"prefix": pathPrefix,
		}
		openApiSpec[xKuskAnnotation] = xKusk
	}
}

func processSubstitutionAnnotation(l logr.Logger, openApiSpec map[string]interface{}, svcAnnotations map[string]string) {
	substitutionAnnotation := annotation(annotationApiPathSubstitution)
	pathSubstitution, ok := svcAnnotations[substitutionAnnotation]
	if !ok {
		// we only substitute if an annotation is explicitly set
		return
	}

	pathPrefixAnnotation := annotation(annotationApiPathPrefix)
	pathPrefix, ok := svcAnnotations[pathPrefixAnnotation]
	if !ok {
		pathPrefix = "/"
	}

	xKusk, ok := openApiSpec[xKuskAnnotation].(map[string]interface{})
	if !ok {
		xKusk = map[string]interface{}{}
	}

	xKuskUpstream, ok := xKusk["upstream"].(map[string]interface{})
	if !ok {
		xKuskUpstream = map[string]interface{}{}
	}

	if _, ok := xKuskUpstream["rewrite"]; !ok && pathPrefix != "/" {
		l.Info(fmt.Sprintf("path prefix is not /. setting path substitution to \"%s\"", pathSubstitution))
		xKuskUpstream["rewrite"] = map[string]interface{}{
			"pattern":      fmt.Sprintf("^%s", pathPrefix),
			"substitution": pathSubstitution,
		}

		xKusk["upstream"] = xKuskUpstream
		openApiSpec[xKuskAnnotation] = xKusk
	}
}

func getEnvoyFleetFromAnnotations(l logr.Logger, svcAnnotations map[string]string) *gateway.EnvoyFleetID {
	defaultEnvoyFleet := &gateway.EnvoyFleetID{
		Name:      "default",
		Namespace: "default",
	}

	envoyFleetAnnotation := annotation(annotationEnvoyFleet)
	if envoyFleet, ok := svcAnnotations[envoyFleetAnnotation]; ok {
		// valid envoy fleet annotation value should be of the form `envofleetname.namespace`
		splitEnvoyFleetString := strings.Split(envoyFleet, ".")
		if len(splitEnvoyFleetString) < 2 {
			// if string is not in the valid form, return the default fleet
			// we should revisit this because this could be seen as a "silent failure"
			l.Info("invalid envoy fleet annotation value, using default envoy fleet", "invalidValue", envoyFleet)
			return defaultEnvoyFleet
		}

		l.Info("using envoyfleet", "envoyfleet", envoyFleet)

		return &gateway.EnvoyFleetID{
			Name:      splitEnvoyFleetString[0],
			Namespace: splitEnvoyFleetString[1],
		}
	}

	l.Info("no envoy fleet annotation found, using default envoy fleet")

	return defaultEnvoyFleet
}
