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

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	testName      = "test-cache"
	testNamespace = "default"
)

type CacheTestSuite struct {
	common.KuskTestSuite
	api *kuskv1.API
}

func (t *CacheTestSuite) SetupTest() {
	rawApi := common.ReadFile("./cache-api.yaml")
	api := &kuskv1.API{}
	t.NoError(yaml.Unmarshal([]byte(rawApi), api))

	api.ObjectMeta.Name = testName
	api.ObjectMeta.Namespace = testNamespace
	api.Spec.Fleet.Name = helpers.APIFleetName
	api.Spec.Fleet.Namespace = helpers.APIFleetNamespace

	if err := t.Cli.Create(context.Background(), api, &client.CreateOptions{}); err != nil {
		message := fmt.Sprintf("apis.gateway.kusk.io %q already exists", testName)
		if strings.Contains(err.Error(), message) {
			t.T().Logf("WARNING: message=%v, err=%v", message, err)
			t.api = api // store `api` for deletion later
			return
		}

		t.Fail(err.Error())
	}

	t.api = api // store `api` for deletion later

	// weird way to wait it out probably needs to be done dynamically
	t.T().Logf("Sleeping for %s", helpers.WaitBeforeStartingTest)
	time.Sleep(helpers.WaitBeforeStartingTest)
}

func (t *CacheTestSuite) TearDownSuite() {
	t.NoError(t.Cli.Delete(context.Background(), t.api, &client.DeleteOptions{}))
}

func (t *CacheTestSuite) TestCacheCacheOn() {
	// We are expecting `cache-control: max-age=2` header and the value of `age` to increase over time.
	// the `uuid` will be cached up until `NoOfRequests`.
	const (
		NoOfRequests = 2
	)

	loadBalancerIP := helpers.GetEnvoyFleetServiceLoadBalancerIP(&t.KuskTestSuite)
	url := fmt.Sprintf("http://%s/uuid", loadBalancerIP)

	uuidCached := ""
	// Do n requests that will get cached, i.e., they'll return the same uuid.
	for x := 0; x < NoOfRequests; x++ {
		func() {
			t.T().Logf("iteration=%v, uuidCached=%v", x, uuidCached)

			req, err := http.NewRequest(http.MethodGet, url, nil)

			t.NoError(err)

			res, err := http.DefaultClient.Do(req)
			t.NoError(err)
			t.Equal(http.StatusOK, res.StatusCode)

			defer func() {
				t.NoError(res.Body.Close())
			}()

			responseBody, err := io.ReadAll(res.Body)
			t.NoError(err)

			body := map[string]string{}
			t.NoError(json.Unmarshal(responseBody, &body))

			if body["uuid"] == "" {
				t.Fail("uuid is empty - expecting a uuid")
			}

			// On first iteration store `uuid` as `uuidCached`.
			if x == 0 {
				uuidCached = body["uuid"]
			}

			if uuidCached != body["uuid"] {
				t.Fail("uuid has changed - expecting the same uuid")
			}

			time.Sleep(2 * time.Second)
		}()
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	t.NoError(err)

	res, err := http.DefaultClient.Do(req)
	t.NoError(err)
	t.Equal(http.StatusOK, res.StatusCode)

	defer func() {
		t.NoError(res.Body.Close())
	}()

	responseBody, err := io.ReadAll(res.Body)
	t.NoError(err)

	body := map[string]string{}
	t.NoError(json.Unmarshal(responseBody, &body))

	t.NotEqual(body["uuid"], uuidCached)
}

func TestCacheTestSuite(t *testing.T) {
	testSuite := CacheTestSuite{}
	suite.Run(t, &testSuite)
}
