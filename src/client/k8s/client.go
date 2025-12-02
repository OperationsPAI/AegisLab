package k8s

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	k8sRestConfig    *rest.Config
	k8sClient        *kubernetes.Clientset
	k8sDynamicClient *dynamic.DynamicClient
	k8sController    *Controller

	k8sRestConfigOnce    sync.Once
	k8sClientOnce        sync.Once
	k8sDynamicClientOnce sync.Once
	controllerOnce       sync.Once
)

func Init(ctx context.Context, callback Callback) {
	GetK8sClient()
	GetK8sDynamicClient()
	go GetK8sController().Run(ctx, callback)
}

func GetK8sClient() *kubernetes.Clientset {
	k8sClientOnce.Do(func() {
		restConfig := GetK8sRestConfig()
		clientset, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			logrus.Fatalf("failed to create Kubernetes clientset: %v", err)
		}

		k8sClient = clientset
	})
	return k8sClient
}

func GetK8sDynamicClient() *dynamic.DynamicClient {
	k8sDynamicClientOnce.Do(func() {
		restConfig := GetK8sRestConfig()
		dynamicClient, err := dynamic.NewForConfig(restConfig)
		if err != nil {
			logrus.Fatalf("failed to create Kubernetes dynamic client: %v", err)
		}

		k8sDynamicClient = dynamicClient
	})
	return k8sDynamicClient
}

func GetK8sRestConfig() *rest.Config {
	k8sRestConfigOnce.Do(func() {
		var restConfig *rest.Config
		var err error
		var currentContext string

		restConfig, err = rest.InClusterConfig()
		if err == nil {
			logrus.Info("Successfully loaded In-Cluster Kubernetes configuration.")
			currentContext = "In-Cluster"
			k8sRestConfig = restConfig
			logrus.Infof("Using Kubernetes Context: %s", currentContext) // 输出 Context
			return
		}

		kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		if _, err := os.Stat(kubeconfigPath); err == nil {
			configLoad := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
				&clientcmd.ConfigOverrides{},
			)

			restConfig, err = configLoad.ClientConfig()
			if err != nil {
				logrus.Warnf("Failed to read external Kubernetes config from %s: %v", kubeconfigPath, err)
			} else {
				rawConfig, _ := configLoad.RawConfig()
				currentContext = rawConfig.CurrentContext

				logrus.Infof("Successfully loaded external Kubernetes configuration from %s.", kubeconfigPath)
				k8sRestConfig = restConfig

				if currentContext != "" {
					logrus.Infof("Using Kubernetes Context: %s", currentContext)
				} else {
					logrus.Warn("Current context name is empty in Kubeconfig.")
				}
				return
			}
		}

		if k8sRestConfig == nil {
			logrus.Fatalf("Failed to establish Kubernetes REST config: Neither In-Cluster nor external Kubeconfig available.")
		}
	})
	return k8sRestConfig
}

func GetK8sController() *Controller {
	controllerOnce.Do(func() {
		k8sController = NewController()
	})
	return k8sController
}
