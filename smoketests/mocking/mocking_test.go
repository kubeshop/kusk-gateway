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
	"sigs.k8s.io/controller-runtime/pkg/client"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/kubeshop/kusk-gateway/smoketests/common"
	"github.com/kubeshop/kusk-gateway/smoketests/helpers"
)

const (
	testName      = "test-mock"
	testNamespace = "default"
)

type MockCheckSuite struct {
	common.KuskTestSuite
	api *kuskv1.API
}

func (t *MockCheckSuite) SetupTest() {
	rawApi := common.ReadFile("./mock-api.yaml")
	api := &kuskv1.API{}
	t.NoError(yaml.Unmarshal([]byte(rawApi), api))

	api.ObjectMeta.Name = testName
	api.ObjectMeta.Namespace = testNamespace
	api.Spec.Fleet.Name = helpers.APIFleetName
	api.Spec.Fleet.Namespace = helpers.APIFleetNamespace

	t.NoError(t.Cli.Create(context.TODO(), api, &client.CreateOptions{}))
	// store `api` for deletion later
	t.api = api

	// weird way to wait it out probably needs to be done dynamically
	t.T().Logf("Sleeping for %s", helpers.WaitBeforeStartingTest)
	time.Sleep(helpers.WaitBeforeStartingTest)
}

func (t *MockCheckSuite) TearDownSuite() {
	t.NoError(t.Cli.Delete(context.Background(), t.api, &client.DeleteOptions{}))
}

func (t *MockCheckSuite) TestEndpoint() {
	const (
		ContentTypeKey      = "content-type"
		expectedContentType = "application/json"
	)

	loadBalancerIP := helpers.GetEnvoyFleetServiceLoadBalancerIP(&t.KuskTestSuite)
	url := fmt.Sprintf("http://%s/hello", loadBalancerIP)
	resp, err := http.Get(url)
	t.NoError(err)

	defer resp.Body.Close()

	t.Equal(200, resp.StatusCode)

	actualContentType := resp.Header.Get(ContentTypeKey)
	t.T().Logf("%s=%v", ContentTypeKey, actualContentType)
	t.Equal(expectedContentType, actualContentType)

	o, _ := io.ReadAll(resp.Body)
	res := map[string]string{}
	t.NoError(json.Unmarshal(o, &res))

	t.Equal("Hello from a mocked response!", res["message"])
}

func TestMockingSuite(t *testing.T) {
	b := MockCheckSuite{}
	suite.Run(t, &b)
}
