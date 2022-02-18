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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gateway "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/internal/envoy/config"
	"github.com/kubeshop/kusk-gateway/internal/envoy/manager"
	"github.com/kubeshop/kusk-gateway/internal/spec"
	"github.com/kubeshop/kusk-gateway/internal/validation"

	helperManagement "github.com/kubeshop/kusk-gateway/internal/helper/management"
	"github.com/kubeshop/kusk-gateway/internal/helper/mocking"
)

const (
	tlsKey = "tls.key"
	tlsCrt = "tls.crt"
)

// KubeEnvoyConfigManager manages all Envoy configurations parsing from CRDs
type KubeEnvoyConfigManager struct {
	client.Client
	Scheme        *runtime.Scheme
	EnvoyManager  *manager.EnvoyConfigManager
	HelperManager *helperManagement.ConfigManager
	Validator     *validation.Proxy
	m             sync.Mutex
}

var (
	configManagerLogger = ctrl.Log.WithName("controller.config-manager")
)

// UpdateConfiguration is the main method to gather all routing configs and to create and apply Envoy config
func (c *KubeEnvoyConfigManager) UpdateConfiguration(ctx context.Context, fleetID gateway.EnvoyFleetID) error {

	l := configManagerLogger
	fleetIDstr := fleetID.String()
	// acquiring this lock is required so that no potentially conflicting updates would happen at the same time
	// this probably should be done on a per-envoy basis but as we have a static config for now this will do
	c.m.Lock()
	defer c.m.Unlock()

	l.Info("Started updating configuration", "fleet", fleetIDstr)
	defer l.Info("Finished updating configuration", "fleet", fleetIDstr)

	envoyConfig := config.New()

	mockingConfig := mocking.NewMockConfig()
	// fetch all APIs and Static Routes to rebuild Envoy configuration
	l.Info("Getting APIs for the fleet", "fleet", fleetIDstr)

	apis, err := c.getDeployedAPIs(ctx, fleetIDstr)
	if err != nil {
		l.Error(err, "Failed getting APIs for the fleet", "fleet", fleetIDstr)
		return err
	}

	parser := spec.NewParser(nil)
	for _, api := range apis {
		l.Info("Processing API configuration", "fleet", fleetIDstr, "api", api.Name)
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

		if err = UpdateConfigFromAPIOpts(envoyConfig, mockingConfig, c.Validator, opts, apiSpec); err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}
		l.Info("API route configuration processed", "fleet", fleetIDstr, "api", api.Name)
	}

	l.Info("Successfully processed APIs", "fleet", fleetIDstr)
	l.Info("Getting Static Routes", "fleet", fleetIDstr)
	staticRoutes, err := c.getDeployedStaticRoutes(ctx, fleetIDstr)
	if err != nil {
		l.Error(err, "Failed getting StaticRoutes for the fleet", "fleet", fleetIDstr)
		return err
	}
	for _, sr := range staticRoutes {
		l.Info("Processing static routes", "fleet", fleetIDstr, "route", sr.Name)
		opts, err := sr.Spec.GetOptionsFromSpec()
		if err != nil {
			return fmt.Errorf("failed to generate options from the static route config: %w", err)
		}

		if err := UpdateConfigFromOpts(envoyConfig, opts); err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}
	}

	l.Info("Successfully processed Static Routes", "fleet", fleetIDstr)

	l.Info("Processing EnvoyFleet configuration", "fleet", fleetIDstr)
	var fleet gateway.EnvoyFleet
	if err := c.Client.Get(ctx, types.NamespacedName{Name: fleetID.Name, Namespace: fleetID.Namespace}, &fleet); err != nil {
		l.Error(err, "Failed to get Envoy Fleet", "fleet", fleetIDstr)
		return fmt.Errorf("failed to get Envoy Fleet %s: %w", fleetIDstr, err)
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
				l.Error(err, "Failure creating new text access log builder", "fleet", fleetIDstr)
				return fmt.Errorf("failure creating new text access log builder: %w", err)
			}
		case config.AccessLogFormatJson:
			accessLogBuilder, err = config.NewJSONAccessLog(fleet.Spec.AccessLog.JsonTemplate)
			if err != nil {
				l.Error(err, "Failure creating new JSON access log builder", "fleet", fleetIDstr)
				return fmt.Errorf("failure creating new JSON access log builder: %w", err)
			}
		default:
			err := fmt.Errorf("unknown access log format %s", fleet.Spec.AccessLog.Format)
			l.Error(err, "Failure adding access logger to Envoy configuration", "fleet", fleetIDstr)
			return err
		}
		httpConnectionManagerBuilder.AddAccessLog(accessLogBuilder.GetAccessLog())
	}
	if err := httpConnectionManagerBuilder.Validate(); err != nil {
		l.Error(err, "Failed validation for HttpConnectionManager", "fleet", fleetIDstr)
		return fmt.Errorf("failed validation for HttpConnectionManager")
	}

	tlsConfig := config.TLS{
		CipherSuites:              fleet.Spec.TLS.CipherSuites,
		TlsMinimumProtocolVersion: fleet.Spec.TLS.TlsMinimumProtocolVersion,
		TlsMaximumProtocolVersion: fleet.Spec.TLS.TlsMaximumProtocolVersion,
	}

	for _, cert := range fleet.Spec.TLS.TlsSecrets {
		var secret v1.Secret
		err := c.Client.Get(ctx, types.NamespacedName{Name: cert.SecretRef, Namespace: cert.Namespace}, &secret)
		if err != nil {
			return fmt.Errorf("failed to get secret %s in namespace %s: %w", cert.SecretRef, cert.Namespace, err)
		}

		key, ok := secret.Data[tlsKey]
		if !ok {
			return fmt.Errorf("%s data not present in secret %s in namepspace %s", key, cert.SecretRef, cert.Namespace)
		}

		crt, ok := secret.Data[tlsCrt]
		if !ok {
			return fmt.Errorf("%s data not present in secret %s in namepspace %s", tlsCrt, cert.SecretRef, cert.Namespace)
		}

		tlsConfig.Certificates = append(tlsConfig.Certificates, config.Certificate{
			Cert: string(crt),
			Key:  string(key),
		})
	}

	listenerBuilder := config.NewListenerBuilder()
	if err := listenerBuilder.AddHTTPManagerFilterChains(httpConnectionManagerBuilder.GetHTTPConnectionManager(), tlsConfig); err != nil {
		return err
	}

	if err := listenerBuilder.Validate(); err != nil {
		l.Error(err, "Failed validation for the Listener", "fleet", fleetIDstr)
		return fmt.Errorf("failed validation for Listener")

	}
	envoyConfig.AddListener(listenerBuilder.GetListener())
	l.Info("Generating configuration snapshot", "fleet", fleetIDstr)
	snapshot, err := envoyConfig.GenerateSnapshot()
	if err != nil {
		l.Error(err, "Envoy configuration snapshot is invalid", "fleet", fleetIDstr)
		return fmt.Errorf("failed to generate snapshot: %w", err)
	}

	l.Info("Configuration snapshot was generated for the fleet", "fleet", fleetIDstr)
	// Helper config is applied first to make Helper ready for the new Envoy routes
	c.HelperManager.ApplyNewFleetConfig(fleetIDstr, mockingConfig)
	if err := c.EnvoyManager.ApplyNewFleetSnapshot(fleetIDstr, snapshot); err != nil {
		l.Error(err, "Envoy configuration failed to apply", "fleet", fleetIDstr)
		return fmt.Errorf("failed to apply snapshot: %w", err)
	}

	l.Info("Configuration snapshot deployed for the fleet", "fleet", fleetIDstr)
	return nil
}

