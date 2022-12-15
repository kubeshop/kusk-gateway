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
	"net/url"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

const (
	loadedFromURI  = "loaded from URI"
	loadedFromFile = "loaded from file"
)

type mockLoader struct{}

func (m mockLoader) LoadFromURI(_ *url.URL) (*openapi3.T, error) {
	return &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:       "Sample API",
			Description: loadedFromURI,
			Version:     "1.0.0",
		},
	}, nil
}

func (m mockLoader) LoadFromFile(_ string) (*openapi3.T, error) {
	return &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:       "Sample API",
			Description: loadedFromFile,
			Version:     "1.0.0",
		},
	}, nil
}

func TestParse(t *testing.T) {
	testCases := []struct {
		name   string
		url    string
		result string
	}{
		{
			name:   "load spec from url",
			url:    "https://someurl.io/swagger.yaml",
			result: loadedFromURI,
		},
		{
			name:   "load spec from local file",
			url:    "some-folder/swagger.yaml",
			result: loadedFromFile,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := require.New(t)

			parser := Parser{loader: mockLoader{}}
			u, err := url.Parse(testCase.url)
			r.NoError(err, "please provide a valid url")

			actual, err := parser.Parse(u.String())
			r.NoError(err, "expected no error when running parse from mocked loader")
			r.True(actual.Info.Description == testCase.result)
		})
	}
}

func TestParseFromReader(t *testing.T) {
	testCases := []struct {
		name   string
		spec   string
		result *openapi3.T
	}{
		{
			name: "swagger",
			spec: `swagger: "2.0"
info:
  title: Sample API
  description: API description in Markdown.
  version: 1.0.0
paths:
  /users:
    get: {}
`,
			result: &openapi3.T{
				OpenAPI: "3.0.3",
				Info: &openapi3.Info{
					Title:       "Sample API",
					Description: "API description in Markdown.",
					Version:     "1.0.0",
				},
				Paths: openapi3.Paths{
					"/users": &openapi3.PathItem{
						Get: &openapi3.Operation{},
					},
				},
			},
		},
		{
			name: "openapi",
			spec: `openapi: "3.0.3"
info:
  title: Sample API
  description: API description in Markdown.
  version: 1.0.0
paths:
  /users:
    get: {}
`,
			result: &openapi3.T{
				OpenAPI: "3.0.3",
				Info: &openapi3.Info{
					Title:       "Sample API",
					Description: "API description in Markdown.",
					Version:     "1.0.0",
				},
				Paths: openapi3.Paths{
					"/users": &openapi3.PathItem{
						Get: &openapi3.Operation{},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := require.New(t)

			actual, err := Parser{loader: openapi3.NewLoader()}.ParseFromReader(strings.NewReader(testCase.spec))
			r.NoError(err, "failed to parse spec from reader")
			r.Equal(testCase.result.OpenAPI, actual.OpenAPI)
			r.Equal(testCase.result.Info.Title, actual.Info.Title)
			r.Equal(testCase.result.Info.Description, actual.Info.Description)
			r.Equal(testCase.result.Info.Version, actual.Info.Version)
			r.NotNil(testCase.result.Paths.Find("/users"))
		})

	}
}
