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
.
*/

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/kuskui"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/mocking/filewatcher"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/templates"
	"github.com/kubeshop/kusk-gateway/pkg/options"
	"github.com/kubeshop/kusk-gateway/pkg/spec"
)

var (
	file  string
	watch bool
)

func init() {
	//add to root command
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVarP(&file, "in", "i", "", "file path or URL to OpenAPI spec file to generate mappings from. e.g. --in apispec.yaml")
	deployCmd.MarkFlagRequired("file")

	deployCmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch file changes and deploy on change")
	deployCmd.Flags().StringVar(&name, "name", "", "name of the API")
	deployCmd.Flags().StringVar(&namespace, "namespace", "default", "name of the API")
	deployCmd.Flags().StringVarP(&envoyFleetName, "envoyfleet.name", "", "kusk-gateway-envoy-fleet", "name of envoyfleet to use for this API. Default: kusk-gateway-envoy-fleet")

	deployCmd.Flags().StringVarP(&envoyFleetNamespace, "envoyfleet.namespace", "", "kusk-system", "namespace of envoyfleet to use for this API. Default: kusk-system")

}

// apiCmd represents the api command
var deployCmd = &cobra.Command{
	Use:           "deploy",
	Short:         "deploy command to deploy your apis",
	SilenceErrors: true,
	Long:          ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		reportError := func(err error) {
			if err != nil {
				errors.NewErrorReporter(cmd, err).Report()
			}
		}

		ctx := context.Background()
		cmd.SilenceUsage = true
		originalManifest, err := getParsedAndValidatedOpenAPISpec(file)
		if err != nil {
			reportError(err)
			return err
		}

		kuskui.PrintSuccess(fmt.Sprintf("successfully parsed %s", file))
		kuskui.PrintStart(fmt.Sprintf("initiallizing deployment to fleet %s", envoyFleetName))

		k8sclient, err := utils.GetK8sClient()
		if err != nil {
			reportError(err)
			return err
		}

		api := &v1alpha1.API{}

		yaml.Unmarshal([]byte(originalManifest), api)
		if len(api.Namespace) == 0 {
			api.Namespace = "default"
		}
		if len(api.Name) == 0 {
			api.Name = name
		}

		if err := k8sclient.Create(ctx, api, &client.CreateOptions{}); err != nil {
			if apierrors.IsAlreadyExists(err) {
				ap := &v1alpha1.API{}
				if err := k8sclient.Get(ctx, client.ObjectKey{Namespace: api.Namespace, Name: api.Name}, ap); err != nil {
					reportError(err)
					return err
				}
				api.SetResourceVersion(ap.GetResourceVersion())
				if err := k8sclient.Update(ctx, api, &client.UpdateOptions{}); err != nil {
					reportError(err)
					return err
				}
				kuskui.PrintSuccess(fmt.Sprintf("api.gateway.kusk.io/%s updated", api.Name))
			} else {
				reportError(err)
				return err
			}
		} else {
			kuskui.PrintInfo(fmt.Sprintf("api.gateway.kusk.io/%s created\n", api.Name))
		}

		if _, e := url.ParseRequestURI(file); e != nil {
			if watch {
				var watcher *filewatcher.FileWatcher

				watcher, err = filewatcher.New(file)
				if err != nil {
					reportError(err)
					return err
				}
				defer watcher.Close()

				done := make(chan os.Signal, 1)
				signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

				if watcher != nil {
					kuskui.PrintInfo(fmt.Sprintf("⏳ watching for API changes in %s", file))
					go watcher.Watch(func() {
						kuskui.PrintStart(fmt.Sprintf("✍️ change detected in %s", file))
						kuskui.PrintSuccess(fmt.Sprintf("successfully parsed %s", file))
						kuskui.PrintStart(fmt.Sprintf("initiallizing deployment to fleet %s", envoyFleetName))

						manifest, err := getParsedAndValidatedOpenAPISpec(file)
						if err != nil {
							reportError(err)
							kuskui.PrintError(err.Error())
							return
						}
						api := &v1alpha1.API{}
						if err := yaml.Unmarshal([]byte(manifest), api); err != nil {
							reportError(err)
							kuskui.PrintError(err.Error())
							return
						}

						if len(api.Namespace) == 0 {
							api.Namespace = "default"
						}
						if len(api.Name) == 0 {
							api.Name = name
						}
						ap := &v1alpha1.API{}
						if err := k8sclient.Get(ctx, client.ObjectKey{Namespace: api.Namespace, Name: api.Name}, ap); err != nil {
							reportError(err)
							kuskui.PrintError(err.Error())
							return
						}
						api.SetResourceVersion(ap.GetResourceVersion())

						if err := k8sclient.Update(ctx, api, &client.UpdateOptions{}); err != nil {
							reportError(err)
							kuskui.PrintError(err.Error())
							return
						} else {
							kuskui.PrintSuccess(fmt.Sprintf("api.gateway.kusk.io/%s updated", api.Name))
						}
					}, done)
				}
				<-done
			}
		} else if e == nil {
			kuskui.PrintWarning("Warning: cannot watch URL. '--watch, -w' flag ignored!")
		}
		return nil
	},
}

func getParsedAndValidatedOpenAPISpec(apiSpecPath string) (string, error) {
	const KuskExtensionKey = "x-kusk"

	parsedApiSpec, err := spec.NewParser(openapi3.NewLoader()).Parse(apiSpecPath)
	if err != nil {
		return "", err
	}

	if _, ok := parsedApiSpec.ExtensionProps.Extensions[KuskExtensionKey]; !ok {
		parsedApiSpec.ExtensionProps.Extensions[KuskExtensionKey] = options.Options{}
	}

	if name == "" {
		// kubernetes manifests cannot have . in the name so replace them
		name = strings.ReplaceAll(parsedApiSpec.Info.Title, ".", "-")
	}

	opts, err := spec.GetOptions(parsedApiSpec)
	if err != nil {
		return "", err
	}

	if err := opts.Validate(); err != nil {
		return "", err
	}

	// override top level upstream service if undefined.
	if serviceName != "" && serviceNamespace != "" && servicePort != 0 {
		xKusk := parsedApiSpec.ExtensionProps.Extensions[KuskExtensionKey].(options.Options)
		xKusk.Upstream = &options.UpstreamOptions{
			Service: &options.UpstreamService{
				Name:      serviceName,
				Namespace: serviceNamespace,
				Port:      servicePort,
			},
		}

		parsedApiSpec.ExtensionProps.Extensions[KuskExtensionKey] = xKusk
	}

	if err := validateExtensionOptions(parsedApiSpec.ExtensionProps.Extensions[KuskExtensionKey]); err != nil {
		return "", err
	}

	if apiSpec, err = getAPISpecString(parsedApiSpec); err != nil {
		return "", err
	}

	var manifest bytes.Buffer

	if err := apiTemplate.Execute(&manifest, templates.APITemplateArgs{
		Name:                name,
		Namespace:           namespace,
		EnvoyfleetName:      envoyFleetName,
		EnvoyfleetNamespace: envoyFleetNamespace,
		Spec:                strings.Split(apiSpec, "\n"),
	}); err != nil {
		return "", err
	}
	return manifest.String(), nil
}
