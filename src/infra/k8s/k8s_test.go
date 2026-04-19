package k8s

import (
	"aegis/config"
	"aegis/utils"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/k0kubun/pp/v3"
	corev1 "k8s.io/api/core/v1"
)

const (
	runK8sIntegrationEnv          = "RUN_K8S_INTEGRATION"
	runK8sIntegrationNamespaceEnv = "RUN_K8S_INTEGRATION_NAMESPACE"
	runK8sIntegrationImageEnv     = "RUN_K8S_INTEGRATION_IMAGE"
	runK8sIntegrationKeepJobEnv   = "RUN_K8S_INTEGRATION_KEEP_JOB"
)

type integrationConfig struct {
	namespace string
	image     string
	keepJob   bool
}

func TestGetVolumeMountConfigs(t *testing.T) {
	config.Init("../..")

	volumeMountConfigs := make([]VolumeMountConfig, 0)
	mapData := config.GetMap("k8s.job.volume_mount")
	for _, cfgData := range mapData {
		cfg, err := utils.ConvertToType[VolumeMountConfig](cfgData)
		if err != nil {
			t.Errorf("invalid volume mount config %v: %v", cfgData, err)
		}

		volumeMountConfigs = append(volumeMountConfigs, cfg)
	}

	volumeMounts := []corev1.VolumeMount{}
	volumes := []corev1.Volume{}
	for _, cfg := range volumeMountConfigs {
		volumeMounts = append(volumeMounts, cfg.GetVolumeMount())
		volumes = append(volumes, cfg.GetVolume())
	}

	pp.Println(volumeMountConfigs) //nolint:errcheck
	pp.Println(volumeMounts)       //nolint:errcheck
	pp.Println(volumes)            //nolint:errcheck
}

func requireIntegrationConfig(t *testing.T) integrationConfig {
	t.Helper()

	if os.Getenv(runK8sIntegrationEnv) != "1" {
		t.Skipf(
			"set %s=1 to run Kubernetes integration test (optional overrides: %s, %s, %s)",
			runK8sIntegrationEnv,
			runK8sIntegrationNamespaceEnv,
			runK8sIntegrationImageEnv,
			runK8sIntegrationKeepJobEnv,
		)
	}

	config.Init("../..")

	namespace := config.GetString("k8s.namespace")
	if override := strings.TrimSpace(os.Getenv(runK8sIntegrationNamespaceEnv)); override != "" {
		namespace = override
	}
	if namespace == "" {
		t.Fatal("kubernetes integration namespace is empty; set k8s.namespace or RUN_K8S_INTEGRATION_NAMESPACE")
	}

	image := strings.TrimSpace(os.Getenv(runK8sIntegrationImageEnv))
	if image == "" {
		image = "busybox:1.36"
	}

	return integrationConfig{
		namespace: namespace,
		image:     image,
		keepJob:   os.Getenv(runK8sIntegrationKeepJobEnv) == "1",
	}
}

func TestK8sGatewayJobLifecycleIntegration(t *testing.T) {
	cfg := requireIntegrationConfig(t)

	gateway := NewGateway(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := gateway.CheckHealth(ctx); err != nil {
		t.Fatalf("kubernetes health precheck failed: %v", err)
	}

	jobName := fmt.Sprintf("aegis-k8s-integration-%d", time.Now().UnixNano())
	command := []string{"sh", "-c", "for i in $(seq 1 5); do echo \"Log line $i\"; sleep 1; done"}
	restartPolicy := corev1.RestartPolicyNever
	backoffLimit := int32(1)
	parallelism := int32(1)
	completions := int32(1)

	envVars := []corev1.EnvVar{
		{Name: "AEGIS_K8S_INTEGRATION", Value: "true"},
	}

	if !cfg.keepJob {
		t.Cleanup(func() {
			cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cleanupCancel()
			_ = gateway.DeleteJob(cleanupCtx, cfg.namespace, jobName)
		})
	}

	t.Logf("running Kubernetes integration against namespace=%s image=%s", cfg.namespace, cfg.image)

	if err := gateway.CreateJob(ctx, &JobConfig{
		Namespace:     cfg.namespace,
		JobName:       jobName,
		Image:         cfg.image,
		Command:       command,
		RestartPolicy: restartPolicy,
		BackoffLimit:  backoffLimit,
		Parallelism:   parallelism,
		Completions:   completions,
		EnvVars:       envVars,
	}); err != nil {
		t.Fatalf("create job failed: %v", err)
	}
	t.Logf("job %s created successfully", jobName)

	job, err := gateway.GetJob(ctx, cfg.namespace, jobName)
	if err != nil {
		t.Fatalf("get job failed: %v", err)
	}
	if job.Name != jobName {
		t.Errorf("expected job name %s, got %s", jobName, job.Name)
	}

	t.Logf("waiting for job %s to complete", jobName)
	if err := gateway.WaitForJobCompletion(ctx, cfg.namespace, jobName); err != nil {
		t.Fatalf("wait for job completion failed: %v", err)
	}
	t.Logf("job %s completed successfully", jobName)

	logs, err := gateway.GetJobPodLogs(ctx, cfg.namespace, jobName)
	if err != nil {
		t.Fatalf("get job pod logs failed: %v", err)
	}
	if len(logs) == 0 {
		t.Fatalf("expected logs for job %s, got none", jobName)
	}

	foundLogLine := false
	for podName, podLogs := range logs {
		t.Logf("pod %s emitted %d log lines", podName, len(podLogs))
		for _, line := range podLogs {
			if strings.Contains(line, "Log line") {
				foundLogLine = true
				break
			}
		}
		if foundLogLine {
			break
		}
	}
	if !foundLogLine {
		t.Fatalf("expected job logs to include test output, got %#v", logs)
	}

	if cfg.keepJob {
		t.Logf("keeping job %s because %s=1", jobName, runK8sIntegrationKeepJobEnv)
		return
	}

	if err := gateway.DeleteJob(ctx, cfg.namespace, jobName); err != nil {
		t.Fatalf("delete job failed: %v", err)
	}
	t.Logf("job %s and its associated pods deleted successfully", jobName)
}
