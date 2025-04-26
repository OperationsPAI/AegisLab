package executor

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client/k8s"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/repository"
	"github.com/CUHK-SE-Group/rcabench/tracing"
	"go.opentelemetry.io/otel/trace"
	corev1 "k8s.io/api/core/v1"
)

type executionPayload struct {
	Image   string
	Tag     string
	Dataset string
	EnvVars map[string]string
}

func executeAlgorithm(ctx context.Context, task *UnifiedTask) error {
	return tracing.WithSpan(ctx, func(ctx context.Context) error {
		span := trace.SpanFromContext(ctx)
		payload, err := parseExecutionPayload(task.Payload)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to parse execution payload")
			return err
		}

		annotations, err := getAnnotations(ctx, task)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to get annotations")
			return err
		}

		record, err := repository.GetDatasetByName(payload.Dataset, consts.DatasetBuildSuccess)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to query database for dataset")
			return fmt.Errorf("failed to query database for dataset %s: %v", payload.Dataset, err)
		}

		algorithm := payload.Image
		if payload.EnvVars != nil {
			if algo, ok := payload.EnvVars[consts.ExecuteEnvVarAlgorithm]; ok {
				algorithm = algo
			}
		}

		executionID, err := repository.CreateExecutionResult(algorithm, task.TaskID, record.ID)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to create execution result")
			return fmt.Errorf("failed to create execution result: %v", err)
		}

		jobName := task.TaskID
		image := fmt.Sprintf("%s/%s:%s", config.GetString("harbor.repository"), payload.Image, payload.Tag)
		labels := map[string]string{
			consts.LabelTaskID:      task.TaskID,
			consts.LabelTraceID:     task.TraceID,
			consts.LabelGroupID:     task.GroupID,
			consts.LabelTaskType:    string(consts.TaskTypeRunAlgorithm),
			consts.LabelAlgorithm:   algorithm,
			consts.LabelDataset:     payload.Dataset,
			consts.LabelExecutionID: strconv.Itoa(executionID),
		}

		return createAlgoJob(ctx, config.GetString("k8s.namespace"), jobName, image, annotations, labels, payload, record)
	})
}

// 解析算法执行任务的 Payload
func parseExecutionPayload(payload map[string]any) (*executionPayload, error) {
	message := "missing or invalid '%s' key in payload"

	image, ok := payload[consts.ExecuteImage].(string)
	if !ok || image == "" {
		return nil, fmt.Errorf(message, consts.ExecuteImage)
	}

	tag, ok := payload[consts.ExecuteTag].(string)
	if !ok || tag == "" {
		return nil, fmt.Errorf(message, consts.ExecuteTag)
	}

	dataset, ok := payload[consts.ExecuteDataset].(string)
	if !ok || dataset == "" {
		return nil, fmt.Errorf(message, consts.ExecuteDataset)
	}

	result := executionPayload{
		Image:   image,
		Tag:     tag,
		Dataset: dataset,
	}
	if e, exists := payload[consts.ExecuteEnvVars].(map[string]any); exists {
		envVars := make(map[string]string, len(e))
		for key, value := range e {
			strValue, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf(message, consts.ExecuteEnvVars)
			}
			envVars[key] = strValue
		}

		result.EnvVars = envVars
	}

	return &result, nil
}

func createAlgoJob(ctx context.Context, jobNamespace, jobName, image string, annotations map[string]string, labels map[string]string, payload *executionPayload, record *dto.DatasetItemWithID) error {
	return tracing.WithSpan(ctx, func(ctx context.Context) error {
		span := trace.SpanFromContext(ctx)
		restartPolicy := corev1.RestartPolicyNever
		backoffLimit := int32(2)
		parallelism := int32(1)
		completions := int32(1)
		command := []string{"bash", "/entrypoint.sh"}

		jobEnvVars, err := getAlgoJobEnvVars(payload, record)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to get job environment variables")
			return err
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
			Annotations:   annotations,
			Labels:        labels,
			EnvVars:       jobEnvVars,
		})
	})
}

func getAlgoJobEnvVars(payload *executionPayload, record *dto.DatasetItemWithID) ([]corev1.EnvVar, error) {
	tz := config.GetString("system.timezone")
	if tz == "" {
		tz = "Asia/Shanghai"
	}

	preDuration, ok := record.Param["pre_duration"].(int)
	if !ok || preDuration == 0 {
		return nil, fmt.Errorf("failed to get the preduration")
	}

	jobEnvVars := []corev1.EnvVar{
		{Name: "TIMEZONE", Value: tz},
		{Name: "NORMAL_START", Value: strconv.FormatInt(record.StartTime.Add(-time.Duration(preDuration)*time.Minute).Unix(), 10)},
		{Name: "NORMAL_END", Value: strconv.FormatInt(record.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_START", Value: strconv.FormatInt(record.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_END", Value: strconv.FormatInt(record.EndTime.Unix(), 10)},
		{Name: "WORKSPACE", Value: "/app"},
		{Name: "INPUT_PATH", Value: fmt.Sprintf("/data/%s", payload.Dataset)},
		{Name: "OUTPUT_PATH", Value: fmt.Sprintf("/data/%s", payload.Dataset)},
	}

	envNameIndexMap := make(map[string]int, len(jobEnvVars))
	for index, jobEnvVar := range jobEnvVars {
		envNameIndexMap[jobEnvVar.Name] = index
	}

	if payload.EnvVars != nil {
		for name, value := range payload.EnvVars {
			if index, exists := envNameIndexMap[name]; exists {
				jobEnvVars[index].Value = value
			} else {
				jobEnvVars = append(jobEnvVars, corev1.EnvVar{
					Name:  name,
					Value: value,
				})
			}
		}
	}

	return jobEnvVars, nil
}
