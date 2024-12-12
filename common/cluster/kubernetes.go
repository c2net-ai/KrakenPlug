package cluster

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

func buildConfigWithDefaultPath(kubeconfig string) (*rest.Config, error) {
	if kubeconfig == "" {
		homeDir, _ := os.UserHomeDir()
		kubeconfig = filepath.Join(homeDir, ".kube", "config")
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func buildConfigFromFlagsOrCluster(configPath string) (*rest.Config, error) {
	cfg, err1 := rest.InClusterConfig()
	if err1 == nil {
		return cfg, nil
	}
	cfg, err2 := buildConfigWithDefaultPath(configPath)
	if err2 == nil {
		return cfg, nil
	}
	return nil, fmt.Errorf("load kubernetes config failed %v %v", err1, err2)
}

func NewClient() (*kubernetes.Clientset, error) {
	restConfig, err := buildConfigFromFlagsOrCluster("")
	if err != nil {
		return nil, fmt.Errorf("load kubernetes config: %v", err)
	}
	return kubernetes.NewForConfigOrDie(restConfig), nil
}
