package k8s

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"aegis/config"
	"aegis/tracing"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JobConfig struct {
	Namespace      string
	JobName        string
	Image          string
	Command        []string
	RestartPolicy  corev1.RestartPolicy
	BackoffLimit   int32
	Parallelism    int32
	Completions    int32
	Annotations    map[string]string
	Labels         map[string]string
	EnvVars        []corev1.EnvVar
	InitContainers []corev1.Container
}

type VolumeMountConfig struct {
	Name      string `json:"name"`
	MountPath string `json:"mount_path"`

	// when source_type is "hostPath"
	HostPath string `json:"host_path,omitempty"`
	Type     string `json:"type,omitempty"`

	// when source_type is "secret"
	SubPath    string `json:"sub_path,omitempty"`
	SourceType string `json:"source_type,omitempty"`
	SecretName string `json:"secret_name,omitempty"`

	// when source_type is "pvc"
	ClaimName string `json:"claim_name,omitempty"`
}

func (v *VolumeMountConfig) GetVolumeMount() corev1.VolumeMount {
	volumeMount := corev1.VolumeMount{
		Name:      v.Name,
		MountPath: v.MountPath,
	}

	if v.SubPath != "" {
		volumeMount.SubPath = v.SubPath
	}

	return volumeMount
}

func (v *VolumeMountConfig) GEtVolume() corev1.Volume {
	volume := corev1.Volume{
		Name: v.Name,
	}

	switch v.SourceType {
	case "secret":
		volume.VolumeSource = corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: v.SecretName,
				Items: []corev1.KeyToPath{
					{
						Key:  v.SubPath,
						Path: v.SubPath,
					},
				},
			},
		}
	case "pvc":
		volume.VolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: v.ClaimName,
			},
		}
	default: // hostPath
		volume.VolumeSource = corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: v.HostPath,
				Type: func() *corev1.HostPathType {
					hostPathType := corev1.HostPathType(v.Type)
					return &hostPathType
				}(),
			},
		}
	}

	return volume
}

func CreateJob(ctx context.Context, jobConfig *JobConfig) error {
	return tracing.WithSpan(ctx, func(ctx context.Context) error {
		span := trace.SpanFromContext(ctx)

		jobConfig.Namespace = config.GetString("k8s.namespace")
		jobConfig.BackoffLimit = int32(0)
		jobConfig.Parallelism = int32(1)
		jobConfig.Completions = int32(1)
		jobConfig.RestartPolicy = corev1.RestartPolicyNever

		volumeMountConfigs := make([]VolumeMountConfig, 0)
		for _, cfgData := range config.GetMap("k8s.job.volume_mount") {
			cfg, err := utils.ConvertToType[VolumeMountConfig](cfgData)
			if err != nil {
				return fmt.Errorf("invalid volume mount config %v: %v", cfgData, err)
			}

			volumeMountConfigs = append(volumeMountConfigs, cfg)
		}

		volumeMounts := []corev1.VolumeMount{}
		volumes := []corev1.Volume{}
		for _, cfg := range volumeMountConfigs {
			volumeMounts = append(volumeMounts, cfg.GetVolumeMount())
			volumes = append(volumes, cfg.GEtVolume())
		}

		jobConfig.Labels["job-name"] = jobConfig.JobName
		if jobConfig.InitContainers != nil {
			for i := range jobConfig.InitContainers {
				if jobConfig.InitContainers[i].VolumeMounts == nil {
					jobConfig.InitContainers[i].VolumeMounts = append(jobConfig.InitContainers[i].VolumeMounts, volumeMounts...)
				}
			}
		}

		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: jobConfig.Annotations,
				Labels:      jobConfig.Labels,
				Name:        jobConfig.JobName,
				Namespace:   jobConfig.Namespace,
			},
			Spec: batchv1.JobSpec{
				Parallelism: &jobConfig.Parallelism,
				Completions: &jobConfig.Completions,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: jobConfig.Labels,
					},
					Spec: corev1.PodSpec{
						RestartPolicy:  jobConfig.RestartPolicy,
						InitContainers: jobConfig.InitContainers,
						Containers: []corev1.Container{
							{
								Name:            jobConfig.JobName,
								Image:           jobConfig.Image,
								Command:         jobConfig.Command,
								Env:             jobConfig.EnvVars,
								VolumeMounts:    volumeMounts,
								ImagePullPolicy: corev1.PullAlways,
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
			span.RecordError(err)
			span.AddEvent("failed to create job")
			return fmt.Errorf("failed to create job: %v", err)
		}

		return nil
	})
}

func deleteJob(ctx context.Context, namespace, name string) error {
	deletePolicy := metav1.DeletePropagationBackground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	logEntry := logrus.WithField("namespace", namespace).WithField("name", name)

	// 1. First check if Job exists and its status
	job, err := k8sClient.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to get Job: %v", err)
	}

	// 2. Check if already being deleted
	if !job.GetDeletionTimestamp().IsZero() {
		logEntry.Info("job is already being deleted")
		return nil
	}

	// 3. Execute deletion (idempotent operation)
	err = k8sClient.BatchV1().Jobs(namespace).Delete(ctx, name, deleteOptions)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("failed to delete Job: %v", err)
	}

	return nil
}

func GetJob(ctx context.Context, namespace, jobName string) (*batchv1.Job, error) {
	job, err := k8sClient.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %v", err)
	}
	return job, nil
}

func GetJobPodLogs(ctx context.Context, namespace, jobName string) (map[string][]string, error) {
	podList, err := k8sClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	logsMap := make(map[string][]string)
	for _, pod := range podList.Items {
		if !isPodReadyForLogs(pod) {
			logrus.WithFields(logrus.Fields{
				"pod":   pod.Name,
				"phase": pod.Status.Phase,
			}).Info("Skipping pod logs - pod not ready")
			continue
		}

		req := k8sClient.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
		logStream, err := req.Stream(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get logs for pod %s: %v", pod.Name, err)
		}
		defer logStream.Close()

		var logLines []string
		scanner := bufio.NewScanner(logStream)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) != "" {
				logLines = append(logLines, line)
			}
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read logs for pod %s: %v", pod.Name, err)
		}

		logsMap[pod.Name] = logLines
	}

	return logsMap, nil
}

func isPodReadyForLogs(pod corev1.Pod) bool {
	switch pod.Status.Phase {
	case corev1.PodPending:
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Running != nil {
				return true
			}
		}

		return false
	case corev1.PodRunning, corev1.PodSucceeded, corev1.PodFailed:
		return true
	default:
		return false
	}
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
