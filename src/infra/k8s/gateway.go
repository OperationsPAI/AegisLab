package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"aegis/consts"

	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Gateway struct {
	controller *Controller
}

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

func NewGateway(controller *Controller) *Gateway {
	if controller == nil {
		controller = getK8sController()
	}
	return &Gateway{controller: controller}
}

func (g *Gateway) GetVolumeMountConfigMap() (map[consts.VolumeMountName]VolumeMountConfig, error) {
	return getVolumeMountConfigMap()
}

func (g *Gateway) CreateJob(ctx context.Context, jobConfig *JobConfig) error {
	return createJob(ctx, jobConfig)
}

func (g *Gateway) GetJob(ctx context.Context, namespace, jobName string) (*batchv1.Job, error) {
	return getJob(ctx, namespace, jobName)
}

func (g *Gateway) WaitForJobCompletion(ctx context.Context, namespace, jobName string) error {
	return waitForJobCompletion(ctx, namespace, jobName)
}

func (g *Gateway) GetJobPodLogs(ctx context.Context, namespace, jobName string) (map[string][]string, error) {
	return getJobPodLogs(ctx, namespace, jobName)
}

func (g *Gateway) DeleteJob(ctx context.Context, namespace, jobName string) error {
	return deleteJob(ctx, namespace, jobName)
}

func (g *Gateway) CheckHealth(ctx context.Context) error {
	if getK8sRestConfig() == nil {
		return fmt.Errorf("kubernetes config not available")
	}
	client := getK8sClient()
	if client == nil {
		return fmt.Errorf("kubernetes client not available")
	}
	if getK8sDynamicClient() == nil {
		return fmt.Errorf("kubernetes dynamic client not available")
	}

	if _, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1}); err != nil {
		return fmt.Errorf("kubernetes API request failed: %w", err)
	}
	return nil
}

func getK8sClient() *kubernetes.Clientset {
	k8sClientOnce.Do(func() {
		restConfig := getK8sRestConfig()
		clientset, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			logrus.Fatalf("failed to create Kubernetes clientset: %v", err)
		}

		k8sClient = clientset
	})
	return k8sClient
}

func getK8sDynamicClient() *dynamic.DynamicClient {
	k8sDynamicClientOnce.Do(func() {
		restConfig := getK8sRestConfig()
		dynamicClient, err := dynamic.NewForConfig(restConfig)
		if err != nil {
			logrus.Fatalf("failed to create Kubernetes dynamic client: %v", err)
		}

		k8sDynamicClient = dynamicClient
	})
	return k8sDynamicClient
}

func getK8sRestConfig() *rest.Config {
	k8sRestConfigOnce.Do(func() {
		restConfig, err := rest.InClusterConfig()
		if err == nil {
			logrus.Info("Successfully loaded In-Cluster Kubernetes configuration.")
			k8sRestConfig = restConfig
			logrus.Infof("Using Kubernetes Context: %s", "In-Cluster")
			return
		}

		logrus.Warn("In-cluster config not found, trying kubeconfig file")
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			logrus.Fatalf("Failed to load Kubernetes config: %v", err)
		}
		if config == nil {
			logrus.Fatalf("Failed to establish Kubernetes REST config: Neither In-Cluster nor external Kubeconfig available.")
		}

		k8sRestConfig = config
	})
	return k8sRestConfig
}

func getK8sController() *Controller {
	controllerOnce.Do(func() {
		k8sController = newController()
	})
	return k8sController
}
