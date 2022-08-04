/*
The MIT License (MIT)

Copyright © 2022 Kubeshop

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
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
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

	apply  bool
	output string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a Kusk Gateway API resource from your OpenAPI spec file",
	Long: `
	Generate accepts your OpenAPI spec file as input either as a local file or a URL pointing to your file
	and generates a Kusk Gateway compatible API resource that you can apply directly into your cluster.

	Configuration of the API resource is done via the x-kusk extension.

	If the OpenAPI spec doesn't have a top-level x-kusk annotation set, it will add them for you and set
	the upstream service, namespace and port to the flag values passed in respectively and set the rest of the settings to defaults.
	This is enough to get you started

	If the x-kusk extension is already present, it will override the the upstream service, namespace and port to the flag values passed in respectively
	and leave the rest of the settings as they are.

	You must specify the name of the envoyfleet you wish to use to expose your API. This is because Kusk Gateway could be managing more than one.
	In the future, we will add the notion of a default envoyfleet which kusk gateway will use when none is specified.

	If you do not specify the envoyfleet namespace, it will default to kusk-system.

	Sample usage

	No name specified
	kusk api generate \
		-i spec.yaml \
		--envoyfleet.name kusk-gateway-envoy-fleet \
		--envoyfleet.namespace kusk-system

	In the above example, kusk will use the openapi spec info.title to generate a manifest name and leave the existing
	x-kusk extension settings

	No api namespace specified
	kusk api generate \
		-i spec.yaml \
		--name httpbin-api \
		--upstream.service httpbin \
		--upstream.port 8080 \
		--envoyfleet.name kusk-gateway-envoy-fleet

	In the above example, as --namespace isn't defined, it will assume the default namespace.

	Namespace specified
	kusk api generate \
		-i spec.yaml \
		--name httpbin-api \
		--upstream.service httpbin \
		--upstream.namespace my-namespace \
		--upstream.port 8080 \
		--envoyfleet.name kusk-gateway-envoy-fleet

	OpenAPI spec at URL
	kusk api generate \
			-i https://raw.githubusercontent.com/$ORG_OR_USER/$REPO/myspec.yaml \
			 --name httpbin-api \
			 --upstream.service httpbin \
			 --upstream.namespace my-namespace \
			 --upstream.port 8080 \
			 --envoyfleet.name kusk-gateway-envoy-fleet

	This will fetch the OpenAPI document from the provided URL and generate a Kusk Gateway API resource
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(envoyFleetName) == 0 {
			if err := promptForFleet(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		cmd.SilenceUsage = true
		parsedApiSpec, err := spec.NewParser(openapi3.NewLoader()).Parse(apiSpecPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if _, ok := parsedApiSpec.ExtensionProps.Extensions["x-kusk"]; !ok {
			parsedApiSpec.ExtensionProps.Extensions["x-kusk"] = options.Options{}
		}

		// if name flag is not defined, use the swagger doc title which is guarunteed to be there
		if name == "" {
			// kubernetes manifests cannot have . in the name so replace them
			name = strings.ReplaceAll(parsedApiSpec.Info.Title, ".", "-")
			name = strings.ReplaceAll(parsedApiSpec.Info.Title, "_", "-")
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
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if apiSpec, err = getAPISpecString(parsedApiSpec); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		var manifest bytes.Buffer
		if err := apiTemplate.Execute(&manifest, templates.APITemplateArgs{
			Name:                name,
			Namespace:           namespace,
			EnvoyfleetName:      envoyFleetName,
			EnvoyfleetNamespace: envoyFleetNamespace,
			Spec:                strings.Split(apiSpec, "\n"),
		}); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if !apply {
			writeOrPrint(output, manifest)
			return
		}

		prompt := promptui.Prompt{
			Label:     "Do you want to proceed apply provide API to your cluster",
			IsConfirm: true,
			Default:   "y",
		}

		result, _ := prompt.Run()

		if result == "y" || result == "Y" || result == "" { // if enter is hit i.e. pick default results is an empty string
			api := &kuskv1.API{}
			if err := yaml.Unmarshal(manifest.Bytes(), api); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			k8sclient, err := utils.GetK8sClient()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			if err := k8sclient.Create(context.Background(), api, &client.CreateOptions{}); err != nil {
				if err.Error() == fmt.Sprintf(`apis.gateway.kusk.io "%s" already exists`, api.Name) { // fugly way to check if the object already exists
					fmt.Println(err)
					os.Exit(0)
				}

				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			} else {
				fmt.Printf("api.gateway.kusk.io/%s created\n", api.Name)
			}
		} else {
			writeOrPrint(output, manifest)
		}
	},
}

func writeOrPrint(output string, manifest bytes.Buffer) {
	if len(output) > 0 {
		if err := os.WriteFile(output, manifest.Bytes(), 0644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		} else {
			fmt.Printf("Successfully saved the API to %q\n", output)
		}
	} else {
		fmt.Println(manifest.String()) // if giving up on applying print it out
	}
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
		"",
		"name of envoyfleet to use for this API",
	)

	generateCmd.Flags().StringVarP(
		&envoyFleetNamespace,
		"envoyfleet.namespace",
		"",
		"kusk-system",
		"namespace of envoyfleet to use for this API. Default: kusk-system",
	)

	generateCmd.Flags().BoolVarP(
		&apply,
		"apply",
		"a",
		false,
		"to automatically apply the manifest to the cluster. Default: false",
	)

	generateCmd.Flags().StringVarP(
		&output,
		"output",
		"o",
		"",
		"path to the location where to save the output of the command",
	)

	apiTemplate = template.Must(template.New("api").Parse(templates.APITemplate))
}

func promptForFleet() error {
	if len(envoyFleetName) > 0 {
		return nil
	}
	k8sclient, err := utils.GetK8sClient()
	if err != nil {
		return err
	}
	fleets := &kuskv1.EnvoyFleetList{}

	if err := k8sclient.List(context.TODO(), fleets, &client.ListOptions{}); err != nil {
		return err
	}
	fmt.Println("Envoyfleets:")
	for i, f := range fleets.Items {
		fmt.Printf("%d. %s/%s\n", i+1, f.Namespace, f.Name)
	}

	fleet := promptGetInput("Please pick a fleet from the list above", len(fleets.Items))

	envoyFleetName = fleets.Items[fleet-1].Name
	envoyFleetNamespace = fleets.Items[fleet-1].Namespace

	return nil
}

func promptGetInput(message string, length int) int {
	index := 0
	validate := func(input string) (err error) {
		if index, err = strconv.Atoi(input); err != nil {
			return fmt.Errorf("%q is not a number", input)
		} else if index > length {
			return fmt.Errorf(fmt.Sprintf("you can only pick values between 1 and %d", length))
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:     message,
		AllowEdit: true,
		Validate:  validate,
	}

	_, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return index
}
