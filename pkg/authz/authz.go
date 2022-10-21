package authz

import (
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/kubeshop/kusk-gateway/pkg/cloudentity"
)

type AuthorizationServer struct {
	log logr.Logger
}

func (a *AuthorizationServer) check(w http.ResponseWriter, r *http.Request) {
	url := r.Header.Get(cloudentity.HeaderAuthorizerURL)
	if url == "" {
		a.log.Info("request missing authorizer url header", "request", r)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	apiGroup := r.Header.Get(cloudentity.HeaderAPIGroup)
	if apiGroup == "" {
		a.log.Info("request missing api group header", "request", r)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cl := cloudentity.New(url)

	valReq := &cloudentity.ValidateRequest{
		APIGroup:    apiGroup,
		Method:      r.Method,
		Path:        r.URL.Path,
		QueryParams: r.URL.Query(),
		Headers:     r.Header,
	}

	err := cl.Validate(r.Context(), valReq)
	if err != nil {
		a.log.Info("cloudentity validate", "error", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusOK)
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
