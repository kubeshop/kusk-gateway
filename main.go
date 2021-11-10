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

package main

import (
	"context"
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	gatewayv1 "github.com/kubeshop/kusk-gateway/api/v1"
	"github.com/kubeshop/kusk-gateway/controllers"
	"github.com/kubeshop/kusk-gateway/envoy/manager"
	"github.com/kubeshop/kusk-gateway/local"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(gatewayv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var envoyControlPlaneAddr string
	var openAPIspec string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&envoyControlPlaneAddr, "envoy-control-plane-bind-address", ":18000", "The address Envoy control plane XDS server binds to.")
	flag.StringVar(&openAPIspec, "in", "", "OpenAPI file with x-kusk extension to start gateway locally, without Kubernetes")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// If -in is specified, use its parameter as OpenAPI file and switch to local startup
	// This will never return
	if openAPIspec != "" {
		setupLog.Info("Specified ", openAPIspec, "file will be consumed locally, skipping K8s initialisation")
		local.RunLocalService(openAPIspec, envoyControlPlaneAddr)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
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

	// TODO: initialize envoyManager with static envoy deployment configuration

	go func() {
		if err := envoyManager.Start(); err != nil {
			setupLog.Error(err, "unable to start Envoy xDS API Server")
			os.Exit(1)
		}
	}()
	if err = (&controllers.EnvoyFleetReconciler{
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		EnvoyManager: envoyManager,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "EnvoyFleet")
		os.Exit(1)
	}
	controllerConfigManager := controllers.KubeEnvoyConfigManager{
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		EnvoyManager: envoyManager,
	}
	if err = (&controllers.APIReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		ConfigManager: &controllerConfigManager,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "API")
		os.Exit(1)
	}

	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err = (&gatewayv1.API{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "API")
			os.Exit(1)
		}
	}
	if err = (&controllers.StaticRouteReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		ConfigManager: &controllerConfigManager,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "StaticRoute")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

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
