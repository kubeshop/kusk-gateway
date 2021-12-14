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

package controllers

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gateway "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/envoy"
	"github.com/kubeshop/kusk-gateway/envoy/manager"
	"github.com/kubeshop/kusk-gateway/spec"
)

// KubeEnvoyConfigManager manages all Envoy configurations parsing from CRDs
type KubeEnvoyConfigManager struct {
	client.Client
	Scheme       *runtime.Scheme
	EnvoyManager *manager.EnvoyConfigManager
	m            sync.Mutex
}

var (
	configManagerLogger = ctrl.Log.WithName("controller.config-manager")
)

// UpdateConfiguration is the main method to gather all routing configs and to create and apply Envoy config
func (c *KubeEnvoyConfigManager) UpdateConfiguration(ctx context.Context) error {

	l := configManagerLogger

	// acquiring this lock is required so that no potentially conflicting updates would happen at the same time
	// this probably should be done on a per-envoy basis but as we have a static config for now this will do
	c.m.Lock()
	defer c.m.Unlock()

	l.Info("Started updating configuration")
	defer l.Info("Finished updating configuration")

	parser := spec.NewParser(nil)
	envoyConfig := envoy.NewConfiguration()

	// fetch all APIs and Static Routes to rebuild Envoy configuration
	l.Info("Getting APIs")
	var apis gateway.APIList
	if err := c.Client.List(ctx, &apis); err != nil {
		return err
	}
	for _, api := range apis.Items {
		l.Info("Processing API %v", "api", api)
		apiSpec, err := parser.ParseFromReader(strings.NewReader(api.Spec.Spec))
		if err != nil {
			return fmt.Errorf("failed to parse OpenAPI spec: %w", err)
		}

		opts, err := spec.GetOptions(apiSpec)
		if err != nil {
			return fmt.Errorf("failed to parse options: %w", err)
		}
		opts.FillDefaults()
		if err := opts.Validate(); err != nil {
			return fmt.Errorf("failed to validate options: %w", err)
		}

		if err = envoyConfig.UpdateConfigFromAPIOpts(opts, apiSpec); err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}
		l.Info("API route configuration processed", "api", api)
	}

	l.Info("Succesfully processed APIs")
	l.Info("Getting Static Routes")
	var staticRoutes gateway.StaticRouteList
	if err := c.Client.List(ctx, &staticRoutes); err != nil {
		return err
	}
	for _, sr := range staticRoutes.Items {
		l.Info("Processing static routes", "route", sr)
		opts, err := sr.Spec.GetOptionsFromSpec()
		if err != nil {
			return fmt.Errorf("failed to generate options from the static route config: %w", err)
		}

		if err := envoyConfig.UpdateConfigFromOpts(opts); err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}
	}

	l.Info("Succesfully processed Static Routes")
	l.Info("Generating configuration snapshot")
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
