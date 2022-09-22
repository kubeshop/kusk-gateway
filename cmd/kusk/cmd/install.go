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
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
)

var (
	noApi            bool
	noDashboard      bool
	noEnvoyFleet     bool
	analyticsEnabled = "true"
	latest           bool
)

func init() {
	clusterCmd.AddCommand(installCmd)

	installCmd.Flags().BoolVar(&noDashboard, "no-dashboard", false, "don't the install dashboard")
	installCmd.Flags().BoolVar(&noApi, "no-api", false, "don't install the api. Setting this flag implies --no-dashboard")
	installCmd.Flags().BoolVar(&noEnvoyFleet, "no-envoy-fleet", false, "don't install any envoy fleets")
	installCmd.Flags().BoolVar(&latest, "latest", false, "get latest Kusk version from Github")
	if enabled, ok := os.LookupEnv("ANALYTICS_ENABLED"); ok {
		analyticsEnabled = enabled
	}
}

const manifests_dir = "/cmd/kusk/manifests"

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
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		reportError := func(err error) {
			if err != nil {
				errors.NewErrorReporter(cmd, err).Report()
			}
		}

		var err error
		var dir string
		if !latest {
			if dir, err = getManifests(); err != nil {
				return err
			}
		} else {
			if dir, err = getManifestsFromUrl(); err != nil {
				return err
			}
		}
		fmt.Println("✅ Installing Kusk...")

		if err := applyk(dir); err != nil {
			fmt.Println("❌ failed installing Kusk", err)
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

		if err := utils.WaitForPodsReady(cmd.Context(), c, namespace, name, time.Duration(5*time.Minute), "component"); err != nil {
			fmt.Println("❌ failed installing kusk", err)
			reportError(err)
			return err
		}
		fmt.Println("✅ Kusk installed")

		if !noEnvoyFleet {
			if err := applyf(filepath.Join(dir, manifests_dir, "fleets.yaml")); err != nil {
				fmt.Println("❌ failed installing Envoy Fleets", err)
				reportError(err)
				return err
			}
			fmt.Println("✅ Envoy Fleets installed")
		} else {
			return nil
		}

		if err := utils.WaitForPodsReady(cmd.Context(), c, namespace, "envoy", time.Duration(5*time.Minute), "component"); err != nil {
			fmt.Println("❌ failed installing kusk", err)
			reportError(err)
			return err
		}

		if !noApi {
			if err := applyf(filepath.Join(dir, manifests_dir, "apis.yaml")); err != nil {
				fmt.Println("❌ failed installing API Server", err)
				reportError(err)
				return err
			}
			fmt.Println("✅ API Server installed")
		} else if noApi {
			return nil
		}

		if err := utils.WaitForPodsReady(cmd.Context(), c, namespace, "kusk-gateway-api", time.Duration(5*time.Minute), "instance"); err != nil {
			fmt.Println("❌ failed installing kusk", err)
			reportError(err)
			return err
		}

		if !noDashboard {
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
			printPortForwardInstructions("dashboard", releaseNamespace, envoyFleetName)
		}

		return nil
	},
}

// invokes kubectl apply -f
func applyf(filename string) error {
	instCmd := NewKubectlCmd()
	instCmd.SetArgs([]string{"apply", fmt.Sprintf("-f=%s", filename)})

	return instCmd.Execute()
}

// invokes kubectl apply -k
func applyk(filename string) error {
	instCmd := NewKubectlCmd()
	instCmd.SetArgs([]string{"apply", fmt.Sprintf("-k=%s", filepath.Join(filename, "/config/default"))})

	return instCmd.Execute()
}

func printPortForwardInstructions(service, releaseNamespace, envoyFleetName string) {
	if service == "dashboard" {
		fmt.Println("💡 Access the dashboard by using the following command")
		fmt.Println("👉 kusk dashboard")
	}

	fmt.Println("💡 Deploy your first API")
	fmt.Println("👉 kusk deploy -i <path or url to your api definition>")

	fmt.Println("💡 Access Help and useful examples to help get you started ")
	fmt.Println("👉 kusk --help")
}
func getManifestsFromUrl() (string, error) {
	ghclient, err := utils.NewGithubClient("", nil)
	if err != nil {
		return "", err
	}

	latest, err := ghclient.GetLatest()
	if err != nil {
		return "", err

	}

	fmt.Println("LATEST", latest)
	if latest == "v1.2.3" {
		return "", fmt.Errorf("you are trying to update to %s which isn't supported with `kusk cluster install --latest`", latest)
	}
	fullURLFile := fmt.Sprintf("https://github.com/kubeshop/kusk-gateway/archive/refs/tags/%s.zip", latest)

	dir := os.TempDir()

	file, err := donwloadFile(dir, fullURLFile)
	if err != nil {
		return "", err
	}

	fmt.Println(file)

	return unzip(file)
}

func unzip(path string) (string, error) {
	dir, _ := filepath.Split(path)
	dst := dir
	archive, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	var distroDir string
	defer archive.Close()

	for i, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			return "", fmt.Errorf("invalid file path")
		}
		if f.FileInfo().IsDir() {
			if i == 0 {
				distroDir = filePath
				fmt.Println("distroDir", distroDir)
			}

			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			fmt.Println("here")
			return "", err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", err
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return "", err
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return "", err
		}

		dstFile.Close()
		fileInArchive.Close()
	}
	return distroDir, nil
}

func donwloadFile(dir, fullURLFile string) (string, error) {
	// Build fileName from fullPath
	fileURL, err := url.Parse(fullURLFile)
	if err != nil {
		log.Fatal(err)
	}
	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]

	// Create blank file
	file, err := os.Create(filepath.Join(dir, fileName))
	if err != nil {
		return "", err
	}
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(fullURLFile)
	if err != nil {
		return "", err

	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	defer file.Close()

	return file.Name(), nil
}

func getManifests() (string, error) {
	tmpdir := os.TempDir()

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
			if strings.Contains(name, "configmap.yaml") {
				tmp := strings.Replace(string(content), `ANALYTICS_ENABLED: "true"`, fmt.Sprintf(`ANALYTICS_ENABLED: "%s"`, analyticsEnabled), -1)
				content = []byte(tmp)
			}
			if _, err := f.Write(content); err != nil {
				return "", nil
			}
		}
	}

	return tmpdir, nil
}
