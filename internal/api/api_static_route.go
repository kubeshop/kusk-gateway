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

// StaticRouteApiController binds http requests to an api service and writes the service results to the http response
type StaticRouteApiController struct {
	service      StaticRouteApiServicer
	errorHandler ErrorHandler
}

// StaticRouteApiOption for how the controller is set up.
type StaticRouteApiOption func(*StaticRouteApiController)

// WithStaticRouteApiErrorHandler inject ErrorHandler into controller
func WithStaticRouteApiErrorHandler(h ErrorHandler) StaticRouteApiOption {
	return func(c *StaticRouteApiController) {
		c.errorHandler = h
	}
}

// NewStaticRouteApiController creates a default api controller
func NewStaticRouteApiController(s StaticRouteApiServicer, opts ...StaticRouteApiOption) Router {
	controller := &StaticRouteApiController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all of the api route for the StaticRouteApiController
func (c *StaticRouteApiController) Routes() Routes {
	return Routes{
		{
			"DeleteStaticRoute",
			strings.ToUpper("Delete"),
			"/staticroutes/{namespace}/{name}",
			c.DeleteStaticRoute,
		},
	}
}

// DeleteStaticRoute - Delete a StaticRoute by namespace and name
func (c *StaticRouteApiController) DeleteStaticRoute(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	namespaceParam := params["namespace"]

	nameParam := params["name"]

	result, err := c.service.DeleteStaticRoute(r.Context(), namespaceParam, nameParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}