func (c *KubeEnvoyConfigManager) getDeployedAPIs(ctx context.Context, fleet string) ([]gateway.API, error) {
	var apiObjs gateway.APIList
	// Get all API objects with this fleet field set
	if err := c.Client.List(ctx, &apiObjs,
		&client.ListOptions{
			FieldSelector: client.MatchingFieldsSelector{
				Selector: fields.AndSelectors(
					fields.OneTermEqualSelector("spec.fleet", fleet),
				),
			},
		},
	); err != nil {
		return nil, fmt.Errorf("failure querying for the deployed APIs: %w", err)
	}
	var apis []gateway.API
	// filter out apis are in the process of deletion
	for _, api := range apiObjs.Items {
		if api.ObjectMeta.DeletionTimestamp.IsZero() {
			apis = append(apis, api)
		}
	}
	return apis, nil
}

func (c *KubeEnvoyConfigManager) getDeployedStaticRoutes(ctx context.Context, fleet string) ([]gateway.StaticRoute, error) {
	var staticRoutesObjs gateway.StaticRouteList
	if err := c.Client.List(ctx, &staticRoutesObjs,
		&client.ListOptions{
			FieldSelector: client.MatchingFieldsSelector{
				Selector: fields.OneTermEqualSelector("spec.fleet", fleet),
			},
		},
	); err != nil {
		return nil, fmt.Errorf("failure querying for the deployed StaticRoutes: %w", err)
	}
	var staticRoutes []gateway.StaticRoute
	// filter out apis are in the process of deletion
	for _, staticRoute := range staticRoutesObjs.Items {
		if staticRoute.ObjectMeta.DeletionTimestamp.IsZero() {
			staticRoutes = append(staticRoutes, staticRoute)
		}
	}
	return staticRoutes, nil
}
