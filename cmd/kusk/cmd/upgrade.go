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
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/kubeshop/testkube/pkg/process"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
)

var (
	installOnUpgrade              bool
	releaseName, releaseNamespace string
)

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
	RunE: func(cmd *cobra.Command, args []string) error {
		reportError := func(err error) {
			if err != nil {
				errors.NewErrorReporter(cmd, err).Report()
			}
		}
		ghclient, err := utils.NewGithubClient("", nil)
		if err != nil {
			reportError(err)
			return err
		}

		latest, err := ghclient.GetLatest()
		if err != nil {
			reportError(err)
			return err
		}

		var dir string
		// if !latest {
		if dir, err = getManifests(); err != nil {
			return err
		}
		// } else {
		// 	if dir, err = getManifestsFromUrl(); err != nil {
		// 		return err
		// 	}
		// }

		c, err := utils.GetK8sClient()
		if err != nil {
			reportError(err)
			return err
		}

		fmt.Println("✅ Checking if kusk is already installed...")

		deployments := appsv1.DeploymentList{}
		if err := c.List(cmd.Context(), &deployments, &client.ListOptions{Namespace: "kusk-system"}); err != nil {
			reportError(err)
			return err
		}
		fmt.Println("DIR", dir)
		for _, deployment := range deployments.Items {
			switch deployment.Name {
			case "kusk-gateway-manager":
				for _, c := range deployment.Spec.Template.Spec.Containers {
					if c.Name == "manager" {
						imagetag := strings.Split(c.Image, ":")
						if len(imagetag) > 1 {
							fmt.Println("imagetag", imagetag)
							latestVersion, err := version.NewVersion(latest)
							if err != nil {
								errors.NewErrorReporter(cmd, err).Report()
							}

							currentVersion, err := version.NewVersion(imagetag[1])
							if err != nil {
								errors.NewErrorReporter(cmd, err).Report()
							}

							if currentVersion != nil && currentVersion.LessThan(latestVersion) {
								if err := applyk(dir); err != nil {
									fmt.Println("❌ failed installing Kusk", err)
									reportError(err)
									return err
								}
							}
							if currentVersion == nil {
								if err := applyk(dir); err != nil {
									fmt.Println("❌ failed installing Kusk", err)
									reportError(err)
									return err
								}
							}
						}
					}
				}

				if err := utils.WaitForPodsReady(cmd.Context(), c, namespace, name, time.Duration(5*time.Minute), "component"); err != nil {
					fmt.Println("❌ failed installing kusk", err)
					reportError(err)
					return err
				}

			case "kusk-gateway-private-envoy-fleet", "kusk-gateway-envoy-fleet":
				if err := applyf(filepath.Join(dir, manifests_dir, "fleets.yaml")); err != nil {
					fmt.Println("❌ failed installing Envoy Fleets", err)
					reportError(err)
					return err
				}

				if err := utils.WaitForPodsReady(cmd.Context(), c, namespace, "envoy", time.Duration(5*time.Minute), "component"); err != nil {
					fmt.Println("❌ failed installing Envoyfleets", err)
					reportError(err)
					return err
				}

				fmt.Println("✅ Envoy Fleets installed")
			case "kusk-gateway-api":
				if err := applyf(filepath.Join(dir, manifests_dir, "apis.yaml")); err != nil {
					fmt.Println("❌ failed installing API Server", err)
					reportError(err)
					return err
				}
				if err := utils.WaitForPodsReady(cmd.Context(), c, namespace, "kusk-gateway-api", time.Duration(5*time.Minute), "instance"); err != nil {
					fmt.Println("❌ failed installing kusk", err)
					reportError(err)
					return err
				}
				fmt.Println("✅ API Server installed")
			case "kusk-gateway-dashboard":
				if err := applyf(filepath.Join(dir, manifests_dir, "dashboard.yaml")); err != nil {
					fmt.Println("❌ failed installing Dashboard", err)
					reportError(err)
					return err
				}
				if err := utils.WaitForPodsReady(cmd.Context(), c, namespace, "kusk-gateway-dashboard", time.Duration(5*time.Minute), "instance"); err != nil {
					fmt.Println("❌ failed installing kusk", err)
					reportError(err)
					return err
				}
				fmt.Println("✅ Dashboard installed")
			}
		}

		// if err := applyk(dir); err != nil {
		// 	fmt.Println("❌ failed installing Kusk", err)
		// 	reportError(err)
		// 	return err
		// }

		return nil
	},
}

func init() {
	clusterCmd.AddCommand(upgradeCmd)

	upgradeCmd.Flags().StringVar(&releaseName, "name", "kusk-gateway", "name of release to update")
	upgradeCmd.Flags().StringVar(&releaseNamespace, "namespace", "kusk-system", "namespace to upgrade in")
	upgradeCmd.Flags().BoolVar(&installOnUpgrade, "install", false, "install components if not installed")
}

func addKubeshopHelmRepo(helmPath string) error {
	_, err := process.Execute(helmPath, "repo", "add", "kubeshop", "https://kubeshop.github.io/helm-charts/")
	if err != nil && !strings.Contains(err.Error(), "Error: repository name (kubeshop) already exists, please specify a different name") {
		return err
	}

	return nil
}

