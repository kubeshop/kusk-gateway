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
	"math/rand"
	"net/url"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/samber/lo"
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
func GetExampleResponse(mediaType *openapi3.MediaType, logger logr.Logger) interface{} {
	logger = logger.WithName("GetExampleResponse")

	if mediaType == nil {
		logger.Info("mediaType is nill ignoring example response")
		return nil
	}

	logger.Info("using `Examples`, if present", "mediaType.Examples", spew.Sprint(mediaType.Examples))

	if mediaType.Examples != nil {
		totalExamples := len(mediaType.Examples)
		examplesTried := map[int]bool{}
		rand := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

		// Creates an array of the map values.
		// See: https://github.com/samber/lo#values
		values := lo.Values[string, *openapi3.ExampleRef](mediaType.Examples)

		for {
			exampleIndex := rand.Intn(totalExamples)

			triedAllExamples := true
			// Generate range from [0, totalExamples)
			// See: https://github.com/samber/lo#range--rangefrom--rangewithsteps
			for index := range lo.Range(totalExamples) {
				if _, ok := examplesTried[index]; !ok {
					triedAllExamples = false
				}
			}

			if triedAllExamples {
				break
			}

			example := values[exampleIndex]
			logger.Info("`mediaType.Examples`", "exampleIndex", exampleIndex, "example.Value", spew.Sprint(example.Value))
			if example.Value != nil && example.Value.Value != nil {
				return example.Value.Value
			} else {
				examplesTried[exampleIndex] = true
			}
		}
	}

	logger.Info("using if `Example`, if present", "mediaType.Example", spew.Sprint(mediaType.Example))

	if mediaType.Example != nil {
		return mediaType.Example
	}

	return nil
}
