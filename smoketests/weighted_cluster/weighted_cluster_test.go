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

package weighted_cluster

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"

	"github.com/kubeshop/kusk-gateway/smoketests/common"
)

const (
	testName         = "test-traffic-splitting-api-1"
	defaultName      = "test-traffic-splitting"
	defaultNamespace = "default"
	port             = 89
)

type WeightedClusterTestSuite struct {
	common.KuskTestSuite
	api   *kuskv1.API
	fleet *kuskv1.EnvoyFleet
}

func TestWeightedClusterTestSuite(t *testing.T) {
	testSuite := WeightedClusterTestSuite{}
	suite.Run(t, &testSuite)
}

func (t *WeightedClusterTestSuite) TearDownSuite() {
	t.NoError(t.Cli.Delete(context.Background(), t.api, &client.DeleteOptions{}))
	t.NoError(t.Cli.Delete(context.Background(), t.fleet, &client.DeleteOptions{}))
}

func (t *WeightedClusterTestSuite) SetupTest() {
	rawApi := common.ReadFile("../samples/weighted/weighted-api.yaml")

	rawFleet := common.ReadFile("../basic/envoyfleet.yaml")
	fleet := &kuskv1.EnvoyFleet{}
	t.NoError(yaml.Unmarshal([]byte(rawFleet), fleet))

	fleet.ObjectMeta.Name = defaultName
	fleet.ObjectMeta.Namespace = defaultNamespace
	fleet.Spec.Service = &kuskv1.ServiceConfig{
		Type: corev1.ServiceTypeLoadBalancer,
		Ports: []corev1.ServicePort{
			{
				Port:       port,
				TargetPort: intstr.FromString("http"),
				Name:       "http",
			},
			{
				Port:       444,
				TargetPort: intstr.FromString("http"),
				Name:       "https",
			},
		},
	}

	t.NoError(t.Cli.Create(context.TODO(), fleet, &client.CreateOptions{}))

	t.fleet = fleet

	api := &kuskv1.API{}
	t.NoError(yaml.Unmarshal([]byte(rawApi), api))

	api.ObjectMeta.Name = defaultName
	api.ObjectMeta.Namespace = defaultNamespace
	api.Spec.Fleet.Name = defaultName
	api.Spec.Fleet.Namespace = defaultNamespace

	if err := t.Cli.Create(context.Background(), api, &client.CreateOptions{}); err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("apis.gateway.kusk.io %q already exists", testName)) {
			return
		}

		t.Fail(err.Error())
	}

	t.api = api // store `api` for deletion later

	duration := 60 * time.Second
	t.T().Logf("Waiting for %s", duration)
	t.NoError(common.WaitForServiceReady(context.TODO(), t.Cli, defaultNamespace, defaultName, duration))
}

func (t *WeightedClusterTestSuite) Test_WeightedCluster() {
	envoyFleetSvc := getEnvoyFleetSvc(&t.KuskTestSuite)
	url := fmt.Sprintf("http://%s:%d/uuids", envoyFleetSvc.Status.LoadBalancer.Ingress[0].IP, port)

	// Once the servicesHitCounts becomes 1 for all the services below, we break from the for loop and terminate the test.
	servicesHitCounts := map[string]int{
		"traffic-splitting-httpbin-2.default.svc.cluster.local.-80": 0,
		"traffic-splitting-httpbin-1.default.svc.cluster.local.-80": 0,
	}

	for {
		request, err := http.NewRequest(http.MethodGet, url, nil)
		t.NoError(err)

		client := &http.Client{}
		response, err := client.Do(request)
		t.NoError(err)
		t.Equal(http.StatusOK, response.StatusCode)

		defer func() {
			t.NoError(response.Body.Close())
		}()

		actual := response.Header.Get("x-kusk-weighted-cluster")
		t.T().Logf("`x-kusk-weighted-cluster`=%v", actual)

		t.NotEqual("", actual)

		// increment the count
		servicesHitCounts[actual]++

		allServicesHit := true
		for _, count := range servicesHitCounts {
			if count <= 0 {
				allServicesHit = false
			}
		}
		if allServicesHit {
			break
		}

		// time.Sleep(time.Second * 2)
	}
}

func getEnvoyFleetSvc(t *common.KuskTestSuite) *corev1.Service {
	t.T().Helper()

	envoyFleetSvc := &corev1.Service{}
	t.NoError(
		t.Cli.Get(
			context.Background(),
			client.ObjectKey{Name: defaultName, Namespace: defaultNamespace},
			envoyFleetSvc,
		),
	)

	return envoyFleetSvc
}
