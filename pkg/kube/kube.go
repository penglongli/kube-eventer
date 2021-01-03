package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	globalKubeClient *KubeClient
)

type KubeClient struct {
	restConfig *rest.Config
	clientSet *kubernetes.Clientset
}

func InitKubeClient() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	globalKubeClient = &KubeClient{
		restConfig: config,
		clientSet: clientSet,
	}
	return nil
}

func GetRestConfig() *rest.Config {
	return globalKubeClient.restConfig
}

func GetClientSet() *kubernetes.Clientset {
	return globalKubeClient.clientSet
}
