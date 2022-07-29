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
	"os/exec"

	"github.com/kubeshop/testkube/pkg/ui"
	"github.com/spf13/cobra"
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
		helmPath, err := exec.LookPath("helm")
		ui.ExitOnError("looking for helm", err)

		ui.Info("adding the kubeshop helm repository")
		err = addKubeshopHelmRepo(helmPath)
		ui.ExitOnError("adding kubeshop repo", err)
		ui.Info(ui.Green("done"))

		ui.Info("fetching the latest charts")
		err = updateHelmRepos(helmPath)
		ui.ExitOnError("updating helm repositories", err)
		ui.Info(ui.Green("done"))

		releases, err := listReleases(helmPath, releaseName, releaseNamespace)
		ui.ExitOnError("listing existing releases", err)

		if _, kuskGatewayInstalled := releases[releaseName]; kuskGatewayInstalled || installOnUpgrade {
			ui.Info("upgrading Kusk Gateway")
			err = installKuskGateway(helmPath, releaseName, releaseNamespace)
			ui.ExitOnError("upgrading kusk gateway", err)
			ui.Info(ui.Green("done"))
		} else {
			ui.Info("kusk gateway not installed and --install not specified, skipping")
		}

		envoyFleetName := fmt.Sprintf("%s-envoy-fleet", releaseName)

		if _, publicEnvoyFleetInstalled := releases[envoyFleetName]; publicEnvoyFleetInstalled || installOnUpgrade {
			if !noEnvoyFleet {
				ui.Info("upgrading Envoy Fleet")
				err = installPublicEnvoyFleet(helmPath, envoyFleetName, releaseNamespace)
				ui.ExitOnError("upgrading envoy fleet", err)
				ui.Info(ui.Green("done"))
			} else {
				ui.Info(ui.LightYellow("--no-envoy-fleet set - skipping envoy fleet installation"))
			}
		} else {
			ui.Info("envoy fleet not installed and --install not specified, skipping")
		}

		envoyFleetName = fmt.Sprintf("%s-private-envoy-fleet", releaseName)

		if _, privateEnvoyFleetInstalled := releases[envoyFleetName]; privateEnvoyFleetInstalled || installOnUpgrade {
			err = installPrivateEnvoyFleet(helmPath, envoyFleetName, releaseNamespace)
			ui.ExitOnError("upgrading envoy fleet", err)
		} else {
			ui.Info("private envoy fleet not installed and --install not specified, skipping")
		}

		apiReleaseName := fmt.Sprintf("%s-api", releaseName)
		if _, apiInstalled := releases[apiReleaseName]; apiInstalled || installOnUpgrade {
			ui.Info("upgrading Kusk API")
			err = installApi(helmPath, apiReleaseName, releaseNamespace, envoyFleetName)
			ui.ExitOnError("upgrading api", err)
			ui.Info(ui.Green("done"))
		} else {
			ui.Info("api not installed and --install not specified, skipping")
		}

		dashboardReleaseName := fmt.Sprintf("%s-dashboard", releaseName)
		if _, dashboardInstalled := releases[dashboardReleaseName]; dashboardInstalled || installOnUpgrade {
			ui.Info("upgrading Kusk Dashboard")
			err = installDashboard(helmPath, dashboardReleaseName, releaseNamespace, envoyFleetName)
			ui.ExitOnError("upgrading dashboard", err)

			ui.Info(ui.Green("done"))
		} else {
			ui.Info("dashboard not installed and --install not specified, skipping")
		}

		ui.Info(ui.Green("upgrade complete"))
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)

	upgradeCmd.Flags().StringVar(&releaseName, "name", "kusk-gateway", "installation name")
	upgradeCmd.Flags().StringVar(&releaseNamespace, "namespace", "kusk-system", "namespace to upgrade in")
	upgradeCmd.Flags().BoolVar(&installOnUpgrade, "install", false, "install components if not installed")
}
