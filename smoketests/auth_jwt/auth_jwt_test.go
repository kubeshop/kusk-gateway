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

package auth_jwt

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
	"sigs.k8s.io/controller-runtime/pkg/client"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/smoketests/common"
	"github.com/kubeshop/kusk-gateway/smoketests/helpers"
)

const (
	testName      = "test-auth-jwt-oauth0"
	testNamespace = "default"
)

const (
	HeaderAuthorization = "Authorization"
)

type AuthJWTTestSuite struct {
	common.KuskTestSuite
	api *kuskv1.API
}

func TestAuthJWTTestSuite(t *testing.T) {
	testSuite := AuthJWTTestSuite{}
	suite.Run(t, &testSuite)
}

func (t *AuthJWTTestSuite) SetupTest() {
	rawApi := common.ReadFile("../../examples/auth/jwt/oauth0/api.yaml")
	api := &kuskv1.API{}
	t.NoError(yaml.Unmarshal([]byte(rawApi), api))

	api.ObjectMeta.Name = testName
	api.ObjectMeta.Namespace = testNamespace
	api.Spec.Fleet.Name = helpers.APIFleetName
	api.Spec.Fleet.Namespace = helpers.APIFleetNamespace

	if err := t.Cli.Create(context.Background(), api, &client.CreateOptions{}); err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("apis.gateway.kusk.io %q already exists", testName)) {
			t.api = api
			return
		}

		t.Fail(err.Error())
	}

	t.api = api // store `api` for deletion later

	// weird way to wait it out probably needs to be done dynamically
	t.T().Logf("Sleeping for %s", helpers.WaitBeforeStartingTest)
	time.Sleep(helpers.WaitBeforeStartingTest)
}

func (t *AuthJWTTestSuite) TearDownSuite() {
	t.NoError(t.Cli.Delete(context.Background(), t.api, &client.DeleteOptions{}))
}

func (t *AuthJWTTestSuite) Test_Auth_JWT_Invalid() {
	// Calling a protected route without a valid bearer token should result in an error.
	const (
		expected = "Jwt is not in the form of Header.Payload.Signature with two dots and 3 sections"
	)

	loadBalancerIP := helpers.GetEnvoyFleetServiceLoadBalancerIP(&t.KuskTestSuite)
	url := fmt.Sprintf("http://%s/uuid", loadBalancerIP)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	request.Header.Add(HeaderAuthorization, "Bearer <invalid_token>")
	t.NoError(err)

	client := makeHTTPClient()
	response, err := client.Do(request)
	t.NoError(err)
	t.Equal(http.StatusUnauthorized, response.StatusCode)

	responseBody, err := io.ReadAll(response.Body)
	t.NoError(err)
	actual := string(responseBody)

	defer func() {
		t.NoError(response.Body.Close())
	}()

	t.Equal(expected, actual)
}

func makeHTTPClient() *http.Client {
	client := &http.Client{}
	return client
}
