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
package spec

import (
	"bytes"
	"fmt"
	"io"
	"net/url"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
)

// isSwagger tries to decode the spec header
func isSwagger(spec []byte) bool {
	// internal agent struct to help us differentiate
	// between openapi spec 2.0 (swagger) and openapi 3+
	var header struct {
		Swagger string `json:"swagger"`
		OpenAPI string `json:"openapi"` // we might need that later to distinguish 3.1.x vs 3.0.x
	}

	_ = yaml.Unmarshal(spec, &header)

	return header.Swagger != ""
}

func isOpenAPI(spec []byte) bool {
	// internal agent struct to help us differentiate
	// between openapi spec 2.0 (swagger) and openapi 3+
	var header struct {
		Swagger string `json:"swagger"`
		OpenAPI string `json:"openapi"` // we might need that later to distinguish 3.1.x vs 3.0.x
	}

	_ = yaml.Unmarshal(spec, &header)

	return header.OpenAPI != ""
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
	spec, err := io.ReadAll(contents)
	if err != nil {
		return nil, fmt.Errorf("could not read contents of api spec: %w", err)
	}

	if isSwagger(spec) {
		return parseSwagger(spec)
	}
	if isOpenAPI(spec) {
		return parseOpenAPI3(spec)
	}
	return nil, fmt.Errorf("provided specs are not OpenAPI/Swagger specs")
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
	return (&openapi3.Loader{IsExternalRefsAllowed: true}).LoadFromData(spec)
}

// GetExampleResponse returns a single example response from the given operation
// if one exists.
func GetExampleResponse(mediaType *openapi3.MediaType) interface{} {
	if mediaType == nil {
		return nil
	}

	if mediaType.Example != nil {
		return mediaType.Example
	}

	// https://github.com/kubeshop/kusk-gateway/issues/298 and https://github.com/kubeshop/kusk-gateway/issues/324
	//
	// Certain examples, like the one below, parses this structure
	//
	// application/json:
	// 	schema:
	// 		type: object
	// 		properties:
	// 			order:
	// 				type: integer
	// 				format: int32
	// 			completed:
	// 				type: boolean
	// 		required:
	// 			- order
	// 			- completed
	// 		example:
	// 			order: 13
	// 			completed: true
	//
	// With the `example` in the `mediaType.Schema.Value.Example` field.
	if mediaType.Schema != nil {
		if mediaType.Schema.Value != nil {
			if mediaType.Schema.Value.Example != nil {
				return mediaType.Schema.Value.Example
			}
		}
	}

	if mediaType.Examples != nil {
		for _, example := range mediaType.Examples {
			if example.Value != nil && example.Value.Value != nil {
				return example.Value.Value
			}
		}
	}

	return nil
}
