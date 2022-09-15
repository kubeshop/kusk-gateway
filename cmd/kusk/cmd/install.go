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
	"os"
	"os/exec"
	"strings"

	"github.com/kubeshop/testkube/pkg/process"
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
			reportError(err)
			os.Exit(1)
		}
		spinner.Success()

		spinner = utils.NewSpinner("Adding Kubeshop repository...")
		err = addKubeshopHelmRepo(helmPath)
		if err != nil {
			spinner.Fail("Adding Kubeshop repository: ", err)
			reportError(err)
			os.Exit(1)
		}
		spinner.Success()

		spinner = utils.NewSpinner("Fetching the latest charts...")
		err = updateHelmRepos(helmPath)
		if err != nil {
			spinner.Fail("Fetching the latest charts: ", err)
			reportError(err)
			os.Exit(1)
		}
		spinner.Success()

		releases, err := listReleases(helmPath, releaseName, releaseNamespace)
		if err != nil {
			pterm.Error.Println("Listing existing releases: ", err)
			reportError(err)
			os.Exit(1)
		}

		if _, kuskGatewayInstalled := releases[releaseName]; !kuskGatewayInstalled {
			spinner = utils.NewSpinner("Installing Kusk Gateway")
			err = installKuskGateway(helmPath, releaseName, releaseNamespace)
			if err != nil {
				spinner.Fail("Installing Kusk Gateway: ", err)
				reportError(err)
				os.Exit(1)
			}

			spinner.Success()
		} else {
			pterm.Info.Println("Kusk Gateway already installed, skipping... To upgrade to a new version run `kusk upgrade`")
		}

		envoyFleetName := fmt.Sprintf("%s-envoy-fleet", releaseName)
		if _, publicEnvoyFleetInstalled := releases[envoyFleetName]; !publicEnvoyFleetInstalled {
			if !noEnvoyFleet {
				spinner = utils.NewSpinner("Installing Envoy Fleet...")
				err = installPublicEnvoyFleet(helmPath, envoyFleetName, releaseNamespace)
				if err != nil {
					spinner.Fail("Installing Envoy Fleet: ", err)
					reportError(err)
					os.Exit(1)
				}
				spinner.Success()
			} else {
				pterm.Info.Println("--no-envoy-fleet set - skipping envoy fleet installation")
			}
		} else {
			pterm.Info.Println("Envoy Fleet already installed, skipping. To upgrade to a new version run `kusk upgrade`")
		}

		if noApi {
			pterm.Info.Println("--no-api set - skipping api installation")
			return
		}

		envoyFleetName = fmt.Sprintf("%s-private-envoy-fleet", releaseName)
		if _, privateEnvoyFleetInstalled := releases[envoyFleetName]; !privateEnvoyFleetInstalled {
			if !noEnvoyFleet {
				spinner = utils.NewSpinner("Installing Private Envoy Fleet...")
				err = installPrivateEnvoyFleet(helmPath, envoyFleetName, releaseNamespace)
				if err != nil {
					spinner.Fail("Installing Envoy Fleet: ", err)
					reportError(err)
					os.Exit(1)
				}
				spinner.Success()
			} else {
				pterm.Info.Println("--no-envoy-fleet set - skipping envoy fleet installation")
			}
		} else {
			pterm.Info.Println("Private Envoy Fleet already installed, skipping. To upgrade to a new version run `kusk upgrade`")
		}

		apiReleaseName := fmt.Sprintf("%s-api", releaseName)
		if _, apiInstalled := releases[apiReleaseName]; !apiInstalled {
			spinner = utils.NewSpinner("Installing Kusk API server...")
			err = installApi(helmPath, apiReleaseName, releaseNamespace, envoyFleetName)
			if err != nil {
				spinner.Fail("Installing Kusk API server: ", err)
				reportError(err)
				os.Exit(1)
			}

			spinner.Success()
		} else {
			pterm.Info.Println("api already installed, skipping. To upgrade to a new version run kusk upgrade")
		}

		if noDashboard {
			pterm.Info.Println("--no-dashboard set - skipping dashboard installation")
			printPortForwardInstructions("api", releaseNamespace, envoyFleetName)
			return
		}

		dashboardReleaseName := fmt.Sprintf("%s-dashboard", releaseName)
		if _, ok := releases[dashboardReleaseName]; !ok {
			spinner = utils.NewSpinner("Installing Kusk Dashboard...")
			err = installDashboard(helmPath, dashboardReleaseName, releaseNamespace, envoyFleetName)
			if err != nil {
				spinner.Fail("Installing Kusk Dashboard...", err)
				reportError(err)
				os.Exit(1)
			}

			spinner.Success()
		} else {
			pterm.Info.Println("Kusk Dashboard already installed, skipping. To upgrade to a new version run `kusk upgrade`")
		}
		printPortForwardInstructions("dashboard", releaseNamespace, envoyFleetName)
	},
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

