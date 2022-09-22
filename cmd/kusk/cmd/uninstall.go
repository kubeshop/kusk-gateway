package cmd

import (
	"fmt"
	"path/filepath"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var confirm bool

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall kusk-gateway, envoy-fleet, api, and dashboard in a single command",
	Long: `
	Uninstall kusk-gateway, envoy-fleet, api, and dashboard in a single command.

	$ kusk uninstall

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

		proceed := confirm

		if !confirm {
			prompt := promptui.Prompt{
				Label:     "uninstall command is irrevirsible. Are you sure you want to proceed?",
				IsConfirm: true,
			}
			result, err := prompt.Run()
			if err != nil {
				return nil
			}

			if result == "N" || result == "n" || result == "" {
				fmt.Println("exiting")
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

			apis := &kuskv1.APIList{}
			if err := c.List(cmd.Context(), apis, &client.ListOptions{}); err != nil {
				reportError(err)
				if err.Error() == `no matches for kind "API" in version "gateway.kusk.io/v1alpha1"` {
					fmt.Println("Kusk Custom Resource Definition API is not installed skipping ")
				} else {
					return err
				}
			}

			fmt.Println("Deleting APIs...")
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
					fmt.Println("Kusk Custom Resource Definition API is not installed skipping ")
				} else {
					return err
				}
			}

			fmt.Println("Deleting fleets...")
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
					fmt.Println("Kusk Custom Resource Definition API is not installed skipping ")
				} else {
					return err
				}
			}

			fmt.Println("Deleting Static Routes...")
			for _, route := range staticRoutes.Items {
				if err := c.Delete(cmd.Context(), &route, &client.DeleteAllOfOptions{}); err != nil {
					reportError(err)
					return err
				}
			}

			deployments := appsv1.DeploymentList{}
			if err := c.List(cmd.Context(), &deployments, &client.ListOptions{Namespace: "kusk-system"}); err != nil {
				reportError(err)
				return err
			}

			for _, deploy := range deployments.Items {
				if deploy.Name == "kusk-gateway-manager" || deploy.Name == "kusk-gateway-private-envoy-fleet" || deploy.Name == "kusk-gateway-envoy-fleet" {
					continue
				}
				if err := c.Delete(cmd.Context(), &deploy, &client.DeleteAllOfOptions{}); err != nil {
					reportError(err)
					return err
				}
			}

			fmt.Println("Deleting Kusk Dashboard service")
			service := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kusk-gateway-dashboard",
					Namespace: "kusk-system",
				},
			}

			if err := c.Delete(cmd.Context(), service, &client.DeleteOptions{}); err != nil {
				reportError(err)
				fmt.Println(err)
			}

			fmt.Println("Uninstalling Kusk...")

			if err := deletek(dir); err != nil {
				fmt.Println("❌ failed uninstalling Kusk", err)
				reportError(err)
				return err
			}

		}
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
