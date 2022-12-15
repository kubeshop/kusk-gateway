package kusk

import (
	"context"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
)

var ErrNotFound = errors.New("error not found")

type Client interface {
	GetEnvoyFleets() (*kuskv1.EnvoyFleetList, error)
	GetEnvoyFleet(namespace, name string) (*kuskv1.EnvoyFleet, error)
	CreateFleet(kuskv1.EnvoyFleet) (*kuskv1.EnvoyFleet, error)
	DeleteFleet(kuskv1.EnvoyFleet) error

	GetApis(namespace string) (*kuskv1.APIList, error)
	GetApi(namespace, name string) (*kuskv1.API, error)
	GetApiByEnvoyFleet(namespace, fleetNamespace, fleetName string) (*kuskv1.APIList, error)
	CreateApi(namespace, name, openapispec, fleetName, fleetnamespace string) (*kuskv1.API, error)
	UpdateApi(namespace, name, openapispec, fleetName, fleetnamespace string) (*kuskv1.API, error)
	DeleteAPI(namespace, name string) error

	GetStaticRoute(namespace, name string) (*kuskv1.StaticRoute, error)
	GetStaticRoutes(namespace string) (*kuskv1.StaticRouteList, error)
	CreateStaticRoute(namespace, name, fleetName, fleetNamespace, specs string) (*kuskv1.StaticRoute, error)
	UpdateStaticRoute(namespace, name, fleetName, fleetNamespace, specs string) (*kuskv1.StaticRoute, error)
	DeleteStaticRoute(kuskv1.StaticRoute) error

	GetSvc(namespace, name string) (*corev1.Service, error)
	ListServices(namespace string) (*corev1.ServiceList, error)
	ListNamespaces() (*corev1.NamespaceList, error)
	GetSecret(name, namespace string) (*v1.Secret, error)

	K8sClient() client.Client
}

type kuskClient struct {
	client              client.Client
	kuskManagedSelector labels.Selector
}

func NewClient(c client.Client) Client {
	// for use when querying apis and static routes to filter out those that are managed by kusk
	r, _ := labels.NewRequirement("kusk-managed", selection.NotIn, []string{"true"})

	return &kuskClient{
		client:              c,
		kuskManagedSelector: labels.NewSelector().Add(*r),
	}
}

func (k *kuskClient) GetEnvoyFleets() (*kuskv1.EnvoyFleetList, error) {

	list := &kuskv1.EnvoyFleetList{}

	if err := k.client.List(context.TODO(), list, &client.ListOptions{}); err != nil {
		return nil, err
	}
	return list, nil
}

func (k *kuskClient) GetEnvoyFleet(namespace, name string) (*kuskv1.EnvoyFleet, error) {
	envoy := &kuskv1.EnvoyFleet{}

	if err := k.client.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, envoy); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrNotFound
		}

		return nil, err
	}
	return envoy, nil
}

func (k *kuskClient) CreateFleet(fleet kuskv1.EnvoyFleet) (*kuskv1.EnvoyFleet, error) {
	if err := k.client.Create(context.TODO(), &fleet, &client.CreateOptions{}); err != nil {
		return nil, err
	}

	return &fleet, nil
}

func (k *kuskClient) DeleteFleet(fleet kuskv1.EnvoyFleet) error {
	return k.client.Delete(context.TODO(), &fleet, &client.DeleteOptions{})
}

func (k *kuskClient) GetApis(namespace string) (*kuskv1.APIList, error) {
	list := &kuskv1.APIList{}
	if err := k.client.List(
		context.TODO(),
		list,
		&client.ListOptions{
			Namespace:     namespace,
			LabelSelector: k.kuskManagedSelector,
		},
	); err != nil {
		return nil, err
	}

	return list, nil
}

func (k *kuskClient) GetApi(namespace, name string) (*kuskv1.API, error) {
	api := &kuskv1.API{}

	if err := k.client.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: name}, api); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return api, nil
}

// GetApiByFleet gets all APIs associated with the EnvoyFleet
func (k *kuskClient) GetApiByEnvoyFleet(namespace, fleetNamespace, fleetName string) (*kuskv1.APIList, error) {
	list := kuskv1.APIList{}
	if err := k.client.List(
		context.TODO(),
		&list,
		&client.ListOptions{
			Namespace:     namespace,
			LabelSelector: k.kuskManagedSelector,
		},
	); err != nil {
		return nil, err
	}

	toReturn := []kuskv1.API{}
	for _, api := range list.Items {
		if api.Spec.Fleet.Name == fleetName && api.Spec.Fleet.Namespace == fleetNamespace {
			toReturn = append(toReturn, api)
		}
	}

	return &kuskv1.APIList{Items: toReturn}, nil
}

func (k *kuskClient) CreateApi(namespace, name, openapispec string, fleetName string, fleetnamespace string) (*kuskv1.API, error) {
	api := &kuskv1.API{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: kuskv1.APISpec{
			Spec: openapispec,
			Fleet: &kuskv1.EnvoyFleetID{
				Name:      fleetName,
				Namespace: fleetnamespace,
			},
		},
	}
	if err := k.client.Create(context.TODO(), api, &client.CreateOptions{}); err != nil {
		return nil, err
	}
	return api, nil
}

