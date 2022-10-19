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
.
*/
package cmd

import (
	"context"
	"fmt"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/kuskui"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var removeGateway = &cobra.Command{
	Use:           "remove",
	Short:         "Removes selected instance of EnvoyFleet",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE:          removeRun,
}

func init() {
	gatewayCMD.AddCommand(removeGateway)
	removeGateway.Flags().BoolVar(&confirm, "no-confirm", false, "uninstall without prompt")
}

func removeRun(cmd *cobra.Command, args []string) error {
	c, err := utils.GetK8sClient()
	if err != nil {
		return err
	}

	fleets := &kuskv1.EnvoyFleetList{}
	if err := c.List(context.TODO(), fleets, &client.ListOptions{}); err != nil {
		return err
	}

	envoySelect := promptui.Select{
		Label: "Select Gateway you want to remove",
	}

	items := []string{}
	for _, f := range fleets.Items {
		items = append(items, fmt.Sprintf("%s/%s", f.Namespace, f.Name))
	}

	envoySelect.Items = items

	index, name, err := envoySelect.Run()
	if err != nil {
		return err
	}
	fmt.Println(name)
	proceed := confirm

	if !confirm {
		prompt := promptui.Prompt{
			Label:     "Removing Gateway instance will render all associated APIs unoperations. Are you sure you want to uninstall Gateway",
			IsConfirm: true,
		}
		result, err := prompt.Run()
		if err != nil {
			return nil
		}

		if result == "N" || result == "n" || result == "" {
			kuskui.PrintInfo("Exiting...")
			return nil
		}
		proceed = true
	}

	if proceed {
		if err := c.Delete(cmd.Context(), &fleets.Items[index], &client.DeleteAllOfOptions{}); err != nil {
			return err
		}
		fmt.Printf("%s fleet removed\n", fleets.Items[index].Name)

	}

	return nil
}
