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

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"

	"github.com/kubeshop/kusk-gateway/smoketests/common"
	"github.com/kubeshop/kusk-gateway/smoketests/helpers"
)

const (
	testName      = "test-traffic-splitting-api"
	testNamespace = "default"
)

type WeightedClusterTestSuite struct {
	common.KuskTestSuite
	api *kuskv1.API
}

func TestWeightedClusterTestSuite(t *testing.T) {
	testSuite := WeightedClusterTestSuite{}
	suite.Run(t, &testSuite)
}

func (t *WeightedClusterTestSuite) TearDownSuite() {
	t.NoError(t.Cli.Delete(context.Background(), t.api, &client.DeleteOptions{}))
}

func (t *WeightedClusterTestSuite) SetupTest() {
	rawApi := common.ReadFile("./weighted-api.yaml")
	api := &kuskv1.API{}
	t.NoError(yaml.Unmarshal([]byte(rawApi), api))

	api.ObjectMeta.Name = testName
	api.ObjectMeta.Namespace = testNamespace
	api.Spec.Fleet.Name = helpers.APIFleetName
	api.Spec.Fleet.Namespace = helpers.APIFleetNamespace

	if err := t.Cli.Create(context.Background(), api, &client.CreateOptions{}); err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("apis.gateway.kusk.io %q already exists", testName)) {
			// store `api` for deletion later
			t.api = api
			return
		}

		t.Fail(err.Error())
	}

	// store `api` for deletion later
	t.api = api

	// weird way to wait it out probably needs to be done dynamically
	t.T().Logf("Sleeping for %s", helpers.WaitBeforeStartingTest)
	time.Sleep(helpers.WaitBeforeStartingTest)
}

func (t *WeightedClusterTestSuite) Test_WeightedCluster() {
	loadBalancerIP := helpers.GetEnvoyFleetServiceLoadBalancerIP(&t.KuskTestSuite)
	url := fmt.Sprintf("http://%s/uuid", loadBalancerIP)

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
