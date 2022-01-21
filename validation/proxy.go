package validation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

// operation holds original route parameters from spec
// it is used to quickly access route parameters by OperationID (extracted from HeaderOperationID header)
type operation struct {
	method string
	path   string
	op     *openapi3.Operation
}

type Service struct {
	ID string

	Host string
	Port uint32

	Spec       *openapi3.T
	Opts       *options.Options
	Router     routers.Router
	Operations map[string]*operation
}

func NewService(id string, host string, port uint32, s *openapi3.T, opts *options.Options) (*Service, error) {
	specInt, err := copystructure.Copy(*s)
	if err != nil {
		return nil, fmt.Errorf("failed to copy spec: %w", err)
	}

	spec := specInt.(openapi3.T)

	// we will always use localhost and attach prefix, if any
	spec.Servers = nil
	if opts.Path != nil && opts.Path.Prefix != "" {
		spec.Servers = []*openapi3.Server{{
			URL: fmt.Sprintf("http://localhost%s", opts.Path.Prefix),
		}}
	}

	router, err := gorillamux.NewRouter(&spec)
	if err != nil {
		if err != nil {
			return nil, fmt.Errorf("failed to create router: %w", err)
		}
	}

	operations := map[string]*operation{}

	for path, pathItem := range spec.Paths {
		for method, op := range pathItem.Operations() {
			operations[GenerateOperationID(method, path)] = &operation{
				method: method,
				path:   path,
				op:     op,
			}
		}
	}

	return &Service{
		ID:         id,
		Host:       host,
		Port:       port,
		Spec:       &spec,
		Opts:       opts,
		Router:     router,
		Operations: operations,
	}, nil
}

type Proxy struct {
	services map[string]*Service

	m sync.RWMutex
}

func NewProxy() *Proxy {
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
		w.WriteHeader(http.StatusBadRequest)

		if multiError, ok := err.(openapi3.MultiError); ok {
			errs := make([]string, len(multiError))
			for i := range multiError {
				errs[i] = multiError[i].Error()
			}

			_ = json.NewEncoder(w).Encode(Error{errs})
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		return err
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

func (p *Proxy) UpdateServices(services []*Service) {
	p.m.Lock()
	defer p.m.Unlock()

	// rebuild the services map
	p.services = make(map[string]*Service, len(services))

	for _, service := range services {
		p.services[service.ID] = service
	}
}

// GenerateOperationID generates a unique, deterministic ID for a given API route,
// safe to be used in a HTTP header value.
func GenerateOperationID(method, path string) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%s:%s", method, path)))
	return hex.EncodeToString(hash.Sum(nil))
}

// GenerateServiceID generates a unique, deterministic ID for a given API service,
// safe to be used in a HTTP header value.
func GenerateServiceID(hostname string, port uint32) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%s:%d", hostname, port)))
	return hex.EncodeToString(hash.Sum(nil))
}
