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
	s.NoError(kuskv1.AddToScheme(scheme))
	s.NoError(corev1.AddToScheme(scheme))
	s.NoError(apps.AddToScheme(scheme))

	s.Cli, err = client.New(config, client.Options{Scheme: scheme})
	s.NoError(err)

	time.Sleep(3 * time.Second)
}
