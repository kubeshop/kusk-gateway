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

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"time"

	// +kubebuilder:scaffold:imports

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	gateway "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	agentManagement "github.com/kubeshop/kusk-gateway/internal/agent/management"
	"github.com/kubeshop/kusk-gateway/internal/controllers"
	"github.com/kubeshop/kusk-gateway/internal/envoy/manager"
	"github.com/kubeshop/kusk-gateway/internal/validation"
	"github.com/kubeshop/kusk-gateway/internal/webhooks"
	"github.com/kubeshop/kusk-gateway/pkg/analytics"
)

var (
	scheme = runtime.NewScheme()
	config managerConfig
)

type managerConfig struct {
	MetricsAddr           string `envconfig:"METRICS_BIND_ADDR" default:":8080"`
	ProbeAddr             string `envconfig:"HEALTH_PROBE_BIND_ADDR" default:":8081"`
	EnvoyControlPlaneAddr string `envconfig:"ENVOY_CONTROL_PLANE_BIND_ADDR" default:":18000"`
	AgentManagerAddr      string `envconfig:"AGENT_MANAGER_BIND_ADDR" default:"18010"`
	EnableLeaderElection  bool   `envconfig:"ENABLE_LEADER_ELECTION" default:"false"`
	LogLevel              string `envconfig:"LOG_LEVEL" default:"INFO"`
	WebhookCertsDir       string `envconfig:"WEBHOOK_CERTS_DIR" default:"/opt/manager/webhook/certs"`
}

