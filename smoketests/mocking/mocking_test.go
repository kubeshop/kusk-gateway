package mocking

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	testName         = "mock-test"
	testPort         = 82
)

type MockCheckSuite struct {
	common.KuskTestSuite
}

func (m *MockCheckSuite) SetupTest() {
	rawApi := common.ReadFile("../samples/hello-world/mock-api.yaml")
	api := &kuskv1.API{}
	m.NoError(yaml.Unmarshal([]byte(rawApi), api))

	api.ObjectMeta.Name = testName
	api.ObjectMeta.Namespace = defaultNamespace
	api.Spec.Fleet.Name = defaultName
	api.Spec.Fleet.Namespace = defaultNamespace

	m.NoError(m.Cli.Create(context.TODO(), api, &client.CreateOptions{}))
	time.Sleep(1 * time.Second) // weird way to wait it out probably needs to be done dynamically
}

func (m *MockCheckSuite) TestEndpoint() {
	envoyFleetSvc := &corev1.Service{}
	m.NoError(
		m.Cli.Get(context.TODO(), client.ObjectKey{Name: defaultName, Namespace: defaultNamespace}, envoyFleetSvc),
	)
	resp, err := http.Get(fmt.Sprintf("http://%s/hello", envoyFleetSvc.Status.LoadBalancer.Ingress[0].IP))
	m.NoError(err)

	defer resp.Body.Close()

	m.Equal(200, resp.StatusCode)

	o, _ := io.ReadAll(resp.Body)
	res := map[string]string{}
	m.NoError(json.Unmarshal(o, &res))

	m.Equal("Hello from a mocked response!", res["message"])

}

func (m *MockCheckSuite) TearDownTest() {
	api := &kuskv1.API{
		ObjectMeta: v1.ObjectMeta{
			Name:      testName,
			Namespace: defaultNamespace,
		},
	}

	m.NoError(m.Cli.Delete(context.TODO(), api, &client.DeleteOptions{}))

	fleet := &kuskv1.EnvoyFleet{
		ObjectMeta: v1.ObjectMeta{
			Name:      defaultName,
			Namespace: defaultNamespace,
		},
	}
	m.NoError(m.Cli.Delete(context.TODO(), fleet, &client.DeleteOptions{}))
}

func TestMockingSuite(t *testing.T) {
	b := MockCheckSuite{}
	suite.Run(t, &b)
}
