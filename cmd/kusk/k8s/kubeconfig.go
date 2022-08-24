package k8s

import (
	"errors"
	"os"
	"path/filepath"
)

func GetKubeConfig() (string, error) {
	if kubeconfig, ok := os.LookupEnv("KUBECONFIG"); ok {
		if kubeconfig == "" {
			return "", errors.New("env var KUBECONFIG set but is empty")
		}

		return kubeconfig, nil
	}

	var homeDir string
	if h := os.Getenv("HOME"); h != "" {
		homeDir = h
	} else {
		homeDir = os.Getenv("USERPROFILE") // windows
	}

	return filepath.Join(homeDir, ".kube", "config"), nil
}
