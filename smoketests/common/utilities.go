package common

import (
	"os"
	"path"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubeconfig() (*rest.Config, error) {
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
