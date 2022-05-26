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
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"sigs.k8s.io/yaml"

	"github.com/kubeshop/kusk-gateway/pkg/options"
)

const kuskExtensionKey = "x-kusk"

func getPathOptions(path *openapi3.PathItem) (options.SubOptions, bool, error) {
	var res options.SubOptions

	ok, err := parseExtension(&path.ExtensionProps, &res)

	return res, ok, err
}

func getOperationOptions(operation *openapi3.Operation) (options.SubOptions, bool, error) {
	var res options.SubOptions

	ok, err := parseExtension(&operation.ExtensionProps, &res)

	return res, ok, err
}

// GetOptions would retrieve and parse x-kusk OpenAPI extension
// that contains Kusk options. If there's no extension found, an empty object will be returned.
// For each found method in the document top and path level x-kusk options will be merged in
// to form OperationFinalSubOptions map that has the complete configuration for each method.
func GetOptions(spec *openapi3.T) (*options.Options, error) {
	res := options.Options{
		OperationFinalSubOptions: make(map[string]options.SubOptions),
	}

	if _, err := parseExtension(&spec.ExtensionProps, &res); err != nil {
		return nil, err
	}

	for path, pathItem := range spec.Paths {
		pathSubOptions, _, err := getPathOptions(pathItem)
		if err != nil {
			return nil, fmt.Errorf("failed to extract path suboptions: %w", err)
		}

		// Merge in top level.
		pathSubOptions.MergeInSubOptions(&res.SubOptions)
		for method, operation := range pathItem.Operations() {
			operationSubOptions, _, err := getOperationOptions(operation)
			if err != nil {
				return nil, fmt.Errorf("failed to extract operation suboptions: %w", err)
			}

			// Merged in path
			operationSubOptions.MergeInSubOptions(&pathSubOptions)
			res.OperationFinalSubOptions[method+path] = operationSubOptions
		}
	}

	return &res, nil
}

func parseExtension(extensionProps *openapi3.ExtensionProps, target interface{}) (bool, error) {
	if extension, ok := extensionProps.Extensions[kuskExtensionKey]; ok {
		if kuskExtension, ok := extension.(json.RawMessage); ok {
			err := yaml.UnmarshalStrict(kuskExtension, target)
			if err != nil {
				return false, fmt.Errorf("failed to parse x-kusk='%s' extension: %w", string(kuskExtension), err)
			}

			return true, nil
		}
	}

	return false, nil
}
