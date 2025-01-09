package executor

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"

	chaosCli "github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/database"
	"gorm.io/gorm"
)

type DatasetPayload struct {
	Benchmark   string
	DatasetName string
}

func parseDatasetPayload(payload map[string]interface{}) (*DatasetPayload, error) {
	datasetName, ok := payload[EvalPayloadDataset].(string)
	if !ok || datasetName == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", EvalPayloadDataset)
	}
	return &DatasetPayload{
		DatasetName: datasetName,
	}, nil
}

func executeBuildDataset(ctx context.Context, taskID string, payload map[string]interface{}) error {
	datasetPayload, err := parseDatasetPayload(payload)
	if err != nil {
		return err
	}

	var faultRecord database.FaultInjectionSchedule
	err = database.DB.Where("injection_name = ?", datasetPayload.DatasetName).First(&faultRecord).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("no matching fault injection record found for dataset: %s", datasetPayload.DatasetName)
		}
		return fmt.Errorf("failed to query database for dataset: %s, error: %v", datasetPayload.DatasetName, err)
	}

	var startTime, endTime time.Time
	if faultRecord.Status == database.DatasetSuccess {
		startTime = faultRecord.StartTime
		endTime = faultRecord.EndTime
	} else if faultRecord.Status == database.DatasetInitial {
		startTime, endTime, err = chaosCli.QueryCRDByName("ts", datasetPayload.DatasetName)
		if err != nil {
			return fmt.Errorf("failed to QueryCRDByName: %s, error: %v", datasetPayload.DatasetName, err)
		}
		if err := database.DB.Model(&faultRecord).Where("injection_name = ?", datasetPayload.DatasetName).
			Updates(map[string]interface{}{
				"start_time": startTime,
				"end_time":   endTime,
			}).Error; err != nil {
			return fmt.Errorf("failed to update start_time and end_time for dataset: %s, error: %v", datasetPayload.DatasetName, err)
		}
	}
	return createDatasetJob(ctx, datasetPayload.DatasetName, "ts", fmt.Sprintf("10.10.10.240/library/clickhouse_dataset:latest"), []string{"python", "/app/prepare_intputs.py"}, startTime, endTime)
}

func createDatasetJob(ctx context.Context, jobname, namespace, image string, command []string, startTime, endTime time.Time) error {
	fc := client.NewK8sClient()
	restartPolicy := corev1.RestartPolicyNever
	backoffLimit := int32(2)
	parallelism := int32(1)
	completions := int32(1)

	tz := config.GetString("system.timezone")
	if tz == "" {
		tz = "Asia/Shanghai"
	}
	envVars := []corev1.EnvVar{
		{Name: "NORMAL_START", Value: strconv.FormatInt(startTime.Add(-20*time.Minute).Unix(), 10)},
		{Name: "NORMAL_END", Value: strconv.FormatInt(startTime.Unix(), 10)},
		{Name: "ABNORMAL_START", Value: strconv.FormatInt(startTime.Unix(), 10)},
		{Name: "ABNORMAL_END", Value: strconv.FormatInt(endTime.Unix(), 10)},
		{Name: "OUTPUT_PATH", Value: fmt.Sprintf("/data/%s", jobname)},
		{Name: "TIMEZONE", Value: tz},
		{Name: "WORKSPACE", Value: "/app"},
	}
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

	err := client.CreateK8sJob(ctx, fc, namespace, jobname, image, command, restartPolicy,
		backoffLimit, parallelism, completions, envVars, volumeMounts, volumes)
	return err
}
