package k8s

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func CreateJob(ctx context.Context, jobConfig JobConfig) error {
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
		return fmt.Errorf("Failed to create job: %v", err)
	}

	return nil
}

func DeleteJob(ctx context.Context, namespace, name string) error {
	deletePolicy := metav1.DeletePropagationBackground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	logEntry := logrus.WithField("namespace", namespace).WithField("name", name)

	// 1. 先检查 Job 是否存在及状态
	job, err := k8sClient.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("Failed to get Job: %v", err)
	}

	// 2. 检查是否已在删除中
	if !job.GetDeletionTimestamp().IsZero() {
		logEntry.Info("Job is already being deleted")
		return nil
	}

	// 3. 执行删除（幂等操作）
	err = k8sClient.BatchV1().Jobs(namespace).Delete(ctx, name, deleteOptions)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("Failed to delete Job: %v", err)
	}

	return nil
}

func GetJob(ctx context.Context, namespace, jobName string) (*batchv1.Job, error) {
	job, err := k8sClient.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Failed to get job: %v", err)
	}
	return job, nil
}

func GetJobPodLogs(ctx context.Context, namespace, jobName string) (map[string]string, error) {
	podList, err := k8sClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to list pods: %v", err)
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
