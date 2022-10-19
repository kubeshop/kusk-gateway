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
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/kuskui"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
)

var (
	installOnUpgrade              bool
	releaseName, releaseNamespace string
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Kusk Gateway, EnvoyFleet, Kusk API, and Kusk Dashboard in a single command",
	Long: `
	Upgrade Kusk Gateway, EnvoyFleet, Kusk API, and Kusk Dashboard in a single command.

	$ kusk cluster upgrade

	Will upgrade or install Kusk Gatewway, Kusk Dashboard, Kusk API, and EnvoyFleets and install them if they are not installed`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reportError := func(err error) {
			if err != nil {
				errors.NewErrorReporter(cmd, err).Report()
			}
		}

		var dir string
		var err error
		if dir, err = getManifests(); err != nil {
			return err
		}

		c, err := utils.GetK8sClient()
		if err != nil {
			reportError(err)
			return err
		}

		kuskui.PrintInfo("Checking if Kusk is already installed...")

		deployments := appsv1.DeploymentList{}
		if err := c.List(cmd.Context(), &deployments, &client.ListOptions{Namespace: kusknamespace}); err != nil {
			reportError(err)
			return err
		}

		if len(deployments.Items) == 0 {
			kuskui.PrintInfo("Kusk is not installed on the cluster")
			os.Exit(0)
		}

		for _, deployment := range deployments.Items {
			switch deployment.Name {
			case "kusk-gateway-manager":
				if !utils.IsUptodate(getVersions(deployment.Name, "manager", deployment)) {
					kuskGatewaySpinner := utils.NewSpinner("Upgrading Kusk Gateway...")
					for _, c := range deployment.Spec.Template.Spec.Containers {
						if c.Name == "manager" {
							if err := applyk(dir); err != nil {
								kuskGatewaySpinner.Fail("Failed upgrading Kusk Gateway", err.Error())
								reportError(err)
								return err
							}
						}
					}

					if err := utils.WaitForPodsReady(cmd.Context(), c, namespace, name, time.Duration(5*time.Minute), "component"); err != nil {
						kuskui.PrintError("Failed upgrading EnvoyFleet", err.Error())
						reportError(err)
						return err
					}

					kuskGatewaySpinner.Success("Upgraded Kusk Gateway")
				}

			case "kusk-gateway-private-envoy-fleet", "kusk-gateway-envoy-fleet":
				envoyFleetSpinner := utils.NewSpinner("Upgrading EnvoyFleet...")
				if err := applyf(filepath.Join(dir, manifests_dir, "fleets.yaml")); err != nil {
					envoyFleetSpinner.Fail("Failed upgrading EnvoyFleet", err.Error())
					reportError(err)
					return err
				}

				if err := utils.WaitForPodsReady(cmd.Context(), c, namespace, "envoy", time.Duration(5*time.Minute), "component"); err != nil {
					envoyFleetSpinner.Fail("Failed upgrading EnvoyFleet", err.Error())
					reportError(err)
					return err
				}

				envoyFleetSpinner.Success("Upgraded EnvoyFleet")
			case kuskgatewayapi:
				if !utils.IsUptodate(getVersions(deployment.Name, kuskgatewayapi, deployment)) {
					kuskApiSpinner := utils.NewSpinner("Upgrading Kusk API...")
					if err := applyf(filepath.Join(dir, manifests_dir, "api_server.yaml")); err != nil {
						kuskApiSpinner.Fail("Failed upgrading Kusk API", err.Error())
						reportError(err)
						return err
					}
					if err := utils.WaitForPodsReady(cmd.Context(), c, namespace, kuskgatewayapi, time.Duration(5*time.Minute), "instance"); err != nil {
						kuskApiSpinner.Fail("Failed upgrading Kusk API", err.Error())
						reportError(err)
						return err
					}

					if err := applyf(filepath.Join(dir, manifests_dir, "api_server_api.yaml")); err != nil {
						kuskApiSpinner.Fail("Failed upgrading Kusk API", err.Error())
						reportError(err)
						return err
					}
					kuskApiSpinner.Success("Upgraded Kusk API")
				}
			case kuskgatewaydashboard:
				if !utils.IsUptodate(getVersions(kuskgatewaydashboard, kuskgatewaydashboard, deployment)) {
					kuskDashboardSpinner := utils.NewSpinner("Upgrading Kusk Dashboard...")
					if err := applyf(filepath.Join(dir, manifests_dir, "dashboard.yaml")); err != nil {
						kuskDashboardSpinner.Fail("Failed upgrading Kusk Dashboard", err.Error())
						reportError(err)
						return err
					}
					if err := utils.WaitForPodsReady(cmd.Context(), c, namespace, kuskgatewaydashboard, time.Duration(5*time.Minute), "instance"); err != nil {
						kuskDashboardSpinner.Fail("Failed upgrading Kusk Dashboard", err.Error())
						reportError(err)
						return err
					}

					if err := applyf(filepath.Join(dir, manifests_dir, "dashboard_envoyfleet.yaml")); err != nil {
						kuskDashboardSpinner.Fail("Failed upgrading Kusk Dashboard", err.Error())
						reportError(err)
						return err
					}

					if err := applyf(filepath.Join(dir, manifests_dir, "dashboard_staticroute.yaml")); err != nil {
						kuskDashboardSpinner.Fail("Failed upgrading Kusk Dashboard", err.Error())
						reportError(err)
						return err
					}
					kuskDashboardSpinner.Success("Upgraded Kusk Dashboard")
				}
			}
		}

		kuskui.PrintSuccess("Kusk upgraded successfully")

		return nil
	},
}

func init() {
	clusterCmd.AddCommand(upgradeCmd)

	upgradeCmd.Flags().StringVar(&releaseName, "name", kuskgateway, "name of release to update")
	upgradeCmd.Flags().StringVar(&releaseNamespace, "namespace", kusknamespace, "namespace to upgrade in")
	upgradeCmd.Flags().BoolVar(&installOnUpgrade, "install", false, "install components if not installed")
}

func getVersions(component, container string, deployment appsv1.Deployment) (latest string, current string) {
	githubClient, err := utils.NewGithubClient("", nil)
	if err != nil {
		return "", ""
	}

	var repoName string
	switch component {
	case kuskgatewaymanager:
		repoName = "kusk-gateway"
	case kuskgatewayapi:
		repoName = "kuskgateway-api-server"
	default:
		repoName = component
	}

	latest, err = githubClient.GetLatest(repoName)
	if err != nil {
		return "", ""
	}

	for _, c := range deployment.Spec.Template.Spec.Containers {
		if c.Name == container {
			current = strings.Split(c.Image, ":")[1]
			break
		}
	}
	return latest, current
}
