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
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/smoketests/common"
)

const (
	defaultNamespace = "default"
	defaultName      = "default"
	testName         = "auth-test"
	testPort         = 82
)

type BasicAuthCheckSuite struct {
	common.KuskTestSuite
}

func (m *BasicAuthCheckSuite) SetupTest() {
	rawApi := common.ReadFile("../samples/hello-world/auth-api.yaml")
	api := &kuskv1.API{}
	m.NoError(yaml.Unmarshal([]byte(rawApi), api))

	api.ObjectMeta.Name = testName
	api.ObjectMeta.Namespace = defaultNamespace
	api.Spec.Fleet.Name = defaultName
	api.Spec.Fleet.Namespace = defaultNamespace

	if err := m.Cli.Create(context.TODO(), api, &client.CreateOptions{}); err != nil {
		message := fmt.Sprintf("apis.gateway.kusk.io %q already exists", testName)
		m.T().Log(message)

		if strings.Contains(err.Error(), message) {
			return
		}

		m.Fail(err.Error(), nil)
	}

	duration := 20 * time.Second
	m.T().Logf("sleeping for %s", duration)
	time.Sleep(duration) // weird way to wait it out probably needs to be done dynamically
}

func (m *BasicAuthCheckSuite) TestAuthorized() {
	envoyFleetSvc := &corev1.Service{}
	m.NoError(
		m.Cli.Get(context.TODO(), client.ObjectKey{Name: defaultName, Namespace: defaultNamespace}, envoyFleetSvc),
	)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/hello", envoyFleetSvc.Status.LoadBalancer.Ingress[0].IP), nil)
	m.NoError(err)
	req.SetBasicAuth("kusk", "kusk")

	resp, err := http.DefaultClient.Do(req)
	m.NoError(err)

	defer resp.Body.Close()
	m.Equal(http.StatusOK, resp.StatusCode)
}

func (m *BasicAuthCheckSuite) TestUnauthorized() {
	envoyFleetSvc := &corev1.Service{}
	m.NoError(
		m.Cli.Get(context.TODO(), client.ObjectKey{Name: defaultName, Namespace: defaultNamespace}, envoyFleetSvc),
	)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/hello", envoyFleetSvc.Status.LoadBalancer.Ingress[0].IP), nil)
	m.NoError(err)
	req.SetBasicAuth("kusk", "kusk123")

	resp, err := http.DefaultClient.Do(req)
	m.NoError(err)

	defer resp.Body.Close()
	m.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (m *BasicAuthCheckSuite) TestUnauthorizedNoCredentials() {
	envoyFleetSvc := &corev1.Service{}
	m.NoError(
		m.Cli.Get(context.TODO(), client.ObjectKey{Name: defaultName, Namespace: defaultNamespace}, envoyFleetSvc),
	)
	resp, err := http.Get(fmt.Sprintf("http://%s/hello", envoyFleetSvc.Status.LoadBalancer.Ingress[0].IP))
	m.NoError(err)

	defer resp.Body.Close()

	m.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (m *BasicAuthCheckSuite) TearDownSuite() {
	api := &kuskv1.API{
		ObjectMeta: v1.ObjectMeta{
			Name:      testName,
			Namespace: defaultNamespace,
		},
	}
	m.NoError(m.Cli.Delete(context.TODO(), api, &client.DeleteOptions{}))
}

func TestBasicAuthCheckSuite(t *testing.T) {
	b := BasicAuthCheckSuite{}
	suite.Run(t, &b)
}
