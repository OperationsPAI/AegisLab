package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	k8sClientInstance client.Client
	once              sync.Once
)

func GetK8sConfig() *rest.Config {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config
}

func NewK8sClient() client.Client {
	once.Do(func() {
		cfg := GetK8sConfig()
		scheme := runtime.NewScheme()

		err := corev1.AddToScheme(scheme)
		if err != nil {
			logrus.Fatalf("Failed to add CoreV1 scheme: %v", err)
		}
		if err := batchv1.AddToScheme(scheme); err != nil {
			logrus.Fatalf("Failed to add batchV1 scheme: %v", err)
		}
		// Create Kubernetes client
		k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
		if err != nil {
			logrus.Fatalf("Failed to create Kubernetes client: %v", err)
		}
		k8sClientInstance = k8sClient
	})
	return k8sClientInstance
}

func CreateK8sJob(k8sClient client.Client, namespace, jobName, image string, command []string, restartPolicy corev1.RestartPolicy, backoffLimit int32, parallelism, completions int32, envVars []corev1.EnvVar, volumeMounts []corev1.VolumeMount, volumes []corev1.Volume) error {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Parallelism: &parallelism, // 并行执行的 Pod 数量
			Completions: &completions, // 总的任务数
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: restartPolicy,
					Containers: []corev1.Container{
						{
							Name:         jobName,
							Image:        image,
							Command:      command,
							Env:          envVars,
							VolumeMounts: volumeMounts,
						},
					},
					Volumes: volumes,
				},
			},
			BackoffLimit: &backoffLimit, // 最大失败重试次数
		},
	}

	err := k8sClient.Create(context.TODO(), job)
	if err != nil {
		return fmt.Errorf("failed to create job: %v", err)
	}

	logrus.Infof("Job %q created successfully.", jobName)
	return nil
}

func WaitForJobCompletion(k8sClient client.Client, namespace, jobName string) error {
	job := &batchv1.Job{}
	for {
		err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: jobName, Namespace: namespace}, job)
		if err != nil {
			return fmt.Errorf("failed to get job: %v", err)
		}

		if job.Status.Succeeded > 0 {
			logrus.Info("Job completed successfully!")
			break
		}

		logrus.Info("Waiting for job to complete...")
		time.Sleep(2 * time.Second)
	}
	return nil
}

func GetK8sJob(k8sClient client.Client, namespace, jobName string) (*batchv1.Job, error) {
	job := &batchv1.Job{}
	err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: jobName, Namespace: namespace}, job)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %v", err)
	}
	return job, nil
}

func DeleteK8sJob(k8sClient client.Client, namespace, jobName string) error {
	// Delete the job
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(context.TODO(), job)
	if err != nil {
		return fmt.Errorf("failed to delete job: %v", err)
	}

	// Delete the pods associated with the job
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels{"job-name": jobName},
	}
	err = k8sClient.List(context.TODO(), podList, listOpts...)
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}

	for _, pod := range podList.Items {
		err = k8sClient.Delete(context.TODO(), &pod)
		if err != nil {
			return fmt.Errorf("failed to delete pod %s: %v", pod.Name, err)
		}
		logrus.Infof("Pod %q deleted successfully.", pod.Name)
	}

	logrus.Infof("Job %q and its pods deleted successfully.", jobName)
	return nil
}
