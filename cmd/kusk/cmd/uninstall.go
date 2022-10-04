package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/kuskui"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"

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
			c, err := utils.GetK8sClient()
			if err != nil {
				reportError(err)
				return err
			}

			kuskui.PrintStart("Checking if kusk is already installed...")

			kuskNamespace := &corev1.Namespace{}
			if err := c.Get(cmd.Context(), client.ObjectKey{Name: kusknamespace}, kuskNamespace); err != nil {
				if err.Error() == fmt.Sprintf(`namespaces "%s" not found`, kusknamespace) {
					kuskui.PrintInfo("Kusk is not installed on cluster.")
					os.Exit(0)
				}
				reportError(err)
				return err
			}

			var dir string
			if dir, err = getManifests(); err != nil {
				reportError(err)
				return err
			}

			apis := &kuskv1.APIList{}
			if err := c.List(cmd.Context(), apis, &client.ListOptions{}); err != nil {
				if err.Error() == `no matches for kind "API" in version "gateway.kusk.io/v1alpha1"` {
					kuskui.PrintInfo("Kusk Custom Resource Definition API is not installed.")
				} else {
					reportError(err)
					return err
				}
			}

			if apis != nil && len(apis.Items) > 0 {
				kuskui.PrintStart("deleting APIs...")
				for _, api := range apis.Items {
					if err := c.Delete(cmd.Context(), &api, &client.DeleteAllOfOptions{}); err != nil {
						reportError(err)
						return err
					}
				}
			}

			fleets := &kuskv1.EnvoyFleetList{}
			if err := c.List(cmd.Context(), fleets, &client.ListOptions{}); err != nil {
				if err.Error() == `no matches for kind "EnvoyFleet" in version "gateway.kusk.io/v1alpha1"` {
					kuskui.PrintInfo("Kusk Custom Resource Definition EnvoyFleet is not installed.")
				} else {
					reportError(err)
					return err
				}
			}

			if fleets != nil && len(fleets.Items) > 0 {
				kuskui.PrintStart("deleting Envoyfleets...")
				for _, fleet := range fleets.Items {
					if err := c.Delete(cmd.Context(), &fleet, &client.DeleteAllOfOptions{}); err != nil {
						reportError(err)
						return err
					}
				}
			}

			staticRoutes := &kuskv1.StaticRouteList{}
			if err := c.List(cmd.Context(), staticRoutes, &client.ListOptions{}); err != nil {
				if err.Error() == `no matches for kind "StaticRoute" in version "gateway.kusk.io/v1alpha1"` {
					kuskui.PrintInfo("Kusk Custom Resource Definition StaticRouote is not installed")
				} else {
					reportError(err)
					return err
				}
			}

			if staticRoutes != nil && len(staticRoutes.Items) > 0 {
				kuskui.PrintStart("deleting Staticroutes...")
				for _, route := range staticRoutes.Items {
					if err := c.Delete(cmd.Context(), &route, &client.DeleteAllOfOptions{}); err != nil {
						reportError(err)
						return err
					}
				}
			}

			kuskui.PrintStart("uninstalling Kusk...")

			if err := deletek(dir); err != nil {
				kuskui.PrintError("❌ failed uninstalling Kusk", err.Error())
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
