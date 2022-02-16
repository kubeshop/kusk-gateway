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

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"time"

	// +kubebuilder:scaffold:imports

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	gateway "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/internal/controllers"
	"github.com/kubeshop/kusk-gateway/internal/envoy/manager"
	"github.com/kubeshop/kusk-gateway/internal/validation"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(gateway.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func initLogger(development bool, level string) (logr.Logger, error) {
	var l zapcore.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		return logr.Logger{}, fmt.Errorf("unable to determine log level: %w", err)
	}

	var config zap.Config

	if development {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	config.Level = zap.NewAtomicLevelAt(l)
	config.Development = development

	zapLogger, err := config.Build()
	if err != nil {
		return logr.Logger{}, fmt.Errorf("cannot create zap logger: %w", err)
	}

	return zapr.NewLogger(zapLogger), nil
}

func initSecretsInformer(
	log logr.Logger,
	config *rest.Config,
	secretsChan chan *corev1.Secret,
) cache.SharedIndexInformer {
	parseSecret := func(u *unstructured.Unstructured) (*corev1.Secret, error) {
		var secret corev1.Secret
		if err := runtime.DefaultUnstructuredConverter.
			FromUnstructured(u.UnstructuredContent(), &secret); err != nil {
			return nil, err
		}

		return &secret, nil
	}

	dynamicConfig := dynamic.NewForConfigOrDie(config)
	resource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	factory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicConfig, time.Minute)
	informer := factory.ForResource(resource).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {},
		UpdateFunc: func(oldObj, newObj interface{}) {
			u := newObj.(*unstructured.Unstructured)

			newSecret, err := parseSecret(u)
			if err != nil {
				log.Error(err, "unable to parse updated secret")
			}

			if newSecret.Type != corev1.SecretTypeTLS {
				return
			}

			oldU := oldObj.(*unstructured.Unstructured)
			oldSecret, err := parseSecret(oldU)
			if err != nil {
				log.Error(err, "unable to parse old secret")
				return
			}

			if reflect.DeepEqual(oldSecret.Data, newSecret.Data) {
				return
			}

			secretsChan <- newSecret
		},
		DeleteFunc: func(obj interface{}) {},
	})

	return informer
}

func main() {
	var (
		metricsAddr           string
		enableLeaderElection  bool
		probeAddr             string
		envoyControlPlaneAddr string
		logLevel              string
		development           bool
	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&envoyControlPlaneAddr, "envoy-control-plane-bind-address", ":18000", "The address Envoy control plane XDS server binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&logLevel, "log-level", "INFO", "level of log detail [DEBUG|INFO|WARN|ERROR|DPANIC|PANIC|FATAL]")
	flag.BoolVar(&development, "development", false, "enable development mode")

	flag.Parse()

	logger, err := initLogger(development, logLevel)
	if err != nil {
		_ = fmt.Errorf("unable to init logger: %w", err)
		os.Exit(1)
	}

	ctrl.SetLogger(logger)

	setupLog := logger.WithName("setup")

	restConfig := ctrl.GetConfigOrDie()

	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "cd734a2d.kusk.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	// TODO: setup logger correctly
	envoyManager := manager.New(context.Background(), envoyControlPlaneAddr, nil)

	go func() {
		if err := envoyManager.Start(); err != nil {
			setupLog.Error(err, "unable to start Envoy xDS API Server")
			os.Exit(1)
		}
	}()

	proxy := validation.NewProxy()

	go func() {
		if err := http.ListenAndServe(":17000", proxy); err != nil {
			setupLog.Error(err, "unable to start validation proxy")
			os.Exit(1)
		}
	}()

	secretsChan := make(chan *corev1.Secret)

	controllerConfigManager := controllers.KubeEnvoyConfigManager{
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		EnvoyManager: envoyManager,
		Validator:    proxy,
		SecretsChan:  secretsChan,
	}

	if err = (&controllers.EnvoyFleetReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		ConfigManager: &controllerConfigManager,
	}).SetupWithManager(mgr); err != nil {
		setupLog.
			WithValues("controller", "EnvoyFleet").
			Error(err, "unable to create controller")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go func() {
		initSecretsInformer(logger, restConfig, secretsChan).Run(ctx.Done())
	}()

	go func() {
		// start process for listening to secrets
		controllerConfigManager.WatchSecrets(ctx.Done())
	}()

	if err = (&controllers.APIReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		ConfigManager: &controllerConfigManager,
	}).SetupWithManager(mgr); err != nil {
		setupLog.
			WithValues("controller", "API").
			Error(err, "unable to create controller")
		os.Exit(1)
	}

	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		hookServer := mgr.GetWebhookServer()
		setupLog.Info("registering API mutating and validating webhooks to the webhook server")
		hookServer.Register(gateway.APIMutatingWebhookPath, &webhook.Admission{Handler: &gateway.APIMutator{Client: mgr.GetClient()}})
		hookServer.Register(gateway.APIValidatingWebhookPath, &webhook.Admission{Handler: &gateway.APIValidator{}})
	}

	if err = (&controllers.StaticRouteReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		ConfigManager: &controllerConfigManager,
	}).SetupWithManager(mgr); err != nil {
		setupLog.
			WithValues("controller", "StaticRoute").
			Error(err, "unable to create controller")
		os.Exit(1)
	}

	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		setupLog.Info("registering StaticRoute mutating and validating webhooks to the webhook server")
		hookServer := mgr.GetWebhookServer()
		hookServer.Register(gateway.StaticRouteMutatingWebhookPath, &webhook.Admission{Handler: &gateway.StaticRouteMutator{Client: mgr.GetClient()}})
		hookServer.Register(gateway.StaticRouteValidatingWebhookPath, &webhook.Admission{Handler: &gateway.StaticRouteValidator{}})
		hookServer.Register(gateway.EnvoyFleetValidatingWebhookPath, &webhook.Admission{Handler: &gateway.EnvoyFleetValidator{Client: mgr.GetClient()}})
	}
	// +kubebuilder:scaffold:builder
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