func updateHelmRepos(helmPath string) error {
	_, err := process.Execute(helmPath, "repo", "update")
	return err
}

type ReleaseDetails struct {
	chart   string
	version string
}

func listReleases(helmPath, releaseName, releaseNamespace string) (map[string]*ReleaseDetails, error) {
	command := []string{
		"ls",
		"-n", releaseNamespace,
		"-o", "json",
	}

	cmd := exec.Command(helmPath, command...)

	buffer := new(bytes.Buffer)
	errBuffer := new(bytes.Buffer)
	cmd.Stdout = buffer
	cmd.Stderr = errBuffer

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not start process: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		// TODO clean error output (currently it has buffer too - need to refactor in cmd)
		return nil, fmt.Errorf("process error: %w\noutput: %s", err, buffer.String())
	}

	if errBuffer.Len() > 0 {
		strError := errBuffer.String()
		// if errBuffer contains an actual error, then return. Otherwise it's a warning
		// so output to user
		if !strings.Contains(strError, "WARNING") {
			return nil, fmt.Errorf("helm list error: %s", errBuffer.String())
		}
		pterm.Warning.Print(strError)
	}

	var releases []struct {
		Name       string `json:"name"`
		Chart      string `json:"chart"`
		AppVersion string `json:"app_version"`
	}

	if err := json.Unmarshal(buffer.Bytes(), &releases); err != nil {
		return nil, err
	}

	releaseMap := make(map[string]*ReleaseDetails)

	for _, release := range releases {
		if strings.HasPrefix(release.Name, releaseName) {
			releaseMap[release.Name] = &ReleaseDetails{
				version: release.AppVersion,
				chart:   release.Chart,
			}
		}
	}

	return releaseMap, nil
}

func installKuskGateway(helmPath, releaseName, releaseNamespace string) error {
	command := []string{
		"upgrade",
		"--install",
		"--wait",
		"--create-namespace",
		"--namespace",
		releaseNamespace,
		"--set", fmt.Sprintf("fullnameOverride=%s", releaseName),
		"--set", fmt.Sprintf("analytics.enabled=%s", analyticsEnabled),
		releaseName,
		"kubeshop/kusk-gateway",
	}

	out, err := process.Execute(helmPath, command...)
	if err != nil {
		return err
	}

	pterm.Debug.Println("Helm install kusk gateway output: ", string(out))

	return nil
}

func installPublicEnvoyFleet(helmPath, releaseName, releaseNamespace string) error {
	return installEnvoyFleet(helmPath, releaseName, releaseNamespace, "LoadBalancer", true)
}

func installPrivateEnvoyFleet(helmPath, releaseName, releaseNamespace string) error {
	return installEnvoyFleet(helmPath, releaseName, releaseNamespace, "ClusterIP", false)
}

func installEnvoyFleet(helmPath, releaseName, releaseNamespace, serviceType string, isDefaultFleet bool) error {
	command := []string{
		"upgrade",
		"--install",
		"--wait",
		"--create-namespace",
		"--namespace",
		releaseNamespace,
		"--set", fmt.Sprintf("fullnameOverride=%s", releaseName),
		"--set", fmt.Sprintf("service.type=%s", serviceType),
		"--set", fmt.Sprintf("default=%t", isDefaultFleet),
		releaseName,
		"kubeshop/kusk-gateway-envoyfleet",
	}

	out, err := process.Execute(helmPath, command...)
	if err != nil {
		return err
	}

	pterm.Debug.Println("Helm install envoy fleet output: ", string(out))

	return nil
}

func installApi(helmPath, releaseName, releaseNamespace, envoyFleetName string) error {
	command := []string{
		"upgrade",
		"--install",
		"--wait",
		"--create-namespace",
		"--namespace",
		releaseNamespace,
		"--set", fmt.Sprintf("fullnameOverride=%s", releaseName),
		"--set", fmt.Sprintf("envoyfleet.name=%s", envoyFleetName),
		"--set", fmt.Sprintf("envoyfleet.namespace=%s", releaseNamespace),
		"--set", fmt.Sprintf("analytics.enabled=%s", analyticsEnabled),
		releaseName,
		"kubeshop/kusk-gateway-api",
	}

	out, err := process.Execute(helmPath, command...)
	if err != nil {
		return err
	}

	pterm.Debug.Println("Helm install api output", string(out))

	return nil
}

func installDashboard(helmPath, releaseName, releaseNamespace, envoyFleetName string) error {
	command := []string{
		"upgrade",
		"--install",
		"--wait",
		"--create-namespace",
		"--namespace",
		releaseNamespace,
		"--set", fmt.Sprintf("fullnameOverride=%s", releaseName),
		"--set", fmt.Sprintf("envoyfleet.name=%s", envoyFleetName),
		"--set", fmt.Sprintf("envoyfleet.namespace=%s", releaseNamespace),
		releaseName,
		"kubeshop/kusk-gateway-dashboard",
	}

	out, err := process.Execute(helmPath, command...)
	if err != nil {
		return err
	}

	pterm.Debug.Println("helm install dashboard output", string(out))

	return nil
}
