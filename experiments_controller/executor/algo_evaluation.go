package executor

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client/k8s"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
	corev1 "k8s.io/api/core/v1"

	"gorm.io/gorm"
)

type AlgorithmExecutionMeta struct {
	Benchmark   string
	Algorithm   string
	DatasetName string
	Service     string
	Tag         string
}

func executeAlgorithm(ctx context.Context, task *UnifiedTask) error {
	meta, err := getAlgorithmPayloadMeta(task.Payload)
	if err != nil {
		return err
	}

	var faultRecord database.FaultInjectionSchedule
	err = database.DB.Where("injection_name = ?", meta.DatasetName).First(&faultRecord).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("no matching fault injection record found for dataset: %s", meta.DatasetName)
		}
		return fmt.Errorf("failed to query database for dataset: %s, error: %v", meta.DatasetName, err)
	}

	startTime := faultRecord.StartTime
	endTime := faultRecord.EndTime

	executionResult := database.ExecutionResult{
		TaskID:    task.TaskID,
		Dataset:   faultRecord.ID,
		Algorithm: meta.Algorithm,
	}
	if err := database.DB.Create(&executionResult).Error; err != nil {
		return fmt.Errorf("failed to create execution result: %v", err)
	}

	jobName := fmt.Sprintf("%s-%s", meta.Algorithm, meta.DatasetName)
	image := fmt.Sprintf("%s/%s:%s", config.GetString("harbor.repository"), meta.Algorithm, meta.Tag)
	labels := map[string]string{
		consts.LabelTaskID:      task.TaskID,
		consts.LabelTraceID:     task.TraceID,
		consts.LabelGroupID:     task.GroupID,
		consts.LabelTaskType:    string(consts.TaskTypeRunAlgorithm),
		consts.LabelAlgorithm:   meta.Algorithm,
		consts.LabelDataset:     meta.DatasetName,
		consts.LabelExecutionID: fmt.Sprint(executionResult.ID),
	}
	jobEnv := &k8s.JobEnv{
		Service:   meta.Service,
		StartTime: startTime,
		EndTime:   endTime,
	}

	return createAlgoJob(ctx, meta.DatasetName, jobName, config.GetString("k8s.namespace"), image, []string{"bash", "/entrypoint.sh"}, labels, jobEnv)
}

// 解析算法执行任务的 Payload
func getAlgorithmPayloadMeta(payload map[string]any) (*AlgorithmExecutionMeta, error) {
	message := "missing or invalid '%s' key in payload"

	benchmark, ok := payload[consts.EvalBench].(string)
	if !ok || benchmark == "" {
		return nil, fmt.Errorf(message, consts.EvalBench)
	}

	algorithm, ok := payload[consts.EvalAlgo].(string)
	if !ok || algorithm == "" {
		return nil, fmt.Errorf(message, consts.EvalAlgo)
	}

	datasetName, ok := payload[consts.EvalDataset].(string)
	if !ok || datasetName == "" {
		return nil, fmt.Errorf(message, consts.EvalDataset)
	}

	service, ok := payload[consts.EvalService].(string)
	if !ok {
		return nil, fmt.Errorf(message, consts.EvalService)
	}

	tag, ok := payload[consts.EvalTag].(string)
	if !ok || tag == "" {
		return nil, fmt.Errorf(message, consts.EvalTag)
	}

	return &AlgorithmExecutionMeta{
		Benchmark:   benchmark,
		Algorithm:   algorithm,
		DatasetName: datasetName,
		Service:     service,
		Tag:         tag,
	}, nil
}

func createAlgoJob(ctx context.Context, datasetName, jobName, jobNamespace, image string, command []string, labels map[string]string, jobEnv *k8s.JobEnv) error {
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
		{Name: "SERVICE", Value: jobEnv.Service},
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