func (m managerConfig) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("METRICS_BIND_ADDR=%s\n", m.MetricsAddr))
	b.WriteString(fmt.Sprintf("HEALTH_PROBE_BIND_ADDR=%s\n", m.ProbeAddr))
	b.WriteString(fmt.Sprintf("ENVOY_CONTROL_PLANE_BIND_ADDR=%s\n", m.EnvoyControlPlaneAddr))
	b.WriteString(fmt.Sprintf("AGENT_MANAGER_BIND_ADDR=%s\n", m.AgentManagerAddr))
	b.WriteString(fmt.Sprintf("ENABLE_LEADER_ELECTION=%t\n", m.EnableLeaderElection))
	b.WriteString(fmt.Sprintf("LOG_LEVEL=%s\n", m.LogLevel))
	b.WriteString(fmt.Sprintf("WEBHOOK_CERTS_DIR=%s\n", m.WebhookCertsDir))

	return b.String()
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(gateway.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme

	if err := envconfig.Process("manager", &config); err != nil {
		panic(fmt.Errorf("unable to process config %w", err))
	}

	fmt.Println(config)
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

// initWebhookCerts creates the Admission webhooks server certificates in the predefined location during each manager start
// and patches the K8s Kusk Gateway Validating and Mutating Admission webhooks configurations with the generated self-signed CA.
func initWebhookCerts(ctx context.Context, webhookCertsDir string, webhookServer *webhook.Server, clientSet *kubernetes.Clientset) error {
	webhookServer.CertDir = webhookCertsDir
	webhookServer.CertName = "tls.crt"
	webhookServer.KeyName = "tls.key"
	webhooksServiceDNSNames, err := webhooks.GetWebhookServiceDNSNames(ctx, clientSet)
	if err != nil {
		return fmt.Errorf("failure looking up the webhooks service: %w", err)
	}

	caCert, err := webhooks.CreateCertificates(webhooksServiceDNSNames, webhookServer.CertDir, webhookServer.CertName, webhookServer.KeyName)
	if err != nil {
		return fmt.Errorf("failure creating webhooks certificates: %w", err)
	}
	if err := webhooks.UpdateWebhookConfiguration(ctx, clientSet, caCert); err != nil {
		return fmt.Errorf("failure patching webhooks configuration: %w", err)
	}
	return nil
}

func main() {
	analytics.SendAnonymousInfo("kusk-gateway manager bootstrapping")
	logger, err := initLogger(false, config.LogLevel)
	if err != nil {
		_ = fmt.Errorf("unable to init logger: %w", err)
		os.Exit(1)
	}
	ctrl.SetLogger(logger)
	setupLog := logger.WithName("setup")

	// Create the context obj with the signal to manage the subroutines termination
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	restConfig := ctrl.GetConfigOrDie()

	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     config.MetricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: config.ProbeAddr,
		LeaderElection:         config.EnableLeaderElection,
		LeaderElectionID:       "cd734a2d.kusk.io",
	})
	if err != nil {
		setupLog.Error(err, "Unable to create controller manager")
		os.Exit(1)
	}
	// Envoy configuration manager (XDS service)
	envoyManager := manager.New(ctx, config.EnvoyControlPlaneAddr, nil)
	go func() {
		setupLog.Info("Starting Envoy xDS API Server")
		if err := envoyManager.Start(); err != nil {
			setupLog.Error(err, "Unable to start Envoy xDS API Server")
			os.Exit(1)
		}
	}()

	// Validation proxy
	proxy := validation.NewProxy()
	go func() {
		if err := http.ListenAndServe(":17000", proxy); err != nil {
			setupLog.Error(err, "Unable to start validation proxy")
			os.Exit(1)
		}
	}()

	// Agent (Envoy sidecar) configuration management service
	agentManager := agentManagement.New(config.AgentManagerAddr, logger)
	go func() {
		if err := agentManager.Start(); err != nil {
			setupLog.Error(err, "Unable to start Agent Manager Server")
			os.Exit(1)
		}
	}()

	secretsChan := make(chan *corev1.Secret)
	controllerConfigManager := controllers.KubeEnvoyConfigManager{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		EnvoyManager:       envoyManager,
		Validator:          proxy,
		AgentManager:       agentManager,
		SecretToEnvoyFleet: map[string]gateway.EnvoyFleetID{},
		WatchedSecretsChan: secretsChan,
	}

	// The watcher for k8s secrets to trigger the refresh of configuration in case certificates secrets change.
	go func() {
		initSecretsInformer(logger, restConfig, secretsChan).Run(ctx.Done())
	}()
	go func() {
		// start process for listening to secrets
		setupLog.Info("Starting K8s secrets watch for the TLS certificates renewal events")
		controllerConfigManager.WatchSecrets(ctx.Done())
	}()

	if err = (&controllers.ServiceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.
			WithValues("controller", "ServiceAnnotationReconciler").
			Error(err, "Unable to create controller")
		os.Exit(1)
	}

	// EnvoyFleet obj controller
	if err = (&controllers.EnvoyFleetReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		ConfigManager: &controllerConfigManager,
	}).SetupWithManager(mgr); err != nil {
		setupLog.
			WithValues("controller", "EnvoyFleet").
			Error(err, "Unable to create controller")
		os.Exit(1)
	}
	webhookServer := mgr.GetWebhookServer()
	if err := initWebhookCerts(ctx, config.WebhookCertsDir, webhookServer, kubernetes.NewForConfigOrDie(restConfig)); err != nil {
		setupLog.Error(err, "Failure initializing admission webhook server certs")
		os.Exit(1)
	}
	setupLog.Info("Created admission webhook server certificates and updated K8s Manager's Admission configs with the generated CA certificate")

	setupLog.Info("Registering EnvoyFleet mutating and validating webhooks to the webhook server")
	webhookServer.Register(gateway.EnvoyFleetMutatingWebhookPath, &webhook.Admission{Handler: &gateway.EnvoyFleetMutator{}})
	webhookServer.Register(gateway.EnvoyFleetValidatingWebhookPath, &webhook.Admission{Handler: &gateway.EnvoyFleetValidator{Client: mgr.GetClient()}})

	// API obj controller
	if err = (&controllers.APIReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		ConfigManager: &controllerConfigManager,
	}).SetupWithManager(mgr); err != nil {
		setupLog.
			WithValues("controller", "API").
			Error(err, "Unable to create controller")
		os.Exit(1)
	}

	setupLog.Info("Registering API mutating and validating webhooks to the webhook server")
	webhookServer.Register(gateway.APIMutatingWebhookPath, &webhook.Admission{Handler: &gateway.APIMutator{Client: mgr.GetClient()}})
	webhookServer.Register(gateway.APIValidatingWebhookPath, &webhook.Admission{Handler: &gateway.APIValidator{}})

	// StaticRoute obj controller
	if err = (&controllers.StaticRouteReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		ConfigManager: &controllerConfigManager,
	}).SetupWithManager(mgr); err != nil {
		setupLog.
			WithValues("controller", "StaticRoute").
			Error(err, "Unable to create controller")
		os.Exit(1)
	}

	setupLog.Info("Registering StaticRoute mutating and validating webhooks to the webhook server")
	webhookServer.Register(gateway.StaticRouteMutatingWebhookPath, &webhook.Admission{Handler: &gateway.StaticRouteMutator{Client: mgr.GetClient()}})
	webhookServer.Register(gateway.StaticRouteValidatingWebhookPath, &webhook.Admission{Handler: &gateway.StaticRouteValidator{}})

	// +kubebuilder:scaffold:builder
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Problem running manager")
		os.Exit(1)
	}
}
