package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var k8sClient *kubernetes.Clientset

func GetK8sConfig() *rest.Config {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config
}

func GetK8sClient() *kubernetes.Clientset {
	if k8sClient == nil {
		config := GetK8sConfig()
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(fmt.Errorf("failed to create Kubernetes clientset: %v", err))
		}
		k8sClient = clientset
	}
	return k8sClient
}

type JobConfig struct {
	Namespace     string
	JobName       string
	Image         string
	Command       []string
	RestartPolicy corev1.RestartPolicy
	BackoffLimit  int32
	Parallelism   int32
	Completions   int32
	EnvVars       []corev1.EnvVar
	Labels        map[string]string // 用于自定义标签
}

func CreateK8sJob(ctx context.Context, jobConfig JobConfig) error {
	clientset := GetK8sClient()

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "nfs-volume",
			MountPath: "/data",
		},
	}
	pvc := config.GetString("nfs.pvc_name")
	if config.GetString("nfs.pvc_name") == "" {
		pvc = "nfs-shared-pvc"
	}
	volumes := []corev1.Volume{
		{
			Name: "nfs-volume",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc,
				},
			},
		},
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobConfig.JobName,
			Namespace: jobConfig.Namespace,
			Labels:    jobConfig.Labels,
		},
		Spec: batchv1.JobSpec{
			Parallelism: &jobConfig.Parallelism,
			Completions: &jobConfig.Completions,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: jobConfig.Labels, // 给 Pod 应用相同的标签
				},
				Spec: corev1.PodSpec{
					RestartPolicy: jobConfig.RestartPolicy,
					Containers: []corev1.Container{
						{
							Name:         jobConfig.JobName,
							Image:        jobConfig.Image,
							Command:      jobConfig.Command,
							Env:          jobConfig.EnvVars,
							VolumeMounts: volumeMounts,
						},
					},
					Volumes: volumes,
				},
			},
			BackoffLimit: &jobConfig.BackoffLimit,
		},
	}

	_, err := clientset.BatchV1().Jobs(jobConfig.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create job: %v", err)
	}

	logrus.Infof("Job %q created successfully in namespace %q.", job.Name, job.Namespace)
	return nil
}

func GetK8sJobPodLogs(ctx context.Context, namespace, jobName string) (map[string]string, error) {
	clientset := GetK8sClient()

	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	logsMap := make(map[string]string)
	for _, pod := range podList.Items {
		req := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
		logStream, err := req.Stream(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get logs for pod %s: %v", pod.Name, err)
		}
		defer logStream.Close()

		logData, err := io.ReadAll(logStream)
		if err != nil {
			return nil, fmt.Errorf("failed to read logs for pod %s: %v", pod.Name, err)
		}
		logsMap[pod.Name] = string(logData)
	}

	return logsMap, nil
}

func WaitForJobCompletion(ctx context.Context, namespace, jobName string) error {
	clientset := GetK8sClient()

	for {
		job, err := clientset.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
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

func GetK8sJob(ctx context.Context, namespace, jobName string) (*batchv1.Job, error) {
	clientset := GetK8sClient()

	job, err := clientset.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %v", err)
	}
	return job, nil
}

func DeleteK8sJob(ctx context.Context, namespace, jobName string) error {
	clientset := GetK8sClient()

	// Delete the job
	err := clientset.BatchV1().Jobs(namespace).Delete(ctx, jobName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete job: %v", err)
	}

	// Delete the pods associated with the job
	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}

	for _, pod := range podList.Items {
		err := clientset.CoreV1().Pods(namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete pod %s: %v", pod.Name, err)
		}
		logrus.Infof("Pod %q deleted successfully.", pod.Name)
	}

	logrus.Infof("Job %q and its pods deleted successfully.", jobName)
	return nil
}

func GetJobPodLogs(ctx context.Context, namespace, jobName string) (map[string]string, error) {
	clientset := GetK8sClient()

	// 获取 Job 关联的 Pods
	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods for job %s: %v", jobName, err)
	}

	// 存储每个 Pod 的日志
	logsMap := make(map[string]string)

	for _, pod := range podList.Items {
		req := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})

		// 获取日志流
		logStream, err := req.Stream(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get logs for pod %s: %v", pod.Name, err)
		}
		defer logStream.Close()

		// 读取日志数据
		logData, err := io.ReadAll(logStream)
		if err != nil {
			return nil, fmt.Errorf("failed to read logs for pod %s: %v", pod.Name, err)
		}
		logsMap[pod.Name] = string(logData)
	}

	return logsMap, nil
}
