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
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/getkin/kin-openapi/openapi3"
	fileWatcher "github.com/kubeshop/kusk-gateway/cmd/kusk/internal/mocking/filewatcher"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/templates"
	"github.com/kubeshop/kusk-gateway/pkg/options"
	"github.com/kubeshop/kusk-gateway/pkg/spec"
	"github.com/kubeshop/testkube/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	file  string
	watch bool
)

func init() {
	//add to root command
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVarP(&file, "file", "f", "", "file to deploy")
	deployCmd.MarkFlagRequired("file")

	deployCmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch file changes and deploy on change")
	deployCmd.Flags().StringVar(&name, "name", "", "name to name API with ")

}

// apiCmd represents the api command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy command to deploy your apis",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true
		manifest, err := getParsedAndValidatedOpenAPISpec(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		//apply
		fmt.Println(manifest)

		if watch {
			var watcher *fileWatcher.FileWatcher
			absoluteApiSpecPath := file
			if err != nil {
				ui.Fail(err)
			}

			watcher, err = fileWatcher.New(absoluteApiSpecPath)
			if err != nil {
				ui.Fail(err)
			}
			defer watcher.Close()

			done := make(chan os.Signal, 1)
			signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

			if watcher != nil {
				ui.Info(ui.White("⏳ watching for file changes in " + file))
				go watcher.Watch(func() {
					ui.Info("✍️ change detected in " + file)
					_, err := getParsedAndValidatedOpenAPISpec(file)
					if err != nil {
						ui.Fail(err)
					}
					fmt.Println("update")
				}, done)
			}
			<-done
		}

	},
}

func getParsedAndValidatedOpenAPISpec(apiSpecPath string) (string, error) {
	parsedApiSpec, err := spec.NewParser(openapi3.NewLoader()).Parse(apiSpecPath)
	if err != nil {
		return "", err
	}

	if _, ok := parsedApiSpec.ExtensionProps.Extensions["x-kusk"]; !ok {
		parsedApiSpec.ExtensionProps.Extensions["x-kusk"] = options.Options{}
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
