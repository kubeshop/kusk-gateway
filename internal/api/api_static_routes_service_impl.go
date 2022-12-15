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

	"github.com/kubeshop/kusk-gateway/pkg/analytics"

	"github.com/kubeshop/kusk-gateway/pkg/kusk"
)

// StaticRoutesApiService is a service that implements the logic for the StaticRoutesApiServicer
// This service should implement the business logic for every endpoint for the StaticRoutesApi API.
// Include any external packages or services that will be required by this service.
type StaticRoutesApiService struct {
	kuskClient kusk.Client
}

// NewStaticRoutesApiService creates a default api service
func NewStaticRoutesApiService(kuskClient kusk.Client) StaticRoutesApiServicer {
	return &StaticRoutesApiService{kuskClient: kuskClient}
}

// GetStaticRoute - Get details for a single static route
func (s *StaticRoutesApiService) GetStaticRoute(ctx context.Context, namespace string, name string) (ImplResponse, error) {
	_ = analytics.SendAnonymousInfo(ctx, s.kuskClient.K8sClient(), "kusk-api-server", "GetStaticRoute")
	staticRoute, err := s.kuskClient.GetStaticRoute(namespace, name)
	if err != nil {
		return GetResponseFromK8sError(err), err
	}
	return Response(http.StatusOK, StaticRouteItem{
		Name:      staticRoute.Name,
		Namespace: staticRoute.Namespace,
	}), nil
}

// GetStaticRouteCRD - Get static route CRD
func (s *StaticRoutesApiService) GetStaticRouteCRD(ctx context.Context, namespace string, name string) (ImplResponse, error) {
	_ = analytics.SendAnonymousInfo(ctx, s.kuskClient.K8sClient(), "kusk-api-server", "GetStaticRouteCRD")
	staticRoute, err := s.kuskClient.GetStaticRoute(namespace, name)
	if err != nil {
		return GetResponseFromK8sError(err), err
	}

	return Response(http.StatusOK, staticRoute), nil
}

// GetStaticRoutes - Get a list of static routes
func (s *StaticRoutesApiService) GetStaticRoutes(ctx context.Context, namespace string) (ImplResponse, error) {
	_ = analytics.SendAnonymousInfo(ctx, s.kuskClient.K8sClient(), "kusk-api-server", "GetStaticRoutes")
	staticRoutes, err := s.kuskClient.GetStaticRoutes(namespace)
	if err != nil {
		return Response(http.StatusInternalServerError, err), err
	}

	toReturn := []StaticRouteItem{}
	for _, sr := range staticRoutes.Items {
		toReturn = append(toReturn, StaticRouteItem{
			Name:      sr.Name,
			Namespace: sr.Namespace,
		})
	}
	return Response(http.StatusOK, toReturn), nil
}

func (s *StaticRoutesApiService) UpdateStaticRoute(ctx context.Context, staticRoute InlineObject) (ImplResponse, error) {
	_ = analytics.SendAnonymousInfo(ctx, s.kuskClient.K8sClient(), "kusk-api-server", "UpdateStaticRoute")
	updatedStaticRoute, err := s.kuskClient.UpdateStaticRoute(staticRoute.Namespace, staticRoute.Name, staticRoute.EnvoyFleetName, staticRoute.EnvoyFleetNamespace, staticRoute.Openapi)
	if err != nil {
		return GetResponseFromK8sError(err), err
	}

	toReturn := StaticRouteItem{
		Name:                updatedStaticRoute.Name,
		Namespace:           updatedStaticRoute.Namespace,
		EnvoyFleetName:      updatedStaticRoute.Spec.Fleet.Name,
		EnvoyFleetNamespace: updatedStaticRoute.Spec.Fleet.Namespace,
	}
	return Response(http.StatusCreated, toReturn), nil
}