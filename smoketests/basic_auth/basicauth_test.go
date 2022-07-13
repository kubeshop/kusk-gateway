package mocking

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
		if strings.Contains(err.Error(), `apis.gateway.kusk.io "auth-test" already exists`) {
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
	req.SetBasicAuth("kubeshop", "kubeshop")

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
	req.SetBasicAuth("kubeshop", "kubeshop123")

	resp, err := http.DefaultClient.Do(req)
	m.NoError(err)

	defer resp.Body.Close()
	m.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (m *BasicAuthCheckSuite) TestForbidden() {
	envoyFleetSvc := &corev1.Service{}
	m.NoError(
		m.Cli.Get(context.TODO(), client.ObjectKey{Name: defaultName, Namespace: defaultNamespace}, envoyFleetSvc),
	)
	resp, err := http.Get(fmt.Sprintf("http://%s/hello", envoyFleetSvc.Status.LoadBalancer.Ingress[0].IP))
	m.NoError(err)

	defer resp.Body.Close()

	m.Equal(http.StatusForbidden, resp.StatusCode)
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
