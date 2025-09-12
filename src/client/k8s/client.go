package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	k8sClient        *kubernetes.Clientset
	k8sDynamicClient *dynamic.DynamicClient
	k8sController    *Controller
	controllerOnce   sync.Once
)

func Init(ctx context.Context, callback Callback) {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(fmt.Errorf("failed to read Kubernetes config: %v", err))
	}

	getK8sClient(restConfig)

	go GetK8sController().Run(ctx, callback)
}

func GetK8sController() *Controller {
	controllerOnce.Do(func() {
		k8sController = NewController()
	})
	return k8sController
}

func getK8sClient(restConfig *rest.Config) {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(fmt.Errorf("failed to create Kubernetes clientset: %v", err))
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		panic(fmt.Errorf("failed to create Kubernetes dynamic clientset: %v", err))
	}

	k8sClient = clientset
	k8sDynamicClient = dynamicClient
}
