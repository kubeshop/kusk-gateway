package authz

import (
	"fmt"
	"net/http"

	"github.com/go-logr/logr"

	"github.com/kubeshop/kusk-gateway/internal/cloudentity"
)

type AuthorizationServer struct {
	log logr.Logger
}

func (a *AuthorizationServer) check(writer http.ResponseWriter, request *http.Request) {
	url := request.Header.Get(cloudentity.HeaderAuthorizerURL)
	if url == "" {
		a.log.Info("request missing authorizer url header", "request", request)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	apiGroup := request.Header.Get(cloudentity.HeaderAPIGroup)
	if apiGroup == "" {
		a.log.Info("request missing api group header", "request", request)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	cl := cloudentity.New(url)

	valReq := &cloudentity.ValidateRequest{
		APIGroup:    apiGroup,
		Method:      request.Method,
		Path:        request.URL.Path,
		QueryParams: request.URL.Query(),
		Headers:     request.Header,
	}

	err := cl.Validate(request.Context(), valReq)
	if err != nil {
		a.log.Info("cloudentity validate", "error", err)
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func NewServer(log logr.Logger) *AuthorizationServer {
	return &AuthorizationServer{log: log}
}

func (a *AuthorizationServer) ListenAndServe(address string) error {
	a.log.Info("authz listening on", "address", address)
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.check)

	err := http.ListenAndServe(address, mux)
	if err != nil {
		return fmt.Errorf("authz listen and serve: %w", err)
	}
	return nil
}
