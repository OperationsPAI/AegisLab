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
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

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

type JobEnv struct {
	Namespace string
	StartTime time.Time
	EndTime   time.Time
}

type Callback interface {
	AddFunc(labels map[string]string)
	UpdateFunc(labels map[string]string)
}

var k8sClient *kubernetes.Clientset

func InitK8s(ctx context.Context, callback Callback) {
	getK8sClient()
	getJobInformer(ctx, callback)
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

func getJobInformer(ctx context.Context, callback Callback) {
	jobInformer := cache.NewSharedInformer(
		cache.NewListWatchFromClient(
			k8sClient.BatchV1().RESTClient(),
			"jobs",
			config.GetString("k8s.namespace"),
			fields.Everything(),
		),
		&batchv1.Job{},
		time.Second*10,
	)

	jobInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			job := obj.(*batchv1.Job)
			logrus.Infof("Job %s created successfully in namespace %s", job.Name, job.Namespace)
			callback.AddFunc(job.Labels)
		},
		UpdateFunc: func(oldObj, newObj any) {
			oldJob := oldObj.(*batchv1.Job)
			newJob := newObj.(*batchv1.Job)

			if callback != nil && oldJob.Name == newJob.Name {
				if oldJob.Status.Succeeded == 0 && newJob.Status.Succeeded > 0 {
					callback.UpdateFunc(newJob.Labels)
					if err := DeleteK8sJob(context.Background(), config.GetString("k8s.namespace"), newJob.Name); err != nil {
						logrus.Error(err)
					}
				}
			}
		},
		DeleteFunc: func(obj any) {
			job := obj.(*batchv1.Job)
			logrus.Infof("Job %s deleted successfully in namespace %s", job.Name, job.Namespace)
		},
	})

	stopCh := make(chan struct{})
	defer close(stopCh)
	go jobInformer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, jobInformer.HasSynced) {
		panic("Timed out waiting for caches to sync")
	}

	<-ctx.Done()
	logrus.Info("Stopping informer...")
}

func CreateK8sJob(ctx context.Context, jobConfig JobConfig) error {
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

	_, err := k8sClient.BatchV1().Jobs(jobConfig.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create job: %v", err)
	}

	return nil
}

func GetK8sJobPodLogs(ctx context.Context, namespace, jobName string) (map[string]string, error) {
	podList, err := k8sClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	logsMap := make(map[string]string)
	for _, pod := range podList.Items {
		req := k8sClient.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
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
	for {
		job, err := k8sClient.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
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
	job, err := k8sClient.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %v", err)
	}
	return job, nil
}

func DeleteK8sJob(ctx context.Context, namespace, jobName string) error {
	deletePolicy := metav1.DeletePropagationBackground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	if err := k8sClient.BatchV1().Jobs(namespace).Delete(ctx, jobName, deleteOptions); err != nil {
		return fmt.Errorf("Failed to delete job %s:%v", jobName, err)
	}

	return nil
}

func GetJobPodLogs(ctx context.Context, namespace, jobName string) (map[string]string, error) {
	// 获取 Job 关联的 Pods
	podList, err := k8sClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods for job %s: %v", jobName, err)
	}

	// 存储每个 Pod 的日志
	logsMap := make(map[string]string)

	for _, pod := range podList.Items {
		req := k8sClient.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})

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
