package cmd

import (
	"fmt"
	"path/filepath"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/kuskui"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var confirm bool

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall kusk-gateway, envoy-fleet, api, and dashboard in a single command",
	Long: `
	Uninstall kusk-gateway, envoy-fleet, api, and dashboard in a single command.

	$ kusk uninstall
	`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		reportError := func(err error) {
			if err != nil {
				errors.NewErrorReporter(cmd, err).Report()
			}
		}

		proceed := confirm

		if !confirm {
			prompt := promptui.Prompt{
				Label:     "Are you sure you want to uninstall kusk",
				IsConfirm: true,
			}
			result, err := prompt.Run()
			if err != nil {
				return nil
			}

			if result == "N" || result == "n" || result == "" {
				kuskui.PrintInfo("exiting")
				return nil
			}
			proceed = true
		}

		if proceed {
			var err error
			var dir string
			if dir, err = getManifests(); err != nil {
				return err
			}

			c, err := utils.GetK8sClient()
			if err != nil {
				reportError(err)
				return err
			}

			if err := deletef(filepath.Join(dir, manifests_dir, "fleets.yaml")); err != nil {
				fmt.Println("❌ failed uninstalling Envoy Fleets", err)
				reportError(err)
				return err
			}

			if err := deletef(filepath.Join(dir, manifests_dir, "api_server_api.yaml")); err != nil {
				fmt.Println("❌ failed uninstalling APIs", err)
				reportError(err)
				return err
			}

			if err := deletef(filepath.Join(dir, manifests_dir, "api_server.yaml")); err != nil {
				fmt.Println("❌ failed uninstalling APIs", err)
				reportError(err)
				return err
			}

			kuskui.PrintStart("deleting Dashboard")
			if err := deletef(filepath.Join(dir, manifests_dir, "dashboard_envoyfleet.yaml")); err != nil {
				fmt.Println("❌ failed uninstalling dashboard", err)
				reportError(err)
				return err
			}

			if err := deletef(filepath.Join(dir, manifests_dir, "dashboard_staticroute.yaml")); err != nil {
				fmt.Println("❌ failed uninstalling dashboard", err)
				reportError(err)
				return err
			}

			if err := deletef(filepath.Join(dir, manifests_dir, "dashboard.yaml")); err != nil {
				fmt.Println("❌ failed uninstalling dashboard", err)
				reportError(err)
				return err
			}

			apis := &kuskv1.APIList{}
			if err := c.List(cmd.Context(), apis, &client.ListOptions{}); err != nil {
				reportError(err)
				if err.Error() == `no matches for kind "API" in version "gateway.kusk.io/v1alpha1"` {
					kuskui.PrintInfo("Kusk Custom Resource Definition API is not installed skipping ")
				} else {
					return err
				}
			}

			kuskui.PrintStart("deleting APIs...")
			for _, api := range apis.Items {
				if err := c.Delete(cmd.Context(), &api, &client.DeleteAllOfOptions{}); err != nil {
					reportError(err)
					return err
				}
			}

			fleets := &kuskv1.EnvoyFleetList{}
			if err := c.List(cmd.Context(), fleets, &client.ListOptions{}); err != nil {
				reportError(err)
				if err.Error() == `no matches for kind "EnvoyFleet" in version "gateway.kusk.io/v1alpha1"` {
					kuskui.PrintInfo("Kusk Custom Resource Definition EnvoyFleet is not installed skipping ")
				} else {
					return err
				}
			}

			kuskui.PrintStart("deleting Envoyfleets...")
			for _, fleet := range fleets.Items {
				if err := c.Delete(cmd.Context(), &fleet, &client.DeleteAllOfOptions{}); err != nil {
					reportError(err)
					return err
				}
			}

			staticRoutes := &kuskv1.StaticRouteList{}
			if err := c.List(cmd.Context(), staticRoutes, &client.ListOptions{}); err != nil {
				reportError(err)
				if err.Error() == `no matches for kind "StaticRoute" in version "gateway.kusk.io/v1alpha1"` {
					kuskui.PrintInfo("Kusk Custom Resource Definition StaticRouote is not installed skipping ")
				} else {
					return err
				}
			}

			kuskui.PrintStart("deleting Staticroutes...")
			for _, route := range staticRoutes.Items {
				if err := c.Delete(cmd.Context(), &route, &client.DeleteAllOfOptions{}); err != nil {
					reportError(err)
					return err
				}
			}

			fmt.Println("Uninstalling Kusk...")

			if err := deletek(dir); err != nil {
				fmt.Println("❌ failed uninstalling Kusk", err)
				reportError(err)
				return err
			}

		}

		kuskui.PrintInfoLightGreen("\nkusk successfully uninstalled from your cluster")

		return nil
	},
}

func init() {
	clusterCmd.AddCommand(uninstallCmd)

	uninstallCmd.Flags().BoolVar(&confirm, "no-confirm", false, "uninstall without prompt")
}

func deletek(filename string) error {
	instCmd := NewKubectlCmd()
	instCmd.SetArgs([]string{"delete", fmt.Sprintf("-k=%s", filepath.Join(filename, "/config/default"))})

	return instCmd.Execute()
}

func deletef(filename string) error {
	instCmd := NewKubectlCmd()
	instCmd.SetArgs([]string{"delete", fmt.Sprintf("-f=%s", filename)})

	return instCmd.Execute()
}
