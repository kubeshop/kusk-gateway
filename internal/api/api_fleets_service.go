/*
 * Kusk Gateway API
 *
 * This is the Kusk Gateway Management API
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package api

import (
	"context"
	"net/http"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/pkg/analytics"

	"github.com/kubeshop/kusk-gateway/pkg/kusk"
)

// FleetsApiService is a service that implements the logic for the FleetsApiServicer
// This service should implement the business logic for every endpoint for the FleetsApi API.
// Include any external packages or services that will be required by this service.
type FleetsApiService struct {
	kuskClient kusk.Client
}

// NewFleetsApiService creates a default api service
func NewFleetsApiService(kuskClient kusk.Client) FleetsApiServicer {
	return &FleetsApiService{
		kuskClient: kuskClient,
	}
}

func (s *FleetsApiService) DeleteFleet(ctx context.Context, namespace string, name string) (ImplResponse, error) {
	_ = analytics.SendAnonymousInfo(ctx, s.kuskClient.K8sClient(), "kusk-api-server", "DeleteFleet")

	if err := s.kuskClient.DeleteFleet(v1alpha1.EnvoyFleet{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}); err != nil {
		return Response(http.StatusInternalServerError, err), err
	}
	return Response(http.StatusOK, nil), nil
}

// GetEnvoyFleet - Get details for a single envoy fleet
func (s *FleetsApiService) GetEnvoyFleet(ctx context.Context, namespace string, name string) (ImplResponse, error) {
	_ = analytics.SendAnonymousInfo(ctx, s.kuskClient.K8sClient(), "kusk-api-server", "GetEnvoyFleet")
	fleet, err := s.kuskClient.GetEnvoyFleet(namespace, name)
	if err != nil {
		return GetResponseFromK8sError(err), err
	}

	return Response(http.StatusOK, s.convertEnvoyFleetCRDtoEnvoyFleetModel(fleet)), nil
}

func (s *FleetsApiService) GetEnvoyFleetCRD(ctx context.Context, namespace string, name string) (ImplResponse, error) {
	_ = analytics.SendAnonymousInfo(ctx, s.kuskClient.K8sClient(), "kusk-api-server", "GetEnvoyFleetCRD")
	fleet, err := s.kuskClient.GetEnvoyFleet(namespace, name)
	if err != nil {
		return GetResponseFromK8sError(err), err
	}

	return Response(http.StatusOK, fleet), nil
}

// GetEnvoyFleets - Get a list of envoy fleets
func (s *FleetsApiService) GetEnvoyFleets(ctx context.Context, namespace string) (ImplResponse, error) {
	_ = analytics.SendAnonymousInfo(ctx, s.kuskClient.K8sClient(), "kusk-api-server", "GetEnvoyFleets")
	fleets, err := s.kuskClient.GetEnvoyFleets()
	if err != nil {
		return GetResponseFromK8sError(err), err
	}
	return Response(http.StatusOK, s.convertEnvoyFleetListCRDtoEnvoyFleetsModel(fleets)), nil
}

func (s *FleetsApiService) convertEnvoyFleetListCRDtoEnvoyFleetsModel(fleets *v1alpha1.EnvoyFleetList) []EnvoyFleetItem {
	toReturn := []EnvoyFleetItem{}
	for _, fleet := range fleets.Items {
		toReturn = append(toReturn, s.convertEnvoyFleetCRDtoEnvoyFleetModel(&fleet))
	}
	return toReturn
}

func (s *FleetsApiService) convertEnvoyFleetCRDtoEnvoyFleetModel(fleet *v1alpha1.EnvoyFleet) EnvoyFleetItem {
	apifs := []ApiItemFleet{}
	apis, err := s.kuskClient.GetApiByEnvoyFleet("", fleet.Namespace, fleet.Name)
	if err == nil {
		for _, api := range apis.Items {
			apifs = append(apifs, ApiItemFleet{
				Name:      api.Name,
				Namespace: api.Namespace,
			})
		}
	}
	srs := []StaticRouteItemFleet{}
	staticRoutes, err := s.kuskClient.GetStaticRoutes("")
	if err == nil {
		for _, sr := range staticRoutes.Items {
			if sr.Spec.Fleet.Name == fleet.Name && sr.Spec.Fleet.Namespace == fleet.Namespace {
				srs = append(srs, StaticRouteItemFleet{
					Name:      sr.Name,
					Namespace: sr.Namespace,
				})
			}
		}
	}

	return EnvoyFleetItem{
		Name:         fleet.Name,
		Namespace:    fleet.Namespace,
		Apis:         apifs,
		StaticRoutes: srs,
	}
}
