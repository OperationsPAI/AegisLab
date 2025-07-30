package k8s

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestCreateGetDeleteK8sJob(t *testing.T) {
	jobName := "example-job"
	namespace := "default"
	image := "busybox"
	command := []string{"sh", "-c", "for i in $(seq 1 5); do echo \"Log line $i\"; sleep 1; done"}
	restartPolicy := corev1.RestartPolicyNever
	backoffLimit := int32(2)
	parallelism := int32(2)
	completions := int32(2)

	envVars := []corev1.EnvVar{
		{Name: "ENV_TEST", Value: "test"},
	}

	// Step 1: Create Job
	if err := CreateJob(context.Background(), JobConfig{
		Namespace:     namespace,
		JobName:       jobName,
		Image:         image,
		Command:       command,
		RestartPolicy: restartPolicy,
		BackoffLimit:  backoffLimit,
		Parallelism:   parallelism,
		Completions:   completions,
		EnvVars:       envVars,
	}); err != nil {
		t.Fatalf("CreateK8sJob failed: %v", err)
	}
	t.Logf("Job %s created successfully.", jobName)

	// Step 2: Get Job
	job, err := GetJob(context.Background(), namespace, jobName)
	if err != nil {
		t.Fatalf("GetK8sJob failed: %v", err)
	}
	t.Logf("Fetched job: %v", job)

	// Ensure job was created with the correct name
	if job.Name != jobName {
		t.Errorf("expected job name %s, got %s", jobName, job.Name)
	}

	// Step 3: Wait for Job completion
	t.Logf("Waiting for job %s to complete...", jobName)
	if err := WaitForJobCompletion(context.Background(), namespace, jobName); err != nil {
		t.Fatalf("WaitForJobCompletion failed: %v", err)
	}
	t.Logf("Job %s completed successfully.", jobName)

	// Step 4: Get Pod Logs
	logs, err := GetJobPodLogs(context.Background(), namespace, jobName)
	if err != nil {
		t.Fatalf("GetJobPodLogs failed: %v", err)
	}

	t.Logf("Logs for job %s:\n", jobName)
	for podName, log := range logs {
		t.Logf("Pod %s logs:\n%s", podName, log)
	}

	// Step 5: Delete Job
	if err := deleteJob(context.Background(), namespace, jobName); err != nil {
		t.Fatalf("DeleteK8sJob failed: %v", err)
	}
	t.Logf("Job %s and its associated pods deleted successfully.", jobName)
}
