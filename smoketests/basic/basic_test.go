package basic

import (
	"context"
	"testing"

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
	testName         = "test"
	testPort         = 82
)

type BasicCheckSuite struct {
	common.KuskTestSuite
}

func (b *BasicCheckSuite) SetupTest() {
	// time.Sleep(10 * time.Second) //crude way to wait it out

	rawFleet := common.ReadFile("envoyfleet.yaml")
	fleet := &kuskv1.EnvoyFleet{}

	b.NoError(yaml.Unmarshal([]byte(rawFleet), fleet))

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

	b.NoError(b.Cli.Create(context.TODO(), fleet, &client.CreateOptions{}))

	rawApi := common.ReadFile("api.yaml")
	api := &kuskv1.API{}
	b.NoError(yaml.Unmarshal([]byte(rawApi), api))

	api.ObjectMeta.Name = testName + "api"
	api.ObjectMeta.Namespace = defaultNamespace
	api.Spec.Fleet.Name = fleet.ObjectMeta.Name
	api.Spec.Fleet.Namespace = fleet.ObjectMeta.Namespace
	b.NoError(b.Cli.Create(context.TODO(), api, &client.CreateOptions{}))
}

func (b *BasicCheckSuite) TestGetAPI() {
	api := &kuskv1.API{}
	b.NoError(b.Cli.Get(context.TODO(), client.ObjectKey{
		Namespace: defaultNamespace,
		Name:      testName + "api",
	}, api))
	b.Equal(api.Name, testName+"api")
}

func (b *BasicCheckSuite) TestGetFleet() {
	api := &kuskv1.EnvoyFleet{}
	b.NoError(b.Cli.Get(context.TODO(), client.ObjectKey{
		Namespace: defaultNamespace,
		Name:      testName + "fleet",
	}, api))
	b.Equal(api.Name, testName+"fleet")
}

func (b *BasicCheckSuite) TearDownTest() {
	api := &kuskv1.API{
		ObjectMeta: v1.ObjectMeta{
			Name:      testName + "api",
			Namespace: defaultNamespace,
		},
	}

	b.NoError(b.Cli.Delete(context.TODO(), api, &client.DeleteOptions{}))

	fleet := &kuskv1.EnvoyFleet{
		ObjectMeta: v1.ObjectMeta{
			Name:      testName + "fleet",
			Namespace: defaultNamespace,
		},
	}
	b.NoError(b.Cli.Delete(context.TODO(), fleet, &client.DeleteOptions{}))
}

func TestBasichCheckSuite(t *testing.T) {
	b := BasicCheckSuite{}
	suite.Run(t, &b)
}
