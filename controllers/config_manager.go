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
	"github.com/kubeshop/kusk-gateway/options"
	"github.com/kubeshop/kusk-gateway/spec"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	envoyConfig := config.New()
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

		if err := opts.FillDefaultsAndValidate(); err != nil {
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
		opts, err := optionsFromStaticRouteSpec(sr.Spec)
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

// optionsFromStaticRouteSpec is a converter to generate Options object from StaticRoutes spec
func optionsFromStaticRouteSpec(spec gateway.StaticRouteSpec) (*options.StaticOptions, error) {
	// 2 dimensional map["path"]["method"]SubOptions
	paths := make(map[string]options.StaticOperationSubOptions)
	opts := &options.StaticOptions{
		Paths: paths,
		Hosts: spec.Hosts,
	}
	if err := opts.FillDefaultsAndValidate(); err != nil {
		return nil, fmt.Errorf("failed to validate options: %w", err)
	}
	for specPath, specMethods := range spec.Paths {
		path := string(specPath)
		opts.Paths[path] = make(options.StaticOperationSubOptions)
		pathMethods := opts.Paths[path]
		for specMethod, specRouteAction := range specMethods {
			methodOpts := &options.StaticSubOptions{}
			pathMethods[specMethod] = methodOpts
			if specRouteAction.Redirect != nil {
				methodOpts.Redirect = specRouteAction.Redirect
				continue
			}
			if specRouteAction.Route != nil {
				methodOpts.Backend = *&specRouteAction.Route.Backend
				if specRouteAction.Route.CORS != nil {
					methodOpts.CORS = specRouteAction.Route.CORS.DeepCopy()
				}
				if specRouteAction.Route.Timeouts != nil {
					methodOpts.Timeouts = specRouteAction.Route.Timeouts
				}
			}
		}
	}
	return opts, opts.Validate()
}
