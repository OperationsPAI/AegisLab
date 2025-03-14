package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/CUHK-SE-Group/rcabench/client/k8s"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DatasetPayload struct {
	Benchmark   string     `json:"benchmark"`
	DatasetName string     `json:"dataset"`
	Namespace   string     `json:"namespace"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
}

func parseDatasetPayload(payload map[string]any) (*DatasetPayload, error) {
	benchmark, ok := payload[BuildBenchmark].(string)
	if !ok || benchmark == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", BuildBenchmark)
	}

	datasetName, ok := payload[BuildDataset].(string)
	if !ok || datasetName == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", BuildDataset)
	}

	namespace, ok := payload[BuildNamespace].(string)
	if !ok || namespace == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", BuildNamespace)
	}

	var startTime, endTime *time.Time
	startTimeStr, ok := payload[BuildStartTime].(string)
	if !ok {
		parsedTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return nil, fmt.Errorf("missing or invalid '%s' key in payload", BuildStartTime)
		}

		startTime = &parsedTime
	}
	endTimeStr, ok := payload[BuildEndTime].(string)
	if !ok {
		parsedTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return nil, fmt.Errorf("missing or invalid '%s' key in payload", BuildEndTime)
		}

		endTime = &parsedTime
	}

	return &DatasetPayload{
		Benchmark:   benchmark,
		DatasetName: datasetName,
		Namespace:   namespace,
		StartTime:   startTime,
		EndTime:     endTime,
	}, nil
}

func createDatasetJob(ctx context.Context, datasetName, jobName, jobNamespace, image string, command []string, labels map[string]string, jobEnv *k8s.JobEnv) error {
	restartPolicy := corev1.RestartPolicyNever
	backoffLimit := int32(2)
	parallelism := int32(1)
	completions := int32(1)

	tz := config.GetString("system.timezone")
	if tz == "" {
		tz = "Asia/Shanghai"
	}
	envVars := []corev1.EnvVar{
		{Name: "NORMAL_START", Value: strconv.FormatInt(jobEnv.StartTime.Add(-time.Duration(config.GetInt("injection.interval"))*time.Minute).Unix(), 10)},
		{Name: "NORMAL_END", Value: strconv.FormatInt(jobEnv.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_START", Value: strconv.FormatInt(jobEnv.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_END", Value: strconv.FormatInt(jobEnv.EndTime.Unix(), 10)},
		{Name: "INPUT_PATH", Value: fmt.Sprintf("/data/%s", datasetName)},
		{Name: "OUTPUT_PATH", Value: fmt.Sprintf("/data/%s", datasetName)},
		{Name: "NAMESPACE", Value: jobEnv.Namespace},
		{Name: "TIMEZONE", Value: tz},
		{Name: "WORKSPACE", Value: "/app"},
	}

	return k8s.CreateJob(ctx, k8s.JobConfig{
		Namespace:     jobNamespace,
		JobName:       jobName,
		Image:         image,
		Command:       command,
		RestartPolicy: restartPolicy,
		BackoffLimit:  backoffLimit,
		Parallelism:   parallelism,
		Completions:   completions,
		EnvVars:       envVars,
		Labels:        labels,
	})
}

func executeBuildDataset(ctx context.Context, task *UnifiedTask) error {
	datasetPayload, err := parseDatasetPayload(task.Payload)
	if err != nil {
		return err
	}

	datasetName := datasetPayload.DatasetName

	var startTime, endTime time.Time
	if datasetPayload.StartTime != nil && datasetPayload.EndTime != nil {
		startTime = *datasetPayload.StartTime
		endTime = *datasetPayload.EndTime
	} else {
		var faultRecord database.FaultInjectionSchedule
		err = database.DB.Where("injection_name = ?", datasetName).First(&faultRecord).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("No matching fault injection record found for dataset: %s", datasetName)
			}
			return fmt.Errorf("failed to query database for dataset: %s, error: %v", datasetName, err)
		}

		var fiPayload FaultInjectionPayload
		if err = json.Unmarshal([]byte(faultRecord.Config), &fiPayload); err != nil {
			return fmt.Errorf("Failed to unmarshal fault injection payload for dataset %s: %v", datasetName, err)
		}
		logrus.Infof("Parsed fault injection payload: %+v", fiPayload)

		startTime, endTime, err = checkExecutionTime(faultRecord, fiPayload.Namespace)
		if err != nil {
			return fmt.Errorf("Failed to checkExecutionTime for dataset %s: %v", datasetName, err)
		}
	}

	jobName := fmt.Sprintf("dataset-%s", datasetName)
	image := fmt.Sprintf("%s/%s_dataset:latest", config.GetString("harbor.repository"), datasetPayload.Benchmark)
	labels := map[string]string{
		LabelTaskID:    task.TaskID,
		LabelTraceID:   task.TraceID,
		LabelGroupID:   task.GroupID,
		LabelTaskType:  string(TaskTypeBuildDataset),
		LabelDataset:   datasetPayload.DatasetName,
		LabelStartTime: strconv.FormatInt(startTime.Unix(), 10),
		LabelEndTime:   strconv.FormatInt(endTime.Unix(), 10),
	}
	jobEnv := &k8s.JobEnv{
		Namespace: datasetPayload.Namespace,
		StartTime: startTime,
		EndTime:   endTime,
	}

	return createDatasetJob(ctx, datasetName, jobName, config.GetString("k8s.namespace"), image, []string{"python", "prepare_inputs.py"}, labels, jobEnv)
}
