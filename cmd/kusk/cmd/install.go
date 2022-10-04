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
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/avast/retry-go/v3"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"

	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/kuskui"
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

	$ kusk cluster install

	Will install kusk-gateway, a public (for your APIS) and private (for the kusk dashboard and api)
	envoy-fleet, api, and dashboard in the kusk-system namespace using helm.

	$ kusk install --latest

	Will pull the latest version of kusk available

	$ kusk cluster install --no-dashboard --no-api --no-envoy-fleet

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

		var (
			dir string
			err error
		)
		if !latest {
			if dir, err = getManifests(); err != nil {
				return err
			}
		} else {
			if dir, err = getManifestsFromUrl(); err != nil {
				return err
			}
		}

		kuskui.PrintStart("installing kusk...")

		c, err := utils.GetK8sClient()
		if err != nil {
			reportError(err)
			return err
		}

		if err := applyk(dir); err != nil {
			kuskui.PrintError("❌ failed installing Kusk")
			reportError(err)
			return err
		}
		namespace := kusknamespace
		name := "kusk-gateway-manager"

		if err := utils.WaitForPodsReady(cmd.Context(), c, namespace, name, 10*time.Minute, "component"); err != nil {
			kuskui.PrintError("failed installing Kusk")
			reportError(err)
			return err
		}

		if err := utils.WaitForDeploymentReady(cmd.Context(), c, namespace, name, 10*time.Minute); err != nil {
			kuskui.PrintError("failed installing Kusk")
			reportError(err)
			return err
		}

		if err := utils.WaitAPIServiceReady(cmd.Context(), c); err != nil {
			kuskui.PrintError("❌ failed installing Kusk")
			reportError(err)
			return err
		}

		if !noEnvoyFleet {
			kuskui.PrintStart("installing Envoyfleet...")
			fleet := &kuskv1.EnvoyFleet{}

			if err := provisionFromManifest(cmd.Context(), c, fleet, filepath.Join(dir, manifests_dir, "fleets.yaml")); err != nil {
				kuskui.PrintError("failed installing Envoyfleets")
				reportError(err)
				return err
			}

			if err := utils.WaitForPodsReady(cmd.Context(), c, namespace, "envoy", time.Duration(5*time.Minute), "component"); err != nil {
				kuskui.PrintError("failed installing Envoyfleets")
				reportError(err)
				return err
			}
		} else {
			kuskui.PrintSuccess("\ninstallation complete\n")
			printPortForwardInstructions(noEnvoyFleet)
			return nil
		}

		if !noApi {
			kuskui.PrintStart("installing API Server...")
			api := &kuskv1.API{}

			if err := provisionFromManifest(cmd.Context(), c, api, filepath.Join(dir, manifests_dir, "api_server_api.yaml")); err != nil {
				kuskui.PrintError("failed installing API Server")
				reportError(err)
				return err
			}

			if err := applyf(filepath.Join(dir, manifests_dir, "api_server.yaml")); err != nil {
				kuskui.PrintError("failed installing API Server")
				reportError(err)
				return err
			}
			if err := utils.WaitForPodsReady(cmd.Context(), c, namespace, kuskgatewayapi, time.Duration(5*time.Minute), "instance"); err != nil {
				kuskui.PrintError("failed installing API Server")
				reportError(err)
				return err
			}
		} else {
			kuskui.PrintSuccess("\ninstallation complete\n")
			printPortForwardInstructions(noApi)
			return nil
		}

		if !noDashboard {
			kuskui.PrintStart("installing Dashboard...")
			sr := &kuskv1.StaticRoute{}
			if err := provisionFromManifest(cmd.Context(), c, sr, filepath.Join(dir, manifests_dir, "dashboard_staticroute.yaml")); err != nil {
				kuskui.PrintError("failed installing Dashboard")
				reportError(err)
				return err
			}

			fleet := &kuskv1.EnvoyFleet{}
			if err := provisionFromManifest(cmd.Context(), c, fleet, filepath.Join(dir, manifests_dir, "dashboard_envoyfleet.yaml")); err != nil {
				kuskui.PrintError("failed installing Dashboard")
				reportError(err)
				return err
			}

			if err := applyf(filepath.Join(dir, manifests_dir, "dashboard.yaml")); err != nil {
				kuskui.PrintError("failed installing Dashboard")
				reportError(err)
				return err
			}
			if err := utils.WaitForPodsReady(cmd.Context(), c, namespace, kuskgatewaydashboard, time.Duration(5*time.Minute), "instance"); err != nil {
				kuskui.PrintError("failed installing kusk")
				reportError(err)
				return err
			}
		}

		fmt.Println("")
		kuskui.PrintSuccess("installation complete\n")
		printPortForwardInstructions(noDashboard)

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

func printPortForwardInstructions(noDashboard bool) {
	if !noDashboard {
		kuskui.PrintInfoGray("💡 Access the dashboard by using the following command")
		kuskui.PrintInfo("👉 kusk dashboard\n")
	}

	kuskui.PrintInfoGray("💡 Deploy your first API")
	kuskui.PrintInfo("👉 kusk deploy -i <path or url to your api definition>\n")

	kuskui.PrintInfoGray("💡 Access Help and useful examples to help get you started")
	kuskui.PrintInfo("👉 kusk --help")
}

func getManifestsFromUrl() (string, error) {
	githubClient, err := utils.NewGithubClient("", nil)
	if err != nil {
		return "", err
	}

	latest, err := githubClient.GetLatest(kuskgateway)
	if err != nil {
		return "", err

	}

	if latest == "v1.2.3" {
		return "", fmt.Errorf("you are trying to update to %s which isn't supported with `kusk cluster install --latest`", latest)
	}
	fullURLFile := fmt.Sprintf("https://github.com/kubeshop/kusk-gateway/archive/refs/tags/%s.zip", latest)

	dir := os.TempDir()

	file, err := downloadFile(dir, fullURLFile)
	if err != nil {
		return "", err
	}

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
			}

			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
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

func downloadFile(dir, fullURLFile string) (string, error) {
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

	resp, err := (&http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}).Get(fullURLFile)

	// Put content on file
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
			if err := os.MkdirAll(filepath.Join(tmpdir, dir), 0700); err != nil {
				return "", err
			}
		}

		f, err := os.Create(filepath.Join(tmpdir, name))
		if err != nil {
			return "", nil
		}

		content, _ := Asset(name)
		if strings.Contains(name, "configmap.yaml") {
			tmp := strings.Replace(string(content), `ANALYTICS_ENABLED: "true"`, fmt.Sprintf(`ANALYTICS_ENABLED: "%s"`, analyticsEnabled), -1)
			content = []byte(tmp)
		}
		if _, err := f.Write(content); err != nil {
			return "", nil
		}
	}

	return tmpdir, nil
}

// path = filepath.Join(dir, manifests_dir, "fleets.yaml")
func provisionFromManifest(ctx context.Context, c client.Client, obj client.Object, path string) error {
	manifest, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(manifest, obj); err != nil {
		return err
	}

	err = retry.Do(
		func() error {
			err := c.Create(ctx, obj, &client.CreateOptions{})
			if err != nil && strings.Contains(err.Error(), "already exists") {
				return nil
			}
			return err
		},
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			return 2 * time.Second
		}))

	if err != nil {
		return err
	}
	return nil
}
