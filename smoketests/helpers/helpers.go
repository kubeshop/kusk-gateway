// MIT License
//
// Copyright (c) 2022 Kubeshop
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package helpers

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubeshop/kusk-gateway/smoketests/common"
)

const (
	WaitBeforeStartingTest = 2 * time.Second
)

/*
	    spec:
	      fleet:
	        name: kusk-gateway-envoy-fleet
		      namespace: kusk-system
*/
const (
	APIFleetName      = "kusk-gateway-envoy-fleet"
	APIFleetNamespace = "kusk-system"
)

func GetEnvoyFleetServiceLoadBalancerIP(t *common.KuskTestSuite) string {
	t.T().Helper()
	apiFleetName := APIFleetName
	apiFleetNamespace := APIFleetNamespace

	envoyFleetService := &corev1.Service{}
	key := client.ObjectKey{Name: apiFleetName, Namespace: apiFleetNamespace}
	err := t.Cli.Get(context.Background(), key, envoyFleetService)
	t.NoError(err)

	t.NotNil(envoyFleetService.Status)
	t.NotNil(envoyFleetService.Status.LoadBalancer)
	t.NotNil(envoyFleetService.Status.LoadBalancer.Ingress)
	hasIngressIP := len(envoyFleetService.Status.LoadBalancer.Ingress) >= 1
	t.True(hasIngressIP)

	return envoyFleetService.Status.LoadBalancer.Ingress[0].IP
}
