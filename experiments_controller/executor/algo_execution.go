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
	"github.com/CUHK-SE-Group/rcabench/repository"
	corev1 "k8s.io/api/core/v1"

	"gorm.io/gorm"
)

type ExecutionPayload struct {
	Algorithm string
	Dataset   string
	Service   string
	Tag       string
}

func executeAlgorithm(ctx context.Context, task *UnifiedTask) error {
	payload, err := parseExecutionPayload(task.Payload)
	if err != nil {
		return err
	}

	record, err := repository.GetDatasetByName(payload.Dataset, consts.DatasetBuildSuccess)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("no matching dataset %s found", payload.Dataset)
		}
		return fmt.Errorf("failed to query database for dataset %s: %v", payload.Dataset, err)
	}

	executionID, err := repository.CreateExecutionResult(payload.Algorithm, task.TaskID, record.ID)
	if err != nil {
		return fmt.Errorf("failed to create execution result: %v", err)
	}

	jobName := fmt.Sprintf("%s-%s", payload.Algorithm, payload.Dataset)
	image := fmt.Sprintf("%s/%s:%s", config.GetString("harbor.repository"), payload.Algorithm, payload.Tag)
	labels := map[string]string{
		consts.LabelTaskID:      task.TaskID,
		consts.LabelTraceID:     task.TraceID,
		consts.LabelGroupID:     task.GroupID,
		consts.LabelTaskType:    string(consts.TaskTypeRunAlgorithm),
		consts.LabelDataset:     payload.Dataset,
		consts.LabelAlgorithm:   payload.Algorithm,
		consts.LabelExecutionID: fmt.Sprint(executionID),
	}
	jobEnv := &k8s.JobEnv{
		Service:   payload.Service,
		StartTime: record.StartTime,
		EndTime:   record.EndTime,
	}

	return createAlgoJob(ctx, payload.Dataset, jobName, config.GetString("k8s.namespace"), image, []string{"bash", "/entrypoint.sh"}, labels, jobEnv)
}

// 解析算法执行任务的 Payload
func parseExecutionPayload(payload map[string]any) (*ExecutionPayload, error) {
	message := "missing or invalid '%s' key in payload"

	algorithm, ok := payload[consts.ExecuteAlgo].(string)
	if !ok || algorithm == "" {
		return nil, fmt.Errorf(message, consts.ExecuteAlgo)
	}

	dataset, ok := payload[consts.ExecuteDataset].(string)
	if !ok || dataset == "" {
		return nil, fmt.Errorf(message, consts.ExecuteDataset)
	}

	service, ok := payload[consts.ExecuteService].(string)
	if !ok {
		return nil, fmt.Errorf(message, consts.ExecuteService)
	}

	tag, ok := payload[consts.ExecuteTag].(string)
	if !ok || tag == "" {
		return nil, fmt.Errorf(message, consts.ExecuteTag)
	}

	return &ExecutionPayload{
		Algorithm: algorithm,
		Dataset:   dataset,
		Service:   service,
		Tag:       tag,
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
