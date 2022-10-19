/*
MIT License

Copyright (c) 2022 Kubeshop

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
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"

	"github.com/kubeshop/kusk-gateway/pkg/options"
)

func TestGetOptions(t *testing.T) {
	trueValue := true

	var testCases = []struct {
		name string
		spec *openapi3.T
		res  options.Options
		err  bool
	}{
		{
			name: "no extensions",
			spec: &openapi3.T{},
			res: options.Options{
				OperationFinalSubOptions: make(map[string]options.SubOptions),
			},
		},
		{
			name: "path level options set",
			spec: &openapi3.T{
				Paths: openapi3.Paths{
					"/pet": &openapi3.PathItem{
						ExtensionProps: openapi3.ExtensionProps{
							Extensions: map[string]interface{}{
								kuskExtensionKey: json.RawMessage(`{"disabled":true}`),
							},
						},
						Get: &openapi3.Operation{},
					},
				},
			},
			res: options.Options{
				OperationFinalSubOptions: map[string]options.SubOptions{
					"GET/pet": {
						Disabled: &trueValue,
					},
				},
			},
		},
		{
			name: "operation level options set",
			spec: &openapi3.T{
				Paths: openapi3.Paths{
					"/pet": &openapi3.PathItem{
						Put: &openapi3.Operation{
							ExtensionProps: openapi3.ExtensionProps{
								Extensions: map[string]interface{}{
									kuskExtensionKey: json.RawMessage(`{"disabled":true}`),
								},
							},
						},
					},
				},
			},
			res: options.Options{
				OperationFinalSubOptions: map[string]options.SubOptions{
					"PUT/pet": {
						Disabled: &trueValue,
					},
				},
			},
		},
		{
			name: "operation level options set",
			spec: &openapi3.T{
				Paths: openapi3.Paths{
					"/pet": &openapi3.PathItem{
						Put: &openapi3.Operation{
							ExtensionProps: openapi3.ExtensionProps{
								Extensions: map[string]interface{}{
									kuskExtensionKey: json.RawMessage(`{"enabled":true}`),
								},
							},
						},
					},
				},
			},
			err: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := require.New(t)

			actual, err := GetOptions(testCase.spec)
			if testCase.err {
				r.True(err != nil, "expected error")
			} else {
				r.Equal(testCase.res, *actual)
			}
		})
	}

}
