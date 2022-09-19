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
	"os"
	"os/exec"

	"github.com/kubeshop/testkube/pkg/ui"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
)

var installOnUpgrade bool

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade kusk-gateway, envoy-fleet, api, and dashboard in a single command",
	Long: `
	Upgrade kusk-gateway, envoy-fleet, api, and dashboard in a single command.

	$ kusk upgrade

	Will upgrade kusk-gateway, a public (for your APIS) and private (for the kusk dashboard and api)
	envoy-fleet, api, and dashboard in the kusk-system namespace using helm.

	$ kusk upgrade --name=my-release --namespace=my-namespace

	Will upgrade a helm release named with --name in the namespace specified by --namespace.

	$ kusk upgrade --install

	Will upgrade kusk-gateway, the dashboard, api, and envoy-fleets and install them if they are not installed`,
	Run: func(cmd *cobra.Command, args []string) {
		reportError := func(err error) {
			if err != nil {
				errors.NewErrorReporter(cmd, err).Report()
			}
		}

		spinner := utils.NewSpinner("Looking for Helm...")
		helmPath, err := exec.LookPath("helm")
		if err != nil {
			spinner.Fail("Looking for Helm: ", err)
			os.Exit(1)
		}
		spinner.Success()

		spinner = utils.NewSpinner("Adding Kubeshop helm repository...")
		err = addKubeshopHelmRepo(helmPath)
		if err != nil {
			spinner.Fail("Adding Kubeshop helm repository: ", err)
			reportError(err)
			os.Exit(1)
		}
		spinner.Success()

		spinner = utils.NewSpinner("Fetching the latest charts...")
		err = updateHelmRepos(helmPath)
		if err != nil {
			spinner.Fail("Looking for Helm: ", err)
			reportError(err)
			os.Exit(1)
		}

		releases, err := listReleases(helmPath, releaseName, releaseNamespace)
		if err != nil {
			spinner.Fail("Listing Helm releases: ", err)
			reportError(err)
			os.Exit(1)
		}
		spinner.Success()

		if _, kuskGatewayInstalled := releases[releaseName]; kuskGatewayInstalled || installOnUpgrade {
			var spinnerMessage string
			if kuskGatewayInstalled {
				spinnerMessage = "Upgrading Kusk Gateway..."
			} else {
				spinnerMessage = "Installing Kusk Gateway..."
			}
			spinner = utils.NewSpinner(spinnerMessage)
			err = installKuskGateway(helmPath, releaseName, releaseNamespace)
			if err != nil {
				spinner.Fail("Upgrading Kusk Gateway: ", err)
				reportError(err)
				os.Exit(1)
			}
			spinner.Success()
		} else {
			pterm.Info.Println("Kusk Gateway not installed and --install not specified, skipping")
		}

		envoyFleetName := fmt.Sprintf("%s-envoy-fleet", releaseName)

		if _, publicEnvoyFleetInstalled := releases[envoyFleetName]; publicEnvoyFleetInstalled || installOnUpgrade {
			if !noEnvoyFleet {
				var spinnerMessage string
				if publicEnvoyFleetInstalled {
					spinnerMessage = "Upgrading public Envoy Fleet..."
				} else {
					spinnerMessage = "Installing public Envoy Fleet..."
				}
				spinner = utils.NewSpinner(spinnerMessage)
				err = installPublicEnvoyFleet(helmPath, envoyFleetName, releaseNamespace)
				if err != nil {
					spinner.Fail("Upgrading Envoy Fleet: ", err)
					reportError(err)
					os.Exit(1)
				}
				spinner.Success()
			} else {
				ui.Info(ui.LightYellow("--no-envoy-fleet set - skipping envoy fleet installation"))
			}
		} else {
			pterm.Info.Println("Envoy Fleet not installed and --install not specified, skipping")
		}

		envoyFleetName = fmt.Sprintf("%s-private-envoy-fleet", releaseName)

		if _, privateEnvoyFleetInstalled := releases[envoyFleetName]; privateEnvoyFleetInstalled || installOnUpgrade {
			var spinnerMessage string
			if privateEnvoyFleetInstalled {
				spinnerMessage = "Upgrading private Envoy Fleet..."
			} else {
				spinnerMessage = "Installing private Envoy Fleet..."
			}
			spinner = utils.NewSpinner(spinnerMessage)
			err = installPrivateEnvoyFleet(helmPath, envoyFleetName, releaseNamespace)
			if err != nil {
				spinner.Fail("Upgrading private Envoy Fleet: ", err)
				reportError(err)
				os.Exit(1)
			}
			spinner.Success()
		} else {
			pterm.Info.Println("Private Envoy Fleet not installed and --install not specified, skipping")
		}

		apiReleaseName := fmt.Sprintf("%s-api", releaseName)
		if _, apiInstalled := releases[apiReleaseName]; apiInstalled || installOnUpgrade {
			var spinnerMessage string
			if apiInstalled {
				spinnerMessage = "Upgrading Kusk API..."
			} else {
				spinnerMessage = "Installing Kusk API..."
			}
			spinner = utils.NewSpinner(spinnerMessage)
			err = installApi(helmPath, apiReleaseName, releaseNamespace, envoyFleetName)
			if err != nil {
				spinner.Fail("Upgrading Kusk API: ", err)
				reportError(err)
				os.Exit(1)
			}
			spinner.Success()
		} else {
			pterm.Info.Println("api not installed and --install not specified, skipping")
		}

		dashboardReleaseName := fmt.Sprintf("%s-dashboard", releaseName)
		if _, dashboardInstalled := releases[dashboardReleaseName]; dashboardInstalled || installOnUpgrade {
			var spinnerMessage string
			if dashboardInstalled {
				spinnerMessage = "Upgrading Kusk Dashboard..."
			} else {
				spinnerMessage = "Installing Kusk Dashboard..."
			}
			spinner = utils.NewSpinner(spinnerMessage)
			err = installDashboard(helmPath, dashboardReleaseName, releaseNamespace, envoyFleetName)
			if err != nil {
				spinner.Fail("Looking for Helm: ", err)
				reportError(err)
				os.Exit(1)
			}
			spinner.Success()
		} else {
			pterm.Info.Println("dashboard not installed and --install not specified, skipping")
		}

		pterm.Success.Printfln("Upgraded succesfully!")
	},
}

func init() {
	clusterCmd.AddCommand(upgradeCmd)

	upgradeCmd.Flags().StringVar(&releaseName, "name", "kusk-gateway", "name of release to update")
	upgradeCmd.Flags().StringVar(&releaseNamespace, "namespace", "kusk-system", "namespace to upgrade in")
	upgradeCmd.Flags().BoolVar(&installOnUpgrade, "install", false, "install components if not installed")
}
