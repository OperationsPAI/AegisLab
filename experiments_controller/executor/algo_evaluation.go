package executor

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/database"
	corev1 "k8s.io/api/core/v1"

	"gorm.io/gorm"
)

type AlgorithmExecutionPayload struct {
	Benchmark   string `json:"benchmark"`
	Algorithm   string `json:"algorithm"`
	DatasetName string `json:"dataset"`
}

// 解析算法执行任务的 Payload
func parseAlgorithmExecutionPayload(payload map[string]any) (*AlgorithmExecutionPayload, error) {
	benchmark, ok := payload[EvalBench].(string)
	if !ok || benchmark == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", EvalBench)
	}

	algorithm, ok := payload[EvalAlgo].(string)
	if !ok || algorithm == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", EvalAlgo)
	}

	datasetName, ok := payload[EvalDataset].(string)
	if !ok || datasetName == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", EvalDataset)
	}

	return &AlgorithmExecutionPayload{
		Benchmark:   benchmark,
		Algorithm:   algorithm,
		DatasetName: datasetName,
	}, nil
}

func createAlgoJob(ctx context.Context, datasetName, jobName, jobNamespace, image string, command []string, labels map[string]string, jobEnv *client.JobEnv) error {
	restartPolicy := corev1.RestartPolicyNever
	backoffLimit := int32(2)
	parallelism := int32(1)
	completions := int32(1)

	tz := config.GetString("system.timezone")
	if tz == "" {
		tz = "Asia/Shanghai"
	}
	envVars := []corev1.EnvVar{
		{Name: "NORMAL_START", Value: strconv.FormatInt(jobEnv.StartTime.Add(-20*time.Minute).Unix(), 10)},
		{Name: "NORMAL_END", Value: strconv.FormatInt(jobEnv.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_START", Value: strconv.FormatInt(jobEnv.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_END", Value: strconv.FormatInt(jobEnv.EndTime.Unix(), 10)},
		{Name: "INPUT_PATH", Value: fmt.Sprintf("/data/%s", datasetName)},
		{Name: "OUTPUT_PATH", Value: fmt.Sprintf("/data/%s", datasetName)},
		{Name: "TIMEZONE", Value: tz},
		{Name: "WORKSPACE", Value: "/app"},
	}

	return client.CreateK8sJob(ctx, client.JobConfig{
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

func executeAlgorithm(ctx context.Context, task *UnifiedTask) error {
	algoPayload, err := parseAlgorithmExecutionPayload(task.Payload)
	if err != nil {
		return err
	}

	var faultRecord database.FaultInjectionSchedule
	err = database.DB.Where("injection_name = ?", algoPayload.DatasetName).First(&faultRecord).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("no matching fault injection record found for dataset: %s", algoPayload.DatasetName)
		}
		return fmt.Errorf("failed to query database for dataset: %s, error: %v", algoPayload.DatasetName, err)
	}

	startTime := faultRecord.StartTime
	endTime := faultRecord.EndTime

	executionResult := database.ExecutionResult{
		TaskID:  task.TaskID,
		Dataset: faultRecord.ID,
		Algo:    algoPayload.Algorithm,
	}
	if err := database.DB.Create(&executionResult).Error; err != nil {
		return fmt.Errorf("failed to create execution result: %v", err)
	}

	jobName := fmt.Sprintf("%s-%s", algoPayload.Algorithm, algoPayload.DatasetName)
	image := fmt.Sprintf("%s/%s:%s", config.GetString("harbor.repository"), algoPayload.Algorithm, "latest")
	labels := map[string]string{
		LabelTaskID:      task.TaskID,
		LabelTraceID:     task.TraceID,
		LabelGroupID:     task.GroupID,
		LabelTaskType:    string(TaskTypeRunAlgorithm),
		LabelAlgorithm:   algoPayload.Algorithm,
		LabelDataset:     algoPayload.DatasetName,
		LabelExecutionID: fmt.Sprint(executionResult.ID),
	}
	jobEnv := &client.JobEnv{
		StartTime: startTime,
		EndTime:   endTime,
	}

	return createAlgoJob(ctx, algoPayload.DatasetName, jobName, config.GetString("k8s.namespace"), image, []string{"python", "run_exp.py"}, labels, jobEnv)
}
