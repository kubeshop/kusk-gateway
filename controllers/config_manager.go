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

package controllers

import (
	"context"
	"fmt"
	"strings"
	"sync"

	gateway "github.com/kubeshop/kusk-gateway/api/v1"
	"github.com/kubeshop/kusk-gateway/envoy/config"
	"github.com/kubeshop/kusk-gateway/envoy/manager"
	"github.com/kubeshop/kusk-gateway/spec"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KubeEnvoyConfigManager manages all Envoy configurations from CRDs
type KubeEnvoyConfigManager struct {
	client.Client
	Scheme       *runtime.Scheme
	EnvoyManager *manager.EnvoyConfigManager
	m            sync.Mutex
}

var (
	configManagerLogger = ctrl.Log.WithName("controller.config-manager")
)

func (c *KubeEnvoyConfigManager) UpdateConfiguration(ctx context.Context) error {

	l := configManagerLogger
	// acquiring this lock is required so that no potentially conflicting updates would happen at the same time
	// this probably should be done on a per-envoy basis but as we have a static config for now this will do
	c.m.Lock()
	defer c.m.Unlock()
	l.Info("Started updating configuration")
	defer l.Info("Finished updating configuration")

	parser := spec.NewParser(nil)

	// fetch all APIs and Static Routes to rebuild Envoy configuration
	var staticRoutes gateway.StaticRouteList

	l.Info("Getting Static Routes")
	if err := c.Client.List(ctx, &staticRoutes); err != nil {
		return err
	}
	l.Info("Getting APIs")
	var apis gateway.APIList
	if err := c.Client.List(ctx, &apis); err != nil {
		return err
	}

	envoyConfig := config.New()
	for _, api := range apis.Items {
		apiSpec, err := parser.ParseFromReader(strings.NewReader(api.Spec.Spec))
		if err != nil {
			return fmt.Errorf("failed to parse OpenAPI spec: %w", err)
		}

		opts, err := spec.GetOptions(apiSpec)
		if err != nil {
			return fmt.Errorf("failed to parse options: %w", err)
		}

		err = opts.FillDefaultsAndValidate()
		if err != nil {
			return fmt.Errorf("failed to validate options: %w", err)
		}

		if err = envoyConfig.UpdateConfigFromOpts(opts, apiSpec); err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}
		l.Info("API route configuration processed", "api", api)
	}
	for _, sr := range staticRoutes.Items {
		l.Info("Static route processed", "route", sr)
	}

	snapshot, err := envoyConfig.GenerateSnapshot()
	if err != nil {
		l.Error(err, "Envoy configuration snapshot is invalid")
		return fmt.Errorf("failed to generate snapshot: %w", err)
	}

	l.Info("Configuration snapshot generated")
	if err := c.EnvoyManager.ApplyNewFleetSnapshot(manager.DefaultFleetName, snapshot); err != nil {
		l.Error(err, "Envoy configuration failed to apply")
		return fmt.Errorf("failed to apply snapshot: %w", err)
	}

	return nil
}
