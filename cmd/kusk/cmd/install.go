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
	"path/filepath"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
)

var (
	noApi            bool
	noDashboard      bool
	noEnvoyFleet     bool
	latest           bool
	analyticsEnabled = "true"
)

func init() {
	clusterCmd.AddCommand(installCmd)
	installCmd.Flags().StringVar(&releaseName, "name", "kusk-gateway", "installation name")
	installCmd.Flags().StringVar(&releaseNamespace, "namespace", "kusk-system", "namespace to install in")

	installCmd.Flags().BoolVar(&noDashboard, "no-dashboard", false, "don't the install dashboard")
	installCmd.Flags().BoolVar(&noApi, "no-api", false, "don't install the api. Setting this flag implies --no-dashboard")
	installCmd.Flags().BoolVar(&noEnvoyFleet, "no-envoy-fleet", false, "don't install any envoy fleets")
	installCmd.Flags().BoolVar(&latest, "latest", false, "if set latest version of kusk-gateway will be installed")

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
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		reportError := func(err error) {
			if err != nil {
				errors.NewErrorReporter(cmd, err).Report()
			}
		}

		dir, err := RestoreManifests()
		if err != nil {
			return err
		}

		spinner := utils.NewSpinner("Installing kusk")

		if err := applyk(dir); err != nil {
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

			if err := applyf("fleets.yaml"); err != nil {
				spinner.Fail("failed installing Envoy Fleets", err)
				reportError(err)
				return err
			}
			spinner.Success()
		}

		if !noApi {
			spinner = utils.NewSpinner("Installing API Server...")

			if err := applyf("apis.yaml"); err != nil {
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

			if err := applyf("dashboard.yaml"); err != nil {
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

// invokes kubectl apply -f
func applyf(filename string) error {
	instCmd := NewKubectlCmd()
	if !latest {
		instCmd.SetArgs([]string{"apply", fmt.Sprintf("-f=%s", getEmbeddedFile(filename))})
	} else {
		ghclient, _ := utils.NewGithubClient("", nil)
		i, _, err := ghclient.GetTags()
		if err != nil {
			return err
		}
		if len(i) > 0 {
			ref_str := strings.Split(i[len(i)-1].Ref, "/")
			ref := ref_str[len(ref_str)-1]
			url := fmt.Sprintf("https://raw.githubusercontent.com/kubeshop/kusk-gateway/v%s/cmd/kusk/cmd/manifests/%s", ref, filename)
			instCmd.SetArgs([]string{"apply", fmt.Sprintf("-f=%s", url)})
		}
	}
	return instCmd.Execute()
}

// invokes kubectl apply -k
func applyk(filename string) error {
	instCmd := NewKubectlCmd()
	if !latest {
		instCmd.SetArgs([]string{"apply", fmt.Sprintf("-k=%s", filepath.Join(filename, "/config/default"))})
	} else {
		// figure out a way to download latest manifests from GH
		ghclient, _ := utils.NewGithubClient("", nil)
		i, _, err := ghclient.GetTags()
		if err != nil {
			return err
		}
		if len(i) > 0 {
			ref_str := strings.Split(i[len(i)-1].Ref, "/")
			ref := ref_str[len(ref_str)-1]
			url := fmt.Sprintf("https://raw.githubusercontent.com/kubeshop/kusk-gateway/v%s/cmd/kusk/cmd/manifests/%s", ref, filename)
			instCmd.SetArgs([]string{"apply", fmt.Sprintf("-f=%s", url)})
		}
	}
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

func RestoreManifests() (string, error) {
	tmpdir, _ := ioutil.TempDir("", "kusk-man")

	manifestNames := AssetNames()
	for _, name := range manifestNames {
		dir := filepath.Dir(name)
		if dir != "." {
			os.MkdirAll(filepath.Join(tmpdir, dir), 0700)
		}
		if f, err := os.Create(filepath.Join(tmpdir, name)); err != nil {
			return "", nil
		} else {
			content, _ := Asset(name)
			if _, err := f.Write(content); err != nil {
				return "", nil
			}
		}
	}

	return tmpdir, nil
}