func (k *kuskClient) UpdateApi(namespace, name, openapispec string, fleetName string, fleetnamespace string) (*kuskv1.API, error) {
	api := &kuskv1.API{}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of API before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		if err := k.client.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: name}, api); err != nil {
			return err
		}

		api.Spec = kuskv1.APISpec{
			Spec: openapispec,
			Fleet: &kuskv1.EnvoyFleetID{
				Name:      fleetName,
				Namespace: fleetnamespace,
			},
		}

		if err := k.client.Update(context.TODO(), api, &client.UpdateOptions{}); err != nil {
			return err
		}

		return nil
	})

	return api, retryErr
}

func (k *kuskClient) DeleteAPI(namespace, name string) error {
	api := &kuskv1.API{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return k.client.Delete(context.TODO(), api, &client.DeleteOptions{})
}

func (k *kuskClient) GetSvc(namespace, name string) (*corev1.Service, error) {
	svc := &corev1.Service{}
	if err := k.client.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: name}, svc); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return svc, nil
}

func (k *kuskClient) ListServices(namespace string) (*corev1.ServiceList, error) {
	list := &corev1.ServiceList{}

	if err := k.client.List(context.TODO(), list, &client.ListOptions{Namespace: namespace}); err != nil {
		return nil, err
	}

	return list, nil

}

func (k *kuskClient) GetStaticRoute(namespace, name string) (*kuskv1.StaticRoute, error) {
	staticRoute := &kuskv1.StaticRoute{}
	if err := k.client.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: name}, staticRoute); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return staticRoute, nil
}

func (k *kuskClient) GetStaticRoutes(namespace string) (*kuskv1.StaticRouteList, error) {
	list := &kuskv1.StaticRouteList{}
	if err := k.client.List(
		context.TODO(),
		list,
		&client.ListOptions{
			Namespace:     namespace,
			LabelSelector: k.kuskManagedSelector,
		},
	); err != nil {
		return nil, err
	}

	return list, nil
}

func (k *kuskClient) CreateStaticRoute(namespace, name, fleetName, fleetNamespace, specs string) (*kuskv1.StaticRoute, error) {
	staticRoute := &kuskv1.StaticRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: kuskv1.StaticRouteSpec{
			Fleet: &kuskv1.EnvoyFleetID{
				Name:      fleetName,
				Namespace: fleetNamespace,
			},
		},
	}

	tmp := &kuskv1.StaticRoute{}

	err := yaml.Unmarshal([]byte(specs), tmp)
	if err != nil {
		fmt.Println(fmt.Errorf("CreateStaticRoute - yaml.Unmarshal failed: specs=%v, %w", specs, err))
	}

	staticRoute.Spec.Hosts = tmp.Spec.Hosts
	staticRoute.Spec.Upstream = tmp.Spec.Upstream

	if err := k.client.Create(context.TODO(), staticRoute, &client.CreateOptions{}); err != nil {
		return nil, err
	}

	return staticRoute, nil
}

func (k *kuskClient) UpdateStaticRoute(namespace, name, fleetName, fleetNamespace, specs string) (*kuskv1.StaticRoute, error) {
	staticRoute := &kuskv1.StaticRoute{}

	// marshal the paths and hosts separately for the static route
	// to use later
	tmp := &kuskv1.StaticRoute{}
	err := yaml.Unmarshal([]byte(specs), tmp)
	if err != nil {
		fmt.Println(fmt.Errorf("UpdateStaticRoute - yaml.Unmarshal failed: specs=%v, %w", specs, err))
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := k.client.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: name}, staticRoute); err != nil {
			return err
		}

		staticRoute.Spec = kuskv1.StaticRouteSpec{
			Fleet: &kuskv1.EnvoyFleetID{
				Name:      fleetName,
				Namespace: fleetNamespace,
			},
			Hosts:    tmp.Spec.Hosts,
			Upstream: tmp.Spec.Upstream,
		}

		return k.client.Update(context.TODO(), staticRoute, &client.UpdateOptions{})
	})

	return staticRoute, retryErr
}

func (k *kuskClient) DeleteStaticRoute(sroute kuskv1.StaticRoute) error {
	return k.client.Delete(context.TODO(), &sroute, &client.DeleteOptions{})
}

func (k *kuskClient) ListNamespaces() (*corev1.NamespaceList, error) {
	list := &corev1.NamespaceList{}
	if err := k.client.List(context.TODO(), list, &client.ListOptions{}); err != nil {
		return nil, err
	}
	return list, nil
}

func (k *kuskClient) GetSecret(name, namespace string) (*v1.Secret, error) {
	sec := &v1.Secret{}
	if err := k.client.Get(context.TODO(), client.ObjectKey{Name: name, Namespace: namespace}, sec); err != nil {
		return nil, err
	}

	return sec, nil
}

func (k *kuskClient) K8sClient() client.Client {
	return k.client
}
