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

package openapi_path

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/smoketests/common"
)

const (
	testName         = "test-openapi-path-with-auth"
	defaultName      = "kusk-gateway-envoy-fleet"
	defaultNamespace = "kusk-system"
)

type OpenAPIPathTestSuite struct {
	common.KuskTestSuite
	api *kuskv1.API
}

func (t *OpenAPIPathTestSuite) SetupTest() {
	rawApi := common.ReadFile("../samples/openapi-path/openapi-path-with-auth.yaml")
	api := &kuskv1.API{}
	t.NoError(yaml.Unmarshal([]byte(rawApi), api))

	api.ObjectMeta.Name = testName
	api.ObjectMeta.Namespace = "default"
	api.Spec.Fleet.Name = defaultName
	api.Spec.Fleet.Namespace = defaultNamespace

	if err := t.Cli.Create(context.Background(), api, &client.CreateOptions{}); err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("apis.gateway.kusk.io %q already exists", testName)) {
			return
		}

		t.Fail(err.Error())
	}

	t.api = api // store `api` for deletion later

	duration := 5 * time.Second
	t.T().Logf("Sleeping for %s", duration)
	time.Sleep(duration) // weird way to wait it out probably needs to be done dynamically
}

func (t *OpenAPIPathTestSuite) TestOpenAPIPathWithAuthOK() {
	// Calling `/openapi.json` should return the below:
	bodyExpected := `
{
    "components": {},
    "info": {
        "description": "test-openapi-path-with-auth",
        "title": "test-openapi-path-with-auth",
        "version": "0.0.1"
    },
    "openapi": "3.0.0",
    "paths": {
        "/": {
            "get": {
                "description": "Returns GET data.",
                "operationId": "/get",
                "responses": {}
            }
        },
        "/uuid": {
            "get": {
                "description": "Returns UUID4.",
                "operationId": "/uuid",
                "responses": {}
            }
        }
    },
    "schemes": [
        "http",
        "https"
    ]
}
	`

	envoyFleetSvc := getEnvoyFleetSvc(&t.KuskTestSuite)
	url := fmt.Sprintf("http://%s/openapi.json", envoyFleetSvc.Status.LoadBalancer.Ingress[0].IP)

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

	t.JSONEq(bodyExpected, string(responseBody))
}

func (t *OpenAPIPathTestSuite) TestOpenAPIPathWithAuthUnauthorized() {
	envoyFleetSvc := getEnvoyFleetSvc(&t.KuskTestSuite)
	url := fmt.Sprintf("http://%s/uuid", envoyFleetSvc.Status.LoadBalancer.Ingress[0].IP)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	t.NoError(err)

	res, err := http.DefaultClient.Do(req)
	t.NoError(err)

	defer func() {
		t.NoError(res.Body.Close())
	}()

	t.Equal(http.StatusUnauthorized, res.StatusCode)
}

func (t *OpenAPIPathTestSuite) TearDownSuite() {
	t.NoError(t.Cli.Delete(context.Background(), t.api, &client.DeleteOptions{}))
}

func TestOpenAPIPathTestSuite(t *testing.T) {
	testSuite := OpenAPIPathTestSuite{}
	suite.Run(t, &testSuite)
}

func getEnvoyFleetSvc(t *common.KuskTestSuite) *corev1.Service {
	t.T().Helper()

	envoyFleetSvc := &corev1.Service{}
	t.NoError(
		t.Cli.Get(context.Background(), client.ObjectKey{Name: defaultName, Namespace: defaultNamespace}, envoyFleetSvc),
	)

	return envoyFleetSvc
}
