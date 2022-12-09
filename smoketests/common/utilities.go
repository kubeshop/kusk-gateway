package common

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetKubeconfig() (*rest.Config, error) {
	var err error
	var config *rest.Config
	k8sConfigExists := false
	homeDir, _ := os.UserHomeDir()
	cubeConfigPath := path.Join(homeDir, ".kube/config")

	if _, err := os.Stat(cubeConfigPath); err == nil {
		k8sConfigExists = true
	}

	if cfg, exists := os.LookupEnv("KUBECONFIG"); exists {
		config, err = clientcmd.BuildConfigFromFlags("", cfg)
	} else if k8sConfigExists {
		config, err = clientcmd.BuildConfigFromFlags("", cubeConfigPath)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
	}
	// default query per second is set to 5
	config.QPS = 40.0
	// default burst is set to 10
	config.Burst = 400.0

	return config, err
}

func ReadFile(path string) string {
	if !filepath.IsAbs(path) {
		path, _ = filepath.Abs(path)
	}
	dat, _ := os.ReadFile(path)

	return string(dat)
}

func WaitForServiceReady(ctx context.Context, k8sClient client.Client, namespace string, name string, timeout time.Duration) error {
	s := &corev1.Service{}
	time.Sleep(1 * time.Second)
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, s); err != nil {
		return err
	}

	if err := wait.PollImmediate(1*time.Second, timeout, IsPodReady(ctx, k8sClient, s)); err != nil {
		return err
	}
	return nil
}

func IsPodReady(ctx context.Context, c client.Client, service *corev1.Service) wait.ConditionFunc {
	return func() (bool, error) {
		if err := c.Get(ctx, client.ObjectKey{Namespace: service.Namespace, Name: service.Name}, service); err != nil {
			return false, err
		}

		if service.Status.LoadBalancer.Ingress != nil && len(service.Status.LoadBalancer.Ingress) > 0 {
			return true, nil
		}
		return false, nil
	}
}
