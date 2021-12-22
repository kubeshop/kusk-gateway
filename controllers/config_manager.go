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

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gateway "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/envoy/config"
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
func (c *KubeEnvoyConfigManager) UpdateConfiguration(ctx context.Context, fleetID gateway.EnvoyFleetID) error {

	l := configManagerLogger

	// acquiring this lock is required so that no potentially conflicting updates would happen at the same time
	// this probably should be done on a per-envoy basis but as we have a static config for now this will do
	c.m.Lock()
	defer c.m.Unlock()

	l.Info("Started updating configuration for the fleet", "fleetID", fleetID.String())
	defer l.Info("Finished updating configuration for the fleet", "fleetID", fleetID.String())

	envoyConfig := config.New()
	// fetch all APIs and Static Routes to rebuild Envoy configuration
	l.Info("Getting APIs for fleet", "fleetID", fleetID)

	var apis gateway.APIList
	// Get all API objects with this fleet id field set
	if err := c.Client.List(ctx, &apis,
		&client.ListOptions{
			FieldSelector: client.MatchingFieldsSelector{
				Selector: fields.OneTermEqualSelector("spec.fleet", fleetID.String()),
			},
		},
	); err != nil {
		return err
	}

	parser := spec.NewParser(nil)
	for _, api := range apis.Items {
		l.Info("Processing API configuration", "api", api)
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

		if err = UpdateConfigFromAPIOpts(envoyConfig, opts, apiSpec); err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}
		l.Info("API route configuration processed", "api", api)
	}

	l.Info("Succesfully processed APIs for the fleet", "fleetID", fleetID.String())
	l.Info("Getting Static Routes for the fleet", "fleetID", fleetID.String())
	var staticRoutes gateway.StaticRouteList
	if err := c.Client.List(ctx, &staticRoutes,
		&client.ListOptions{
			FieldSelector: client.MatchingFieldsSelector{
				Selector: fields.OneTermEqualSelector("spec.fleet", fleetID.String()),
			},
		},
	); err != nil {
		return err
	}
	for _, sr := range staticRoutes.Items {
		l.Info("Processing static routes", "route", sr)
		opts, err := sr.Spec.GetOptionsFromSpec()
		if err != nil {
			return fmt.Errorf("failed to generate options from the static route config: %w", err)
		}

		if err := UpdateConfigFromOpts(envoyConfig, opts); err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}
	}

	l.Info("Succesfully processed Static Routes")

	l.Info("Processing EnvoyFleet configuration")
	var fleet gateway.EnvoyFleet
	if err := c.Client.Get(ctx, types.NamespacedName{Name: fleetID.Name, Namespace: fleetID.Namespace}, &fleet); err != nil {
		l.Error(err, "Failed to get Envoy Fleet", "fleetID", fleetID.String())
		return fmt.Errorf("failed to get Envoy Fleet %s: %w", fleetID.String(), err)
	}
	httpConnectionManagerBuilder := config.NewHCMBuilder()
	if fleet.Spec.AccessLog != nil {
		var accessLogBuilder *config.AccessLogBuilder
		var err error
		// Depending on the Format (text or json) we send different format templates or empty interface
		switch fleet.Spec.AccessLog.Format {
		case config.AccessLogFormatText:
			accessLogBuilder, err = config.NewTextAccessLog(fleet.Spec.AccessLog.TextTemplate)
			if err != nil {
				l.Error(err, "Failure creating new text access log builder")
				return fmt.Errorf("failure creating new text access log builder: %w", err)
			}
		case config.AccessLogFormatJson:
			accessLogBuilder, err = config.NewJSONAccessLog(fleet.Spec.AccessLog.JsonTemplate)
			if err != nil {
				l.Error(err, "Failure creating new JSON access log builder")
				return fmt.Errorf("failure creating new JSON access log builder: %w", err)
			}
		default:
			err := fmt.Errorf("unknown access log format %s", fleet.Spec.AccessLog.Format)
			l.Error(err, "Failure adding access logger to Envoy configuration")
			return err
		}
		httpConnectionManagerBuilder.AddAccessLog(accessLogBuilder.GetAccessLog())
	}
	if err := httpConnectionManagerBuilder.Validate(); err != nil {
		l.Error(err, "Failed validation for HttpConnectionManager")
		return fmt.Errorf("failed validation for HttpConnectionManager")
	}
	listenerBuilder := config.NewListenerBuilder()
	listenerBuilder.AddHTTPManagerFilterChain(httpConnectionManagerBuilder.GetHTTPConnectionManager())
	if err := listenerBuilder.Validate(); err != nil {
		l.Error(err, "Failed validation for the Listener")
		return fmt.Errorf("failed validation for Listener")

	}
	envoyConfig.AddListener(listenerBuilder.GetListener())
	l.Info("Generating configuration snapshot")
	snapshot, err := envoyConfig.GenerateSnapshot()
	if err != nil {
		l.Error(err, "Envoy configuration snapshot is invalid")
		return fmt.Errorf("failed to generate snapshot: %w", err)
	}

	l.Info("Configuration snapshot generated for the fleet", "fleetID", fleetID.String())
	if err := c.EnvoyManager.ApplyNewFleetSnapshot(fleetID.String(), snapshot); err != nil {
		l.Error(err, "Envoy configuration failed to apply")
		return fmt.Errorf("failed to apply snapshot: %w", err)
	}

	return nil
}
