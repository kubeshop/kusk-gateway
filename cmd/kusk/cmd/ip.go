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
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
)

func init() {
	rootCmd.AddCommand(ipCmd)
}

var ipCmd = &cobra.Command{
	Use:           "ip",
	Short:         "return IP address of the default envoyfleet",
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		reportError := func(err error) {
			if err != nil {
				errors.NewErrorReporter(cmd, err).Report()
			}
		}

		cmd.SilenceUsage = true

		k8sclient, err := utils.GetK8sClient()
		if err != nil {
			reportError(err)
			return err
		}

		envoyFleets := &kuskv1.EnvoyFleetList{}

		if err := k8sclient.List(context.TODO(), envoyFleets, &client.ListOptions{}); err != nil {
			reportError(err)
			return err
		}
		if len(envoyFleets.Items) == 0 {
			err := fmt.Errorf("there are no envoyfleets in your cluster")
			reportError(err)
			return err
		}
		defaultFleet := kuskv1.EnvoyFleet{}
		for _, f := range envoyFleets.Items {
			if f.Spec.Default {
				defaultFleet = f
				break
			}
		}

		if len(defaultFleet.Name) == 0 {
			err := fmt.Errorf("there is no default envoyfleet in your cluster")
			reportError(err)
		}

		list := &corev1.ServiceList{}

		labelSelector, err := labels.Parse("app.kubernetes.io/managed-by=kusk-gateway-manager,fleet=" + defaultFleet.Name + "." + defaultFleet.Namespace)
		if err != nil {
			reportError(err)
			return err
		}

		if err := k8sclient.List(context.TODO(), list, &client.ListOptions{LabelSelector: labelSelector}); err != nil {
			reportError(err)
			return err
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
			err := fmt.Errorf("your envoyfleet doesn't have public IP address assigned. Try portforwarding %q", fmt.Sprintf("kubectl port-forward svc/%s  -n %s 8080:%d", svc.Name, svc.Namespace, svc.Spec.Ports[0].Port))
			reportError(err)
			return err
		}
		if ip == "" {
			err := fmt.Errorf("your envoyfleet doesn't have public IP address assigned yet retry or try portforwarding %q", fmt.Sprintf("kubectl port-forward svc/%s  -n %s 8080:%d", svc.Name, svc.Namespace, svc.Spec.Ports[0].Port))
			reportError(err)
			return err

		}
		fmt.Println(ip)
		return nil
	},
}
