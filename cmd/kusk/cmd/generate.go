/*
The MIT License (MIT)

# Copyright © 2022 Kubeshop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"os"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/templates"
	"github.com/kubeshop/kusk-gateway/pkg/options"
	"github.com/kubeshop/kusk-gateway/pkg/spec"
)

var (
	apiTemplate *template.Template
	apiSpecPath string

	name      string
	namespace string
	apiSpec   string

	serviceName      string
	serviceNamespace string
	servicePort      uint32

	envoyFleetName      string
	envoyFleetNamespace string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:           "generate",
	Short:         "Generate a Kusk Gateway API resource from your OpenAPI spec file",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		reportError := func(err error) {
			if err != nil {
				errors.NewErrorReporter(cmd, err).Report()
			}
		}

		parsedApiSpec, err := spec.NewParser(&openapi3.Loader{IsExternalRefsAllowed: true}).Parse(apiSpecPath)
		if err != nil {
			reportError(err)
			return err
		}

		if _, ok := parsedApiSpec.ExtensionProps.Extensions["x-kusk"]; !ok {
			parsedApiSpec.ExtensionProps.Extensions["x-kusk"] = options.Options{}
		}

		// if name flag is not defined, use the swagger doc title which is guarunteed to be there
		if name == "" {
			// kubernetes manifests cannot have . in the name so replace them
			name = strings.ReplaceAll(parsedApiSpec.Info.Title, ".", "-")
		}

		opts, err := spec.GetOptions(parsedApiSpec)
		if err != nil {
			reportError(err)
			return err
		}

		if err := opts.Validate(); err != nil {
			reportError(err)
			return err
		}

		// override top level upstream service if undefined.
		if serviceName != "" && serviceNamespace != "" && servicePort != 0 {
			xKusk := parsedApiSpec.ExtensionProps.Extensions["x-kusk"].(options.Options)
			xKusk.Upstream = &options.UpstreamOptions{
				Service: &options.UpstreamService{
					Name:      serviceName,
					Namespace: serviceNamespace,
					Port:      servicePort,
				},
			}

			parsedApiSpec.ExtensionProps.Extensions["x-kusk"] = xKusk
		}

		if err := validateExtensionOptions(parsedApiSpec.ExtensionProps.Extensions["x-kusk"]); err != nil {
			reportError(err)
			return err
		}

		if apiSpec, err = getAPISpecString(parsedApiSpec); err != nil {
			reportError(err)
			return err
		}

		if err := apiTemplate.Execute(os.Stdout, templates.APITemplateArgs{
			Name:                name,
			Namespace:           namespace,
			EnvoyfleetName:      envoyFleetName,
			EnvoyfleetNamespace: envoyFleetNamespace,
			Spec:                strings.Split(apiSpec, "\n"),
		}); err != nil {
			reportError(err)
			return err
		}
		return nil
	},
}

func validateExtensionOptions(extension interface{}) error {
	b, err := yaml.Marshal(extension)
	if err != nil {
		return err
	}

	var o options.Options
	if err := yaml.Unmarshal(b, &o); err != nil {
		return err
	}

	o.FillDefaults()

	if err := o.Validate(); err != nil {
		return err
	}

	return nil
}

func getAPISpecString(apiSpec *openapi3.T) (string, error) {
	bApi, err := apiSpec.MarshalJSON()
	if err != nil {
		return "", err
	}

	yamlAPI, err := yaml.JSONToYAML(bApi)
	if err != nil {
		return "", nil
	}

	return string(yamlAPI), nil
}

func init() {
	rootCmd.AddCommand(generateCmd)
	// This should be deprecated soon.
	// See `apiCmd.Deprecated`.
	apiCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(
		&name,
		"name",
		"",
		"",
		"the name to give the API resource e.g. --name my-api",
	)

	generateCmd.Flags().StringVarP(
		&namespace,
		"namespace",
		"n",
		"default",
		"the namespace of the API resource e.g. --namespace my-namespace, -n my-namespace",
	)

	generateCmd.Flags().StringVarP(
		&apiSpecPath,
		"in",
		"i",
		"",
		"file path or URL to OpenAPI spec file to generate mappings from. e.g. --in apispec.yaml",
	)
	generateCmd.MarkFlagRequired("in")

	generateCmd.Flags().StringVarP(
		&serviceName,
		"upstream.service",
		"",
		"",
		"name of upstream service",
	)

	generateCmd.Flags().StringVarP(
		&serviceNamespace,
		"upstream.namespace",
		"",
		"default",
		"namespace of upstream service",
	)

	generateCmd.Flags().Uint32VarP(
		&servicePort,
		"upstream.port",
		"",
		80,
		"port of upstream service",
	)

	generateCmd.Flags().StringVarP(
		&envoyFleetName,
		"envoyfleet.name",
		"",
		"kusk-gateway-envoy-fleet",
		"name of envoyfleet to use for this API. Default: kusk-gateway-envoy-fleet",
	)

	generateCmd.Flags().StringVarP(
		&envoyFleetNamespace,
		"envoyfleet.namespace",
		"",
		kusknamespace,
		"namespace of envoyfleet to use for this API. Default: kusk-system",
	)

	apiTemplate = template.Must(template.New("api").Parse(templates.APITemplate))
}

var generateDescription = `Description:

Generate accepts your OpenAPI spec file as input either as a local file or a URL and generates a Kusk 
Gateway compatible API resource that you can apply directly into your cluster. 

It does this via the x-kusk extension which will be added automatically if one is not already set. It will 
set the upstream service, namespace and port to the flag values passed, respectively, and set the rest of 
the settings to defaults.

If the x-kusk extension is already present, it will override the upstream service, namespace and port to 
the flag values passed, respectively, and leave the rest of the settings as they are.`

var generateHelp = `
No name specified:
kusk generate \
	-i spec.yaml \
	--envoyfleet.name kusk-gateway-envoy-fleet \
	--envoyfleet.namespace kusk-system

In the above example, kusk will use the openapi spec info.title to generate a manifest name and leave the existing
x-kusk extension settings

No API namespace specified:
kusk api generate \
	-i spec.yaml \
	--name httpbin-api \
	--upstream.service httpbin \
	--upstream.port 8080 \
	--envoyfleet.name kusk-gateway-envoy-fleet

In the above example, as --namespace isn't defined, it will assume the default namespace.

Namespace specified:
kusk api generate \
	-i spec.yaml \
	--name httpbin-api \
	--upstream.service httpbin \
	--upstream.namespace my-namespace \
	--upstream.port 8080 \
	--envoyfleet.name kusk-gateway-envoy-fleet

OpenAPI definition form URL:
This will fetch the OpenAPI document from the provided URL and generate a Kusk Gateway API resource

kusk api generate \
	-i https://raw.githubusercontent.com/$ORG_OR_USER/$REPO/myspec.yaml \
	--name httpbin-api \
	--upstream.service httpbin \
	--upstream.namespace my-namespace \
	--upstream.port 8080 \
	--envoyfleet.name kusk-gateway-envoy-fleet
`
