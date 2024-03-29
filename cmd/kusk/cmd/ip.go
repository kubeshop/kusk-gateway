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

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/kuskui"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
)

var (
	ipEnvoyFleetName      string
	ipEnvoyFleetNamespace string
)

func init() {
	rootCmd.AddCommand(ipCmd)

	ipCmd.Flags().StringVarP(&ipEnvoyFleetName, "envoyfleet.name", "", "", "Envoy Fleet name")
	ipCmd.Flags().StringVarP(&ipEnvoyFleetNamespace, "envoyfleet.namespace", "", "kusk-system", "Envoy Fleet namespace")
}

var ipCmd = &cobra.Command{
	Use:           "ip",
	Short:         "return IP address of the default envoyfleet",
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		reportError := func(err error) {
			if err != nil {
				errors.NewErrorReporter(cmd, err).Report()
			}
		}

		cmd.SilenceUsage = true

		k8sclient, err := utils.GetK8sClient()
		if err != nil {
			reportError(err)
			kuskui.PrintError(err.Error())
			return
		}

		var envoyFleet *kuskv1.EnvoyFleet
		if ipEnvoyFleetName != "" {
			envoyFleet, err = getNamedEnvoyFleet(cmd.Context(), ipEnvoyFleetName, ipEnvoyFleetNamespace, k8sclient)
		} else {
			envoyFleet, err = getDefaultEnvoyFleet(cmd.Context(), k8sclient)
		}

		if err != nil {
			reportError(err)
			kuskui.PrintError(err.Error())
			return
		}

		list := &corev1.ServiceList{}

		labelSelector, err := labels.Parse("app.kubernetes.io/managed-by=kusk-gateway-manager,fleet=" + envoyFleet.Name + "." + envoyFleet.Namespace)
		if err != nil {
			reportError(err)
			kuskui.PrintError(err.Error())
			return
		}

		if err := k8sclient.List(context.TODO(), list, &client.ListOptions{LabelSelector: labelSelector}); err != nil {
			reportError(err)
			kuskui.PrintError(err.Error())
			return
		}

		ip := ""
		svc := corev1.Service{}
		if len(list.Items) > 0 {
			svc = list.Items[0]
			for _, s := range list.Items {
				if s.Spec.Type == "LoadBalancer" && len(s.Status.LoadBalancer.Ingress) > 0 {
					ip = s.Status.LoadBalancer.Ingress[0].IP
					break
				}
			}
		}
		//kubectl port-forward svc/kusk-gateway-dashboard -n kusk-system 8080:80
		if svc.Spec.Type == "ClusterIP" {
			kuskui.PrintWarning(fmt.Sprintf("EnvoyFleet doesn't have an External IP address assigned. Try port-forwarding by running: \n\n kubectl port-forward svc/%s -n %s 8080:%d", svc.Name, svc.Namespace, svc.Spec.Ports[0].Port))
			return
		}
		if ip == "" {
			kuskui.PrintWarning(fmt.Sprintf("EnvoyFleet doesn't have an External IP address assigned yet. Try port-forwarding by running: \n\n kubectl port-forward svc/%s -n %s 8080:%d", svc.Name, svc.Namespace, svc.Spec.Ports[0].Port))
			return
		}
		fmt.Println(ip)
	},
}

func getNamedEnvoyFleet(ctx context.Context, name, namespace string, k8sclient client.Client) (*kuskv1.EnvoyFleet, error) {
	envoyFleet := &kuskv1.EnvoyFleet{}
	if err := k8sclient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, envoyFleet); err != nil {
		return nil, fmt.Errorf("unable to get named envoy fleet %s in the %s namespace: %w", name, namespace, err)
	}
	return envoyFleet, nil
}

func getDefaultEnvoyFleet(ctx context.Context, k8sclient client.Client) (*kuskv1.EnvoyFleet, error) {
	envoyFleets := &kuskv1.EnvoyFleetList{}

	if err := k8sclient.List(ctx, envoyFleets, &client.ListOptions{}); err != nil {
		return nil, fmt.Errorf("unable to list fleets while trying to find default fleet: %w", err)
	}

	if len(envoyFleets.Items) == 0 {
		return nil, fmt.Errorf("there are no envoyfleets in your cluster")
	}

	defaultFleet := kuskv1.EnvoyFleet{}
	for _, f := range envoyFleets.Items {
		if f.Spec.Default {
			defaultFleet = f
			break
		}
	}

	if len(defaultFleet.Name) == 0 {
		return nil, fmt.Errorf("there is no default envoyfleet in your cluster")
	}

	return &defaultFleet, nil
}
