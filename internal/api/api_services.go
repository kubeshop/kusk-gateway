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

// ServicesApiController binds http requests to an api service and writes the service results to the http response
type ServicesApiController struct {
	service      ServicesApiServicer
	errorHandler ErrorHandler
}

// ServicesApiOption for how the controller is set up.
type ServicesApiOption func(*ServicesApiController)

// WithServicesApiErrorHandler inject ErrorHandler into controller
func WithServicesApiErrorHandler(h ErrorHandler) ServicesApiOption {
	return func(c *ServicesApiController) {
		c.errorHandler = h
	}
}

// NewServicesApiController creates a default api controller
func NewServicesApiController(s ServicesApiServicer, opts ...ServicesApiOption) Router {
	controller := &ServicesApiController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all of the api route for the ServicesApiController
func (c *ServicesApiController) Routes() Routes {
	return Routes{
		{
			"GetService",
			strings.ToUpper("Get"),
			"/services/{namespace}/{name}",
			c.GetService,
		},
		{
			"GetServices",
			strings.ToUpper("Get"),
			"/services",
			c.GetServices,
		},
	}
}

// GetService - Get details for a single service
func (c *ServicesApiController) GetService(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	namespaceParam := params["namespace"]

	nameParam := params["name"]

	result, err := c.service.GetService(r.Context(), namespaceParam, nameParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// GetServices - Get a list of services handled by kusk-gateway
func (c *ServicesApiController) GetServices(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	namespaceParam := query.Get("namespace")
	result, err := c.service.GetServices(r.Context(), namespaceParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}
