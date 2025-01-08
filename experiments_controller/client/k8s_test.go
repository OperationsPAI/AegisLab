package client

import (
	"context"
	"testing"
	"time"

	"github.com/k0kubun/pp/v3"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestCreateGetDeleteK8sJob(t *testing.T) {
	fc := NewK8sClient()
	jobName := "example-job"
	namespace := "default"
	image := "busybox"
	command := []string{"echo", "hello"}
	restartPolicy := corev1.RestartPolicyNever
	backoffLimit := int32(2)
	parallelism := int32(2)
	completions := int32(2)

	envVars := []corev1.EnvVar{
		{Name: "ENV_TEST", Value: "test"},
	}
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "test-volume",
			MountPath: "/data",
		},
	}
	volumes := []corev1.Volume{
		{
			Name: "test-volume",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: "my-pvc",
				},
			},
		},
	}

	// Create
	if err := CreateK8sJob(fc, namespace, jobName, image, command, restartPolicy,
		backoffLimit, parallelism, completions, envVars, volumeMounts, volumes); err != nil {
		t.Fatalf("CreateK8sJob failed: %v", err)
	}

	// Get
	job, err := GetK8sJob(fc, namespace, jobName)
	if err != nil {
		t.Fatalf("GetK8sJob failed: %v", err)
	}
	pp.Print(job)
	if job.Name != jobName {
		t.Errorf("expected job name %s, got %s", jobName, job.Name)
	}

	time.Sleep(5 * time.Second)
	// Delete
	if err := DeleteK8sJob(fc, namespace, jobName); err != nil {
		t.Fatalf("DeleteK8sJob failed: %v", err)
	}

	// Confirm deletion
	checkJob := &batchv1.Job{}
	err = fc.Get(context.TODO(), client.ObjectKey{Name: jobName, Namespace: namespace}, checkJob)
	if err == nil {
		t.Errorf("expected an error when getting deleted job, but got none")
	} else if !errors.IsNotFound(err) {
		t.Errorf("unexpected error: %v", err)
	}
}
