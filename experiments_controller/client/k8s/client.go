package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var k8sClient kubernetes.Interface

func Init(ctx context.Context, callback Callback) {
	getK8sClient()
	controller := NewController()
	go controller.Run(ctx, callback)
}

func getK8sClient() {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(fmt.Errorf("failed to read Kubernetes config: %v", err))
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Errorf("failed to create Kubernetes clientset: %v", err))
	}

	k8sClient = clientset
}
