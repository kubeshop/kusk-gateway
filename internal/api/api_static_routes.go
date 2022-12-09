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
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// StaticRoutesApiController binds http requests to an api service and writes the service results to the http response
type StaticRoutesApiController struct {
	service      StaticRoutesApiServicer
	errorHandler ErrorHandler
}

// StaticRoutesApiOption for how the controller is set up.
type StaticRoutesApiOption func(*StaticRoutesApiController)

// WithStaticRoutesApiErrorHandler inject ErrorHandler into controller
func WithStaticRoutesApiErrorHandler(h ErrorHandler) StaticRoutesApiOption {
	return func(c *StaticRoutesApiController) {
		c.errorHandler = h
	}
}

// NewStaticRoutesApiController creates a default api controller
func NewStaticRoutesApiController(s StaticRoutesApiServicer, opts ...StaticRoutesApiOption) Router {
	controller := &StaticRoutesApiController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all of the api route for the StaticRoutesApiController
func (c *StaticRoutesApiController) Routes() Routes {
	return Routes{
		{
			"GetStaticRoute",
			strings.ToUpper("Get"),
			"/staticroutes/{namespace}/{name}",
			c.GetStaticRoute,
		},
		{
			"GetStaticRouteCRD",
			strings.ToUpper("Get"),
			"/staticroutes/{namespace}/{name}/crd",
			c.GetStaticRouteCRD,
		},
		{
			"GetStaticRoutes",
			strings.ToUpper("Get"),
			"/staticroutes",
			c.GetStaticRoutes,
		},
		{
			"UpdateStaticRoute",
			strings.ToUpper("Put"),
			"/staticroutes/{namespace}/{name}",
			c.UpdateStaticRoute,
		},
	}
}

// GetStaticRoute - Get details for a single static route
func (c *StaticRoutesApiController) GetStaticRoute(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	namespaceParam := params["namespace"]

	nameParam := params["name"]

	result, err := c.service.GetStaticRoute(r.Context(), namespaceParam, nameParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// GetStaticRouteCRD - Get static route CRD
func (c *StaticRoutesApiController) GetStaticRouteCRD(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	namespaceParam := params["namespace"]

	nameParam := params["name"]

	result, err := c.service.GetStaticRouteCRD(r.Context(), namespaceParam, nameParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// GetStaticRoutes - Get a list of static routes
func (c *StaticRoutesApiController) GetStaticRoutes(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	namespaceParam := query.Get("namespace")
	result, err := c.service.GetStaticRoutes(r.Context(), namespaceParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *StaticRoutesApiController) UpdateStaticRoute(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	namespaceParam := params["namespace"]
	nameParam := params["name"]

	existingStaticRoute, err := c.service.GetStaticRoute(r.Context(), namespaceParam, nameParam)
	if err != nil {
		c.errorHandler(w, r, err, &existingStaticRoute)
		return
	}

	staticRouteItem, err := decodeBodyToInlineObject(r.Body)
	if err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}

	staticRouteItem.Name = nameParam
	staticRouteItem.Namespace = namespaceParam

	result, err := c.service.UpdateStaticRoute(r.Context(), staticRouteItem)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}
