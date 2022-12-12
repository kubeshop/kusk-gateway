package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClientSet() (*kubernetes.Clientset, error) {
	k8sConfig, err := getKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s config: %w", err)
	}

	clientSet, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	return clientSet, nil
}

func GetServiceContainerLogStream(
	ctx context.Context,
	namespace, svcName, containerName string,
	tailLineCount int64,
	corev1Client typev1.CoreV1Interface,
) (io.ReadCloser, error) {
	servicePods, err := getPodsForSvc(ctx, svcName, namespace, corev1Client)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods for service %s/%s: %w", namespace, svcName, err)
	}

	if len(servicePods.Items) == 0 {
		return nil, fmt.Errorf("no pods found for service %s/%s", namespace, svcName)
	}

	pod := servicePods.Items[0]
	container, err := getContainerFromPod(&pod, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get container %s from pod %s: %w", containerName, pod.Name, err)
	}

	podLogOptions := corev1.PodLogOptions{
		Follow:    true,
		TailLines: &tailLineCount,
		Container: container.Name,
	}

	return corev1Client.
		Pods(pod.Namespace).
		GetLogs(pod.Name, &podLogOptions).
		Stream(ctx)
}

func getKubeConfig() (*rest.Config, error) {
	var (
		config          *rest.Config
		err             error
		k8sConfigExists bool
	)
	homeDir, _ := os.UserHomeDir()
	kubeConfigPath := path.Join(homeDir, ".kube/config")

	if _, err := os.Stat(kubeConfigPath); err == nil {
		k8sConfigExists = true
	}

	if cfg, exists := os.LookupEnv("KUBECONFIG"); exists {
		config, err = clientcmd.BuildConfigFromFlags("", cfg)
	} else if k8sConfigExists {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
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

func getPodsForSvc(ctx context.Context, svcName, namespace string, k8sClient typev1.CoreV1Interface) (*corev1.PodList, error) {
	svc, err := k8sClient.
		Services(namespace).
		Get(ctx, svcName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service %s/%s: %w", namespace, svcName, err)
	}
	set := labels.Set(svc.Spec.Selector)
	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	return k8sClient.Pods(namespace).List(ctx, listOptions)
}

func getContainerFromPod(pod *corev1.Pod, containerName string) (*corev1.Container, error) {
	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			return &container, nil
		}
	}
	return nil, fmt.Errorf("container %s not found in pod %s", containerName, pod.Name)
}
