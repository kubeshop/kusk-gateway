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
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
)

var (
	noApi                         bool
	noDashboard                   bool
	noEnvoyFleet                  bool
	releaseName, releaseNamespace string

	analyticsEnabled = "true"
)

func init() {
	clusterCmd.AddCommand(installCmd)
	installCmd.Flags().StringVar(&releaseName, "name", "kusk-gateway", "installation name")
	installCmd.Flags().StringVar(&releaseNamespace, "namespace", "kusk-system", "namespace to install in")

	installCmd.Flags().BoolVar(&noDashboard, "no-dashboard", false, "don't the install dashboard")
	installCmd.Flags().BoolVar(&noApi, "no-api", false, "don't install the api. Setting this flag implies --no-dashboard")
	installCmd.Flags().BoolVar(&noEnvoyFleet, "no-envoy-fleet", false, "don't install any envoy fleets")

	if enabled, ok := os.LookupEnv("ANALYTICS_ENABLED"); ok {
		analyticsEnabled = enabled
	}
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install kusk-gateway, envoy-fleet, api, and dashboard in a single command",
	Long: `
	Install kusk-gateway, envoy-fleet, api, and dashboard in a single command.

	$ kusk install

	Will install kusk-gateway, a public (for your APIS) and private (for the kusk dashboard and api)
	envoy-fleet, api, and dashboard in the kusk-system namespace using helm.

	$ kusk install --name=my-release --namespace=my-namespace

	Will create a helm release named with --name in the namespace specified by --namespace.

	$ kusk install --no-dashboard --no-api --no-envoy-fleet

	Will install kusk-gateway, but not the dashboard, api, or envoy-fleet.
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reportError := func(err error) {
			if err != nil {
				errors.NewErrorReporter(cmd, err).Report()
			}
		}

		spinner := utils.NewSpinner("Installing kusk")

		if err := apply("deployment.yaml"); err != nil {
			spinner.Fail("failed installing kusk", err)
			reportError(err)
			return err
		}
		namespace := "kusk-system"
		name := "kusk-gateway-manager"

		c, err := utils.GetK8sClient()
		if err != nil {
			reportError(err)
			return err
		}

		if err := utils.WaitForPodsReady(cmd.Context(), c, namespace, name, time.Duration(5*time.Minute)); err != nil {
			spinner.Fail("failed installing kusk", err)
			reportError(err)
			return err
		}
		spinner.Success()

		if !noEnvoyFleet {
			spinner = utils.NewSpinner("Installing Envoy Fleets...")

			if err := apply("fleets.yaml"); err != nil {
				spinner.Fail("failed installing Envoy Fleets", err)
				reportError(err)
				return err
			}
			spinner.Success()
		}

		if !noApi {
			spinner = utils.NewSpinner("Installing API Server...")

			if err := apply("apis.yaml"); err != nil {
				spinner.Fail("failed installing API Server", err)
				reportError(err)
				return err
			}
			spinner.Success()
		} else if noApi {
			return nil
		}

		if !noDashboard {
			spinner = utils.NewSpinner("Installing Dashboard...")

			if err := apply("dashboard.yaml"); err != nil {
				spinner.Fail("failed installing Dashboard", err)
				reportError(err)
				return err
			}
			spinner.Success()
			printPortForwardInstructions("dashboard", releaseNamespace, envoyFleetName)
		}
		return nil
	},
}

func apply(filename string) error {
	instCmd := NewKubectlCmd()
	instCmd.SetArgs([]string{"apply", fmt.Sprintf("-f=%s", getEmbeddedFile(filename))})
	return instCmd.Execute()
}
func getEmbeddedFile(filename string) string {
	apis, _ := manifest.ReadFile(fmt.Sprintf("manifests/%s", filename))
	tmpfile, _ := ioutil.TempFile("", "kusk-man")

	tmpfile.Write(apis)
	return tmpfile.Name()
}

func printPortForwardInstructions(service, releaseNamespace, envoyFleetName string) {
	if service == "dashboard" {
		pterm.Info.Println("kusk dashboard is now available. To access it run: kusk dashboard")
		return
	}

	pterm.Info.Println(
		"To access the api , port forward to the envoy-fleet service that exposes it\n" +
			fmt.Sprintf("\t$ kubectl port-forward -n %s svc/%s 8080:80\n", releaseNamespace, envoyFleetName) +
			"\tand go http://localhost:8080/api",
	)
}
