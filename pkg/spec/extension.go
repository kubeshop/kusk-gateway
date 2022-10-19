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
				return false, fmt.Errorf("failed to parse extension: %w. Check the extensions supported by Kusk at  https://docs.kusk.io/extension", err)
			}

			return true, nil
		}
	}

	return false, nil
}

func PostProcessedDef(apiSpec openapi3.T, opt options.Options) *openapi3.T {
	postProcessed := apiSpec
	postProcessed.Paths = openapi3.Paths{}
	delete(postProcessed.ExtensionProps.Extensions, kuskExtensionKey)

	for path, pathItem := range apiSpec.Paths {
		pathOptions, _, _ := getPathOptions(pathItem)
		for method := range pathItem.Operations() {
			if pathOptions.Hidden != nil && *pathOptions.Hidden {
				item := &openapi3.PathItem{}
				fOpt := opt.OperationFinalSubOptions[method+path]
				if fOpt.Hidden != nil && !*fOpt.Hidden {
					if pathOptions.Hidden != nil && *pathOptions.Hidden {
						if item = parsePathItem(pathItem); len(item.Operations()) > 0 {
							postProcessed.Paths[path] = item
						}
					}
				}
			} else {
				delete(pathItem.ExtensionProps.Extensions, kuskExtensionKey)
				postProcessed.Paths[path] = pathItem
			}
		}
	}

	return &postProcessed
}

func parsePathItem(pathItem *openapi3.PathItem) (result *openapi3.PathItem) {
	result = pathItem
	delete(result.ExtensionProps.Extensions, kuskExtensionKey)
	for operation, oper := range pathItem.Operations() {
		opts, _, _ := getOperationOptions(oper)
		if opts.Hidden != nil && !*opts.Hidden {
			delete(oper.ExtensionProps.Extensions, kuskExtensionKey)
			result.SetOperation(operation, oper)
		} else if opts.Hidden != nil && *opts.Hidden {
			result.SetOperation(operation, nil)
		} else if opts.Hidden == nil {
			result.SetOperation(operation, nil)
		}

	}
	return result
}
