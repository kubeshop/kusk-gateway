package validation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/mitchellh/copystructure"

	"github.com/kubeshop/kusk-gateway/options"
)

var (
	reAdjustSubstitutions = regexp.MustCompile(`\\(\d+)`)
)

const (
	HeaderServiceID   = "X-Kusk-Service-ID"
	HeaderOperationID = "X-Kusk-Operation-ID"
)

type operation struct {
	method string
	path   string
	op     *openapi3.Operation
}

type Service struct {
	Host string
	Port uint32

	Spec       *openapi3.T
	Opts       *options.Options
	Router     routers.Router
	Operations map[string]*operation
}

type Proxy struct {
	services map[string]*Service

	m sync.RWMutex
}

func New() *Proxy {
	return &Proxy{
		services: map[string]*Service{},
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serviceID := r.Header.Get(HeaderServiceID)

	p.m.RLock()
	service, ok := p.services[serviceID]
	p.m.RUnlock()

	if !ok {
		http.Error(w, "no such service in validation proxy", http.StatusBadGateway)
		return
	}

	operationID := r.Header.Get(HeaderOperationID)

	operation, ok := service.Operations[operationID]
	if !ok {
		http.Error(w, "no such operation in validation proxy", http.StatusBadGateway)
		return
	}

	// this is needed for validation router to find the correct route
	// host will be changed during the proxying
	r.Host = "localhost"

	if err := p.validate(r, service, operation); err != nil {
		// TODO: proper error handling
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := p.proxy(r, service, operation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.WriteHeader(resp.StatusCode)

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	_, _ = io.Copy(w, resp.Body)
	_ = resp.Body.Close()
}

func (p *Proxy) validate(r *http.Request, service *Service, operation *operation) error {
	p.m.RLock()
	defer p.m.RUnlock()

	route, pathParams, err := service.Router.FindRoute(r)
	if err != nil {
		return fmt.Errorf("failed to find route to validate against: %w", err)
	}

	return openapi3filter.ValidateRequest(r.Context(), &openapi3filter.RequestValidationInput{
		Request:     r,
		PathParams:  pathParams,
		QueryParams: nil,
		Route:       route,
		Options: &openapi3filter.Options{
			MultiError: true,
		},
	})
}

func (p *Proxy) applyRewriteOptions(r *http.Request, service *Service, operation *operation) {
	subOptions, ok := service.Opts.OperationFinalSubOptions[operation.method+operation.path]
	if ok && subOptions.Path != nil {
		if subOptions.Path.Rewrite.Pattern != "" {
			substitution := reAdjustSubstitutions.ReplaceAllString(subOptions.Path.Rewrite.Substitution, "$$1")

			r.URL.Path = regexp.MustCompile(subOptions.Path.Rewrite.Pattern).ReplaceAllString(r.URL.Path, substitution)
		}
	}

}

func (p *Proxy) proxy(r *http.Request, service *Service, operation *operation) (*http.Response, error) {
	r.RequestURI = ""
	r.Host = ""
	r.URL.Scheme = "http"
	r.URL.Host = fmt.Sprintf("%s:%d", service.Host, service.Port)

	p.applyRewriteOptions(r, service, operation)

	// TODO: proper HTTP client with timeouts
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to proxy request: %w", err)
	}

	return resp, nil
}

func (p *Proxy) Add(serviceID string, host string, port uint32, s *openapi3.T, opts *options.Options) {
	p.m.Lock()
	defer p.m.Unlock()

	specInt, err := copystructure.Copy(*s)
	if err != nil {
		panic(err)
	}

	spec := specInt.(openapi3.T)

	spec.Servers = nil

	if opts.Path != nil && opts.Path.Prefix != "" {
		spec.Servers = []*openapi3.Server{{
			URL: fmt.Sprintf("http://localhost%s", opts.Path.Prefix),
		}}
	}

	router, err := gorillamux.NewRouter(&spec)
	if err != nil {
		panic(err)
	}

	operations := map[string]*operation{}

	for path, pathItem := range spec.Paths {
		for method, op := range pathItem.Operations() {
			operations[generateOperationID(method, path)] = &operation{
				method: method,
				path:   path,
				op:     op,
			}
		}
	}

	p.services[serviceID] = &Service{
		Host:       host,
		Port:       port,
		Spec:       &spec,
		Opts:       opts,
		Router:     router,
		Operations: operations,
	}
}

func (p *Proxy) Remove(serviceID string) {
	p.m.Lock()
	defer p.m.Unlock()

	delete(p.services, serviceID)
}

// generateOperationID generates a unique, deterministic ID for a given API route,
// safe to be used in a HTTP header value.
func generateOperationID(method, path string) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%s:%s", method, path)))
	return hex.EncodeToString(hash.Sum(nil))
}
