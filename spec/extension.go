package spec

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"

	"github.com/kubeshop/kusk-gateway/options"
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

// GetOptions would retrieve and parse x-kusk top-level OpenAPI extension
// that contains Kusk options. If there's no extension found, an empty object will be returned.
func GetOptions(spec *openapi3.T) (*options.Options, error) {
	var res options.Options

	if _, err := parseExtension(&spec.ExtensionProps, &res); err != nil {
		return nil, err
	}

	for path, pathItem := range spec.Paths {
		pathSubOptions, ok, err := getPathOptions(pathItem)
		if err != nil {
			return nil, fmt.Errorf("failed to extract path suboptions: %w", err)
		}

		if ok {
			if res.PathSubOptions == nil {
				res.PathSubOptions = map[string]options.SubOptions{}
			}

			res.PathSubOptions[path] = pathSubOptions
		}

		for method, operation := range pathItem.Operations() {
			operationSubOptions, ok, err := getOperationOptions(operation)
			if err != nil {
				return nil, fmt.Errorf("failed to extract operation suboptions: %w", err)
			}

			if ok {
				if res.OperationSubOptions == nil {
					res.OperationSubOptions = map[string]options.SubOptions{}
				}

				res.OperationSubOptions[method+path] = operationSubOptions
			}
		}
	}

	return &res, nil
}

func parseExtension(extensionProps *openapi3.ExtensionProps, target interface{}) (bool, error) {
	if extension, ok := extensionProps.Extensions[kuskExtensionKey]; ok {
		if kuskExtension, ok := extension.(json.RawMessage); ok {
			err := yaml.Unmarshal(kuskExtension, target)
			if err != nil {
				return false, fmt.Errorf("failed to parse extension: %w", err)
			}

			return true, nil
		}
	}

	return false, nil
}
