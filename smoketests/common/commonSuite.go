package common

import (
	"time"

	"github.com/stretchr/testify/suite"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kuskv1 "github.com/kubeshop/kusk-gateway/api/v1alpha1"
)

type KuskTestSuite struct {
	Cli client.Client
	suite.Suite
}

func (s *KuskTestSuite) SetupSuite() {
	setupAndWaitForReady(s)
}

func setupAndWaitForReady(s *KuskTestSuite) {
	config, err := GetKubeconfig()
	s.NoError(err)

	scheme := runtime.NewScheme()
	s.NoError(kuskv1.AddToScheme(scheme))
	s.NoError(corev1.AddToScheme(scheme))
	s.NoError(apps.AddToScheme(scheme))

	s.Cli, err = client.New(config, client.Options{Scheme: scheme})
	s.NoError(err)

	// weird way to wait it out probably needs to be done dynamically
	waitBeforeStartingTest := time.Second * 1
	s.T().Logf("Sleeping for %s", waitBeforeStartingTest)
	time.Sleep(waitBeforeStartingTest)
}
