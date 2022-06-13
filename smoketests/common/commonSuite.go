package common

import (
	"context"
	"time"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
	"github.com/stretchr/testify/suite"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	KuskNamespace = "kusk-system"
	KuskManager   = "kusk-gateway-manager"
)

type KuskTestSuite struct {
	Cli client.Client
	suite.Suite
}

func (s *KuskTestSuite) SetupSuite() {
	s.setupAndWaitForReady()
}

func (s *KuskTestSuite) setupAndWaitForReady() {

	config, err := GetKubeconfig()
	s.NoError(err)

	scheme := runtime.NewScheme()
	kuskv1.AddToScheme(scheme)
	corev1.AddToScheme(scheme)
	apps.AddToScheme(scheme)

	s.Cli, err = client.New(config, client.Options{Scheme: scheme})
	s.NoError(err)

	deploy := apps.Deployment{}
	counter := 0
	for {
		if counter == 100 {
			break
		}
		s.NoError(s.Cli.Get(context.Background(), client.ObjectKey{Namespace: KuskNamespace, Name: KuskManager}, &deploy))

		if deploy.Status.AvailableReplicas > 0 && deploy.Status.ReadyReplicas > 0 {
			break
		} else {
			time.Sleep(2 * time.Second)
			counter++
		}
	}
}
