package mocking

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kubeshop/kusk-gateway/smoketests/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	rawFleet := common.ReadFile("envoyfleet.yaml")
	fleet := &kuskv1.EnvoyFleet{}

	m.NoError(yaml.Unmarshal([]byte(rawFleet), fleet))

	fleet.ObjectMeta.Namespace = defaultNamespace
	fleet.ObjectMeta.Name = testName + "fleet"
	fleet.Spec.Service = &kuskv1.ServiceConfig{
		Type: corev1.ServiceTypeLoadBalancer,
		Ports: []corev1.ServicePort{
			{
				Port:       testPort,
				TargetPort: intstr.FromString("http"),
				Name:       "http",
			},
			{
				Port:       444,
				TargetPort: intstr.FromString("http"),
				Name:       "https",
			},
		},
	}
	m.NoError(m.Cli.Create(context.TODO(), fleet, &client.CreateOptions{}))

	rawApi := common.ReadFile("../samples/hello-world/mock-api.yaml")
	api := &kuskv1.API{}
	m.NoError(yaml.Unmarshal([]byte(rawApi), api))

	api.ObjectMeta.Name = testName
	api.ObjectMeta.Namespace = defaultNamespace
	api.Spec.Fleet.Name = testName + "fleet"
	api.Spec.Fleet.Namespace = defaultNamespace
	fmt.Println("*****", api)
	fmt.Println("*****")
	m.NoError(m.Cli.Create(context.TODO(), api, &client.CreateOptions{}))
}

func (m *MockCheckSuite) TestEndpoint() {
	time.Sleep(5 * time.Second)
	resp, err := http.Get("http://127.0.0.1/hello")
	m.NoError(err)

	defer resp.Body.Close()

	m.Equal(200, resp.StatusCode)

	o, _ := io.ReadAll(resp.Body)
	res := map[string]string{}
	fmt.Println("******", string(o))
	m.NoError(json.Unmarshal(o, &res))

	m.Equal(res["message"], "Hello from a mocked response!")

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
			Name:      testName + "fleet",
			Namespace: defaultNamespace,
		},
	}
	m.NoError(m.Cli.Delete(context.TODO(), fleet, &client.DeleteOptions{}))
}

func TestMockingSuite(t *testing.T) {
	b := MockCheckSuite{}
	suite.Run(t, &b)
}
