package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var k8sClient kubernetes.Interface

func Init(ctx context.Context, callback Callback) {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(fmt.Errorf("failed to read Kubernetes config: %v", err))
	}

	getK8sClient(restConfig)

	controller := NewController(restConfig)
	if controller != nil {
		go controller.Run(ctx, callback)
	}
}

func getK8sClient(restConfig *rest.Config) {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(fmt.Errorf("failed to create Kubernetes clientset: %v", err))
	}

	k8sClient = clientset
}
