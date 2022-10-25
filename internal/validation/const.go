/*
MIT License

# Copyright (c) 2022 Kubeshop

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
package validation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/mitchellh/copystructure"

	"github.com/kubeshop/kusk-gateway/pkg/options"
)

var (
	reAdjustSubstitutions = regexp.MustCompile(`\\(\d+)`)
)

const (
	HeaderServiceID     = "X-Kusk-Service-ID"
	HeaderOperationID   = "X-Kusk-Operation-ID"
	HeaderOperationName = "X-Kusk-Operation-Name"
)

// ValidationUpdater adds and updates Services to the validation service
type ValidationUpdater interface {
	UpdateServices(services []*Service)
}

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
