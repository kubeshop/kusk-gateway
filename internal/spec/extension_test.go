package spec

import (
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"

	"github.com/kubeshop/kusk-gateway/internal/options"
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
