package kusk

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
)

func TestCreateEnvoyFleet(t *testing.T) {
	assertions := require.New(t)
	fleet := kuskv1.EnvoyFleet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: kuskv1.EnvoyFleetSpec{
			Service: &kuskv1.ServiceConfig{
				Type: corev1.ServiceTypeLoadBalancer,
			},
		},
	}

	_, err := NewClient(getFakeClient()).CreateFleet(fleet)
	assertions.NoError(err)
}

func TestDeleteFleet(t *testing.T) {
	assertions := require.New(t)

	fleet := kuskv1.EnvoyFleet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "default",
		},
	}
	testClient := NewClient(getFakeClient())
	assertions.NoError(testClient.DeleteFleet(fleet))
}
func TestClientGetEnvoyFleets(t *testing.T) {
	assertions := require.New(t)
	testClient := NewClient(getFakeClient())
	fleets, err := testClient.GetEnvoyFleets()
	assertions.NoError(err)

	assertions.NotEqual(len(fleets.Items), 0)
}

func TestClientGetEnvoyFleet(t *testing.T) {
	name := "default"
	namespace := "default"
	testClient := NewClient(getFakeClient())
	fleet, err := testClient.GetEnvoyFleet(namespace, name)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf(`envoyfleet.gateway.kusk.io "%s" not found`, name)) {
			t.Error(err)
			t.Fail()
			return
		}
		t.Error(err)
		t.Fail()
		return
	}

	if fleet.ObjectMeta.Name != name {
		t.Error("name does not match")
		t.Fail()
		return
	}
}

func TestGetApis(t *testing.T) {
	testClient := NewClient(getFakeClient())
	apis, err := testClient.GetApis("default")
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	fmt.Println(len(apis.Items))
}

func TestGetApi(t *testing.T) {
	testClient := NewClient(getFakeClient())
	api, err := testClient.GetApi("default", "sample")
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	fmt.Println(api.Spec.Spec)
}
func TestGetNotFoundApi(t *testing.T) {
	testClient := NewClient(getFakeClient())
	_, err := testClient.GetApi("default", "not-found")
	if err == ErrNotFound {
		return
	}
	t.Errorf("%s does not equal to ErrNotFound", err)
	t.Fail()
}

func TestDeleteAPI(t *testing.T) {
	assertions := require.New(t)
	testClient := NewClient(getFakeClient())
	err := testClient.DeleteAPI("default", "sample")
	assertions.NoError(err)
}

func TestUpdateAPI(t *testing.T) {
	assertions := require.New(t)

	tc := NewClient(getFakeClient())

	_, err := tc.UpdateApi("default", "non-existent", "", "test", "default")
	assertions.Error(err)

	_, err = tc.UpdateApi("default", "sample", "", "test", "default")
	assertions.NoError(err)
}

func TestGetSvc(t *testing.T) {
	testClient := NewClient(getFakeClient())
	_, err := testClient.GetSvc("default", "kubernetes")
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
}

func TestListServices(t *testing.T) {
	testClient := NewClient(getFakeClient())
	_, err := testClient.ListServices("default")
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
}

func TestCreateStaticRoute(t *testing.T) {
	assertions := require.New(t)

	namespace := "default"
	name := "static-route-example-1-top-level-upstream"
	fleetNamespace := "default"
	fleetName := "default"
	specs := `
spec:
  upstream:
    service:
      name: static-route-example-1-top-level-upstream
      namespace: default
      port: 80
`
	testClient := NewClient(getFakeClient())
	staticRoute, err := testClient.CreateStaticRoute(namespace, name, fleetNamespace, fleetName, specs)

	assertions.NoError(err)
	assertions.NotNil(staticRoute)

	assertions.Equal(fleetNamespace+"."+fleetName, staticRoute.Spec.Fleet.String())
}

func TestGetStaticRoutes(t *testing.T) {
	assertions := require.New(t)

	namespace := "default"
	testClient := NewClient(getFakeClient())
	staticRoutes, err := testClient.GetStaticRoutes(namespace)

	assertions.NoError(err)
	assertions.NotNil(staticRoutes)
	assertions.Len(staticRoutes.Items, 1)
}

func TestDeleteStaticRoute(t *testing.T) {
	assertions := require.New(t)

	name := "static-route-1"
	namespace := "default"
	testClient := NewClient(getFakeClient())
	err := testClient.DeleteStaticRoute(kuskv1.StaticRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	})

	assertions.NoError(err)
}

func getFakeClient() client.Client {
	scheme := runtime.NewScheme()
	_ = kuskv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	initObjects := []client.Object{
		&kuskv1.API{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sample",
				Namespace: "default",
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kubernetes",
				Namespace: "default",
			},
		},
		&kuskv1.EnvoyFleet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "default",
				Namespace: "default",
			},
		},
		&kuskv1.StaticRoute{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "static-route-1",
				Namespace: "default",
			},
			Spec: kuskv1.StaticRouteSpec{
				Fleet: &kuskv1.EnvoyFleetID{
					Name:      "/static-route-1",
					Namespace: "default",
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(initObjects...).
		Build()
	return fakeClient
}
