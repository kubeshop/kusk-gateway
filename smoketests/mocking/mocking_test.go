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
	"sigs.k8s.io/controller-runtime/pkg/client"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/smoketests/common"
)

const (
	testName          = "test-mock"
	testNamespace     = "default"
	apiFleetName      = "kusk-gateway-envoy-fleet"
	apiFleetNamespace = "kusk-system"
	port              = 80
)

type MockCheckSuite struct {
	common.KuskTestSuite
	api *kuskv1.API
}

func (m *MockCheckSuite) SetupTest() {
	rawApi := common.ReadFile("./mock-api.yaml")
	api := &kuskv1.API{}
	m.NoError(yaml.Unmarshal([]byte(rawApi), api))

	api.ObjectMeta.Name = testName
	api.ObjectMeta.Namespace = testNamespace
	api.Spec.Fleet.Name = apiFleetName
	api.Spec.Fleet.Namespace = apiFleetNamespace

	m.NoError(m.Cli.Create(context.TODO(), api, &client.CreateOptions{}))
	m.api = api

	time.Sleep(3 * time.Second) // weird way to wait it out probably needs to be done dynamically
}

func (t *MockCheckSuite) TearDownSuite() {
	t.NoError(t.Cli.Delete(context.Background(), t.api, &client.DeleteOptions{}))
}

func (m *MockCheckSuite) TestEndpoint() {
	const (
		ContentTypeKey      = "content-type"
		expectedContentType = "application/json"
	)

	envoyFleetSvc := &corev1.Service{}
	m.NoError(
		m.Cli.Get(context.TODO(), client.ObjectKey{Name: apiFleetName, Namespace: apiFleetNamespace}, envoyFleetSvc),
	)
	resp, err := http.Get(fmt.Sprintf("http://%s/hello", envoyFleetSvc.Status.LoadBalancer.Ingress[0].IP))
	m.NoError(err)

	defer resp.Body.Close()

	m.Equal(200, resp.StatusCode)

	actualContentType := resp.Header.Get(ContentTypeKey)
	m.T().Logf("%s=%v", ContentTypeKey, actualContentType)
	m.Equal(expectedContentType, actualContentType)

	o, _ := io.ReadAll(resp.Body)
	res := map[string]string{}
	m.NoError(json.Unmarshal(o, &res))

	m.Equal("Hello from a mocked response!", res["message"])
}

func TestMockingSuite(t *testing.T) {
	b := MockCheckSuite{}
	suite.Run(t, &b)
}