func addKubeshopHelmRepo(helmPath string) error {
	commandArguments := getHelmCommandArguments("repo", "add", "kubeshop", "https://kubeshop.github.io/helm-charts/")

	out, err := process.Execute(
		helmPath,
		commandArguments...,
	)
	// if isLevelDebug() {
	pterm.Info.Printf("%v output:\n%v\n", helmPath+" "+strings.Join(commandArguments, " "), string(out))
	// }
	if err != nil && !strings.Contains(err.Error(), "Error: repository name (kubeshop) already exists, please specify a different name") {
		return err
	}

	return nil
}

func updateHelmRepos(helmPath string) error {
	commandArguments := getHelmCommandArguments("repo", "update")
	out, err := process.Execute(helmPath, commandArguments...)
	if isLevelDebug() {
		commandExecuted := helmPath + " " + strings.Join(commandArguments, " ")
		if err != nil {
			pterm.Info.Printf("%v output:\n%v\n", commandExecuted, string(out))
		} else {
			pterm.Info.Printf("%v output:\n%v\n%v error: \n%v\n", commandExecuted, string(out), commandExecuted, err)
		}
	}
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
	cmd.Env = os.Environ()

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

	if isLevelDebug() {
		commandExecuted := helmPath + " " + strings.Join(command, " ")
		pterm.Info.Printf("%v output:\n%v\n", commandExecuted, buffer.String())
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
	commandArguments := getHelmCommandArguments(
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
	)

	out, err := process.Execute(helmPath, commandArguments...)
	if err != nil {
		return err
	}

	if isLevelDebug() {
		pterm.Info.Printf("%v install kusk gateway output:\n%v\n", helmPath+" "+strings.Join(commandArguments, " "), string(out))
	}

	return nil
}

func installPublicEnvoyFleet(helmPath, releaseName, releaseNamespace string) error {
	return installEnvoyFleet(helmPath, releaseName, releaseNamespace, "LoadBalancer", true)
}

func installPrivateEnvoyFleet(helmPath, releaseName, releaseNamespace string) error {
	return installEnvoyFleet(helmPath, releaseName, releaseNamespace, "ClusterIP", false)
}

func installEnvoyFleet(helmPath, releaseName, releaseNamespace, serviceType string, isDefaultFleet bool) error {
	commandArguments := getHelmCommandArguments(
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
	)

	out, err := process.Execute(helmPath, commandArguments...)
	if err != nil {
		return err
	}

	if isLevelDebug() {
		pterm.Info.Printf("%v install envoy fleet output:\n%v\n", helmPath+" "+strings.Join(commandArguments, " "), string(out))
	}

	return nil
}

func installApi(helmPath, releaseName, releaseNamespace, envoyFleetName string) error {
	commandArguments := getHelmCommandArguments(
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
	)

	out, err := process.Execute(helmPath, commandArguments...)
	if err != nil {
		return err
	}

	if isLevelDebug() {
		pterm.Info.Printf("%v install api output:\n%v\n", helmPath+" "+strings.Join(commandArguments, " "), string(out))
	}

	return nil
}

func installDashboard(helmPath, releaseName, releaseNamespace, envoyFleetName string) error {
	commandArguments := getHelmCommandArguments(
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
	)

	out, err := process.Execute(helmPath, commandArguments...)
	if err != nil {
		return err
	}

	if isLevelDebug() {
		pterm.Info.Printf("%v install dashboard output:\n%v\n", helmPath+" "+strings.Join(commandArguments, " "), string(out))
	}

	return nil
}
