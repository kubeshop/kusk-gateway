package spec

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
)

// isSwagger tries to decode the spec header
func isSwagger(spec []byte) bool {
	// internal helper struct to help us differentiate
	// between openapi spec 2.0 (swagger) and openapi 3+
	var header struct {
		Swagger string `json:"swagger"`
		OpenAPI string `json:"openapi"` // we might need that later to distinguish 3.1.x vs 3.0.x
	}

	_ = yaml.Unmarshal(spec, &header)

	return header.Swagger != ""
}

type Loader interface {
	LoadFromURI(location *url.URL) (*openapi3.T, error)
	LoadFromFile(location string) (*openapi3.T, error)
}

type Parser struct {
	loader Loader
}

func NewParser(loader Loader) Parser {
	return Parser{
		loader: loader,
	}
}

// Parse is the entrypoint for the spec package
// Accepts a path that should be parseable into a resource locater
// i.e. a URL or relative file path
func (p Parser) Parse(path string) (*openapi3.T, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid resource path %s: %w", path, err)
	}

	var spec *openapi3.T
	if isURLRelative := u.Host == ""; isURLRelative {
		spec, err = p.loader.LoadFromFile(path)
	} else {
		spec, err = p.loader.LoadFromURI(u)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to load spec: %w", err)
	}

	// we need to marshal the struct back to yaml while we support
	// both openapi spec 2.0 and 3.0, so we can differentiate between the two
	// and convert 2.0 to 3.0 if needed
	bSpec, err := yaml.Marshal(&spec)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal spec to yaml: %w", err)
	}

	return p.ParseFromReader(bytes.NewReader(bSpec))
}

// ParseFromReader allows for providing your own Reader implementation
// to parse the API spec from
func (p Parser) ParseFromReader(contents io.Reader) (*openapi3.T, error) {
	spec, err := ioutil.ReadAll(contents)
	if err != nil {
		return nil, fmt.Errorf("could not read contents of api spec: %w", err)
	}

	if isSwagger(spec) {
		return parseSwagger(spec)
	}

	return parseOpenAPI3(spec)
}

func parseSwagger(spec []byte) (*openapi3.T, error) {
	spec, err := yaml.YAMLToJSON(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to convert YAML to JSON: %w", err)
	}

	var swaggerSpec openapi2.T

	err = swaggerSpec.UnmarshalJSON(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Swagger: %w", err)
	}

	return openapi2conv.ToV3(&swaggerSpec)
}

func parseOpenAPI3(spec []byte) (*openapi3.T, error) {
	return openapi3.NewLoader().LoadFromData(spec)
}
