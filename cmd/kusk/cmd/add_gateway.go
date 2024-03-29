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
.
*/
package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
)

var (
	gatewayName      string
	svcType          string
	defaultGateway   bool
	gatewayNamespace string
)

var addGatewayCMD = &cobra.Command{
	Use:           "add",
	Aliases:       []string{"create"},
	Short:         "Installs instance of Envoyfleet",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          addRun,
	PreRunE: func(cmd *cobra.Command, args []string) error {

		if len(svcType) > 0 {
			if svcType != "ClusterIP" && svcType != "LoadBalancer" {
				return fmt.Errorf("svcType values can only be ClusterIP or LoadBalancer")
			}
		}

		if len(gatewayName) > 0 {
			errs := validation.IsQualifiedName(gatewayName)
			if len(errs) > 0 {
				return fmt.Errorf(strings.Join(errs, ","))
			}
		}

		if defaultGateway {
			c, err := utils.GetK8sClient()
			fleets := &kuskv1.EnvoyFleetList{}

			if err != nil {
				return err
			}

			if err := c.List(context.TODO(), fleets, &client.ListOptions{}); err != nil {
				return err
			}

			for _, f := range fleets.Items {
				if f.Spec.Default {
					return fmt.Errorf("there is already a default gateway in you cluster")
				}
			}
		}
		return nil
	},
}

func init() {
	gatewayCMD.AddCommand(addGatewayCMD)

	addGatewayCMD.Flags().StringVarP(&gatewayNamespace, "namespace", "n", "", "namespace where the new gateway will be created")
	addGatewayCMD.Flags().StringVarP(&svcType, "serviceType", "s", "", "Service type of the gateway. Supported options LoadBalancer, ClusterIP")
	addGatewayCMD.Flags().StringVarP(&gatewayName, "name", "", "", "name of the gateway")
	addGatewayCMD.Flags().BoolVarP(&defaultGateway, "default", "", false, "Indicates if the geteway is the default gateway in the cluster")
}

func addRun(cmd *cobra.Command, args []string) error {
	reportError := func(err error) {
		if err != nil {
			errors.NewErrorReporter(cmd, err).Report()
		}
	}

	var err error
	fleet := &kuskv1.EnvoyFleet{}
	fleets := &kuskv1.EnvoyFleetList{}

	c, err := utils.GetK8sClient()
	if err != nil {
		reportError(err)
		return err
	}

	if len(gatewayName) == 0 {
		gatewayName, err = namePrompt.Run()
		if err != nil {
			fmt.Println(err)
		}
	}

	fleet.Name = gatewayName

	if len(gatewayName) == 0 {
		return nil
	}

	if err := c.List(context.TODO(), fleets, &client.ListOptions{}); err != nil {
		return err
	}
	thereIsDefault := false
	for _, f := range fleets.Items {
		if f.Spec.Default {
			thereIsDefault = true
			break
		}
	}

	if !thereIsDefault {
		_, t, _ := defaultPrompt.Run()
		deflt, _ := strconv.ParseBool(t)
		fleet.Spec.Default = deflt
	}

	if len(gatewayNamespace) == 0 {
		namespaces := &corev1.NamespaceList{}
		if err := c.List(cmd.Context(), namespaces, &client.ListOptions{}); err != nil {
			reportError(err)
			return err
		}

		namespacesPrompt := promptui.Select{
			Label: "Select a namespace",
		}

		namespaceNames := []string{}
		for _, ns := range namespaces.Items {
			namespaceNames = append(namespaceNames, ns.Name)
		}
		namespacesPrompt.Items = namespaceNames
		_, gatewayNamespace, _ = namespacesPrompt.Run()
	}

	fleet.Namespace = gatewayNamespace

	if len(svcType) == 0 {
		_, svcType, _ = serviceTypePrompt.Run()
	}
	fleet.Spec.Service = &kuskv1.ServiceConfig{
		Type: corev1.ServiceType(svcType),
	}

	fleet.Spec.Service.Ports = []corev1.ServicePort{
		{
			Port: 80,
		},
	}

	if err := c.Create(cmd.Context(), fleet, &client.CreateOptions{}); err != nil {
		reportError(err)
		return err
	}

	fmt.Printf("%s fleet created\n", fleet.Name)
	return nil
}

var defaultPrompt = promptui.Select{
	Label: "Do you want your gateway to be the default in the cluster",
	Items: []bool{true, false},
}
var serviceTypePrompt = promptui.Select{
	Label: "Pick service type you want to use",
	Items: []string{"LoadBalancer", "ClusterIP"},
}

var namePrompt = promptui.Prompt{
	Label: "Please input name for the new gateway instance",
}
