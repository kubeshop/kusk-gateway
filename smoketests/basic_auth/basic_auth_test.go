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

package basic_auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/smoketests/common"
	"github.com/kubeshop/kusk-gateway/smoketests/helpers"
)

const (
	testName      = "test-basic-auth"
	testNamespace = "default"
)

type BasicAuthCheckSuite struct {
	common.KuskTestSuite
	api *kuskv1.API
}

func (t *BasicAuthCheckSuite) SetupTest() {
	rawApi := common.ReadFile("./basic_auth_api.yaml")
	api := &kuskv1.API{}
	t.NoError(yaml.Unmarshal([]byte(rawApi), api))

	api.ObjectMeta.Name = testName
	api.ObjectMeta.Namespace = testNamespace
	api.Spec.Fleet.Name = helpers.APIFleetName
	api.Spec.Fleet.Namespace = helpers.APIFleetNamespace

	if err := t.Cli.Create(context.TODO(), api, &client.CreateOptions{}); err != nil {
		message := fmt.Sprintf("apis.gateway.kusk.io %q already exists", testName)
		if strings.Contains(err.Error(), message) {
			t.T().Logf("WARNING: err=%v, message=%v", err, message)
			t.api = api
			return
		}

		t.Fail(err.Error(), nil)
	}

	t.api = api

	// weird way to wait it out probably needs to be done dynamically
	t.T().Logf("Sleeping for %s", helpers.WaitBeforeStartingTest)
	time.Sleep(helpers.WaitBeforeStartingTest)
}

func (t *BasicAuthCheckSuite) TearDownSuite() {
	t.NoError(t.Cli.Delete(context.Background(), t.api, &client.DeleteOptions{}))
}

func (t *BasicAuthCheckSuite) TestAuthorized() {
	loadBalancerIP := helpers.GetEnvoyFleetServiceLoadBalancerIP(&t.KuskTestSuite)
	url := fmt.Sprintf("http://%s/hello", loadBalancerIP)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	t.NoError(err)
	req.SetBasicAuth("kusk", "kusk")

	resp, err := http.DefaultClient.Do(req)
	t.NoError(err)

	defer resp.Body.Close()
	t.Equal(http.StatusOK, resp.StatusCode)
}

func (t *BasicAuthCheckSuite) TestUnauthorized() {
	loadBalancerIP := helpers.GetEnvoyFleetServiceLoadBalancerIP(&t.KuskTestSuite)
	url := fmt.Sprintf("http://%s/hello", loadBalancerIP)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	t.NoError(err)
	req.SetBasicAuth("kusk", "kusk123")

	resp, err := http.DefaultClient.Do(req)
	t.NoError(err)

	defer resp.Body.Close()
	t.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (t *BasicAuthCheckSuite) TestUnauthorizedNoCredentials() {
	loadBalancerIP := helpers.GetEnvoyFleetServiceLoadBalancerIP(&t.KuskTestSuite)
	url := fmt.Sprintf("http://%s/hello", loadBalancerIP)

	resp, err := http.Get(url)
	t.NoError(err)

	defer resp.Body.Close()

	t.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func TestBasicAuthCheckSuite(t *testing.T) {
	b := BasicAuthCheckSuite{}
	suite.Run(t, &b)
}
