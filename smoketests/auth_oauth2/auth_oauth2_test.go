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

// Note these tests just do the following, as we cannot go through the entire OAuth2 `authorization_code` `grant_type` without using a browser or user interaction:
//
// 1. Checks that the API spec for oauth2 can be applied successfully.
// 2.
package auth_oauth2

import (
	"context"
	"fmt"
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
	testName          = "test-auth-oauth2-oauth0-authorization-code-grant"
	testNamespace     = "default"
	apiFleetName      = "kusk-gateway-envoy-fleet"
	apiFleetNamespace = "kusk-system"
)

type AuthOAuth2TestSuite struct {
	common.KuskTestSuite
	api *kuskv1.API
}

func TestAuthOAuth2TestSuite(t *testing.T) {
	testSuite := AuthOAuth2TestSuite{}
	suite.Run(t, &testSuite)
}

func (t *AuthOAuth2TestSuite) SetupTest() {
	rawApi := common.ReadFile("../../examples/auth/oauth2/authorization-code-grant/api.yaml")
	api := &kuskv1.API{}
	t.NoError(yaml.Unmarshal([]byte(rawApi), api))

	api.ObjectMeta.Name = testName
	api.ObjectMeta.Namespace = testNamespace
	api.Spec.Fleet.Name = apiFleetName
	api.Spec.Fleet.Namespace = apiFleetNamespace

	if err := t.Cli.Create(context.Background(), api, &client.CreateOptions{}); err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("apis.gateway.kusk.io %q already exists", testName)) {
			return
		}

		t.Fail(err.Error())
	}

	t.api = api // store `api` for deletion later

	duration := 4 * time.Second
	t.T().Logf("Sleeping for %s", duration)
	time.Sleep(duration) // weird way to wait it out probably needs to be done dynamically
}

func (t *AuthOAuth2TestSuite) TearDownSuite() {
	t.NoError(t.Cli.Delete(context.Background(), t.api, &client.DeleteOptions{}))
}

func (t *AuthOAuth2TestSuite) TestUUIDPathReturnsARedirect() {
	// Calling `/uuid` should return the below:
	// Actualy redirect is `"https://kubeshop-kusk-gateway-oauth2.eu.auth0.com/authorize?client_id=upRN78W8GzV4TwFRp0ekZfLx2UnqJJs8&scope=openid&response_type=code&redirect_uri=http%3A%2F%2F192.168.49.2%2Foauth2%2Fcallback&state=http%3A%2F%2F192.168.49.2%2Fuuid"`
	// but we'll ignore the rest of the string and just do a contains comparison instead of equals.
	redirectExpected := `https://kubeshop-kusk-gateway-oauth2.eu.auth0.com/authorize?client_id=upRN78W8GzV4TwFRp0ekZfLx2UnqJJs8&scope=openid&response_type=code`

	envoyFleetSvc := getEnvoyFleetSvc(&t.KuskTestSuite)
	url := fmt.Sprintf("http://%s/uuid", envoyFleetSvc.Status.LoadBalancer.Ingress[0].IP)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	t.NoError(err)

	client := makeHTTPClient()
	response, err := client.Do(request)
	t.NoError(err)
	t.Equal(http.StatusFound, response.StatusCode)

	defer func() {
		t.NoError(response.Body.Close())
	}()

	redirect, err := response.Location()
	t.NoError(err)
	t.Contains(redirect.String(), redirectExpected)
}

func (t *AuthOAuth2TestSuite) TestOAuth2IsDisabledOnRootPath() {
	// Calling `/` should be fine even if OAuth2 is configured as the root path is un-protected.

	envoyFleetSvc := getEnvoyFleetSvc(&t.KuskTestSuite)
	url := fmt.Sprintf("http://%s/", envoyFleetSvc.Status.LoadBalancer.Ingress[0].IP)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	t.NoError(err)

	client := makeHTTPClient()
	response, err := client.Do(request)
	t.NoError(err)
	t.Equal(http.StatusOK, response.StatusCode)

	defer func() {
		t.NoError(response.Body.Close())
	}()
}

func (t *AuthOAuth2TestSuite) Test_OAuth2_StaticRoute() {
	t.T().Skipf("%s skipping - TODO(MBana): Implement test", t.T().Name())
}

func getEnvoyFleetSvc(t *common.KuskTestSuite) *corev1.Service {
	t.T().Helper()

	envoyFleetSvc := &corev1.Service{}
	t.NoError(
		t.Cli.Get(
			context.Background(),
			client.ObjectKey{Name: apiFleetName, Namespace: apiFleetNamespace},
			envoyFleetSvc,
		),
	)

	return envoyFleetSvc
}

// makeHTTPClient
// Make the Go HTTP Client NOT Follow Redirects Automatically.
func makeHTTPClient() *http.Client {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return client
}
