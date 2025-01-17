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
	if faultRecord.Status == DatasetSuccess {
		startTime = faultRecord.StartTime
		endTime = faultRecord.EndTime
	} else if faultRecord.Status == DatasetInitial {
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
	return createDatasetJob(ctx, datasetPayload.DatasetName, fmt.Sprintf("dataset-%s", datasetPayload.DatasetName), config.GetString("k8s.namespace"), "10.10.10.240/library/clickhouse_dataset:latest", []string{"python", "/app/prepare_intputs.py"}, startTime, endTime)
}

func createDatasetJob(ctx context.Context, datasetName, jobname, namespace, image string, command []string, startTime, endTime time.Time) error {
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
		{Name: "INPUT_PATH", Value: fmt.Sprintf("/data/%s", jobname)},
		{Name: "OUTPUT_PATH", Value: fmt.Sprintf("/data/%s", jobname)},
		{Name: "TIMEZONE", Value: tz},
		{Name: "WORKSPACE", Value: "/app"},
	}

	return client.CreateK8sJob(ctx, client.JobConfig{
		Namespace:     namespace,
		JobName:       jobname,
		Image:         image,
		Command:       command,
		RestartPolicy: restartPolicy,
		BackoffLimit:  backoffLimit,
		Parallelism:   parallelism,
		Completions:   completions,
		EnvVars:       envVars,
		Labels:        map[string]string{"job_type": "build_dataset"},
	})
}
