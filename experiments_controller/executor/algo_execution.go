package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/LGU-SE-Internal/rcabench/client/k8s"
	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/LGU-SE-Internal/rcabench/tracing"
	"github.com/LGU-SE-Internal/rcabench/utils"
	"go.opentelemetry.io/otel/trace"
	corev1 "k8s.io/api/core/v1"
)

type executionPayload struct {
	algorithm dto.AlgorithmItem
	dataset   string
	envVars   map[string]string
}

func executeAlgorithm(ctx context.Context, task *dto.UnifiedTask) error {
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

		record, err := repository.GetDatasetByName(payload.dataset, consts.DatasetBuildSuccess)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to query database for dataset")
			return fmt.Errorf("failed to query database for dataset %s: %v", payload.dataset, err)
		}

		container, err := repository.GetContaineInfo(&dto.GetContainerFilterOptions{
			Type:  consts.ContainerTypeAlgorithm,
			Name:  payload.algorithm.Name,
			Image: payload.algorithm.Image,
			Tag:   payload.algorithm.Tag,
		})
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to get container info for algorithm")
			return fmt.Errorf("failed to get container info for algorithm %s: %v", payload.algorithm.Name, err)
		}

		executionID, err := repository.CreateExecutionResult(task.TaskID, container.Name, record.Name)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to create execution result")
			return fmt.Errorf("failed to create execution result: %v", err)
		}

		itemJson, err := json.Marshal(payload.algorithm)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to marshal algorithm item")
			return fmt.Errorf("failed to marshal algorithm item: %v", err)
		}

		annotations[consts.AnnotationAlgorithm] = string(itemJson)

		jobName := task.TaskID
		fullImage := fmt.Sprintf("%s:%s", container.Image, container.Tag)
		labels := map[string]string{
			consts.LabelTaskID:      task.TaskID,
			consts.LabelTraceID:     task.TraceID,
			consts.LabelGroupID:     task.GroupID,
			consts.LabelTaskType:    string(consts.TaskTypeRunAlgorithm),
			consts.LabelDataset:     payload.dataset,
			consts.LabelExecutionID: strconv.Itoa(executionID),
		}

		return createAlgoJob(ctx, config.GetString("k8s.namespace"), jobName, fullImage, annotations, labels, payload, container, record)
	})
}

// 解析算法执行任务的 Payload
func parseExecutionPayload(payload map[string]any) (*executionPayload, error) {
	message := "missing or invalid '%s' key in payload"

	algorithm, err := utils.ConvertToType[dto.AlgorithmItem](payload[consts.ExecuteAlgorithm])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to AlgorithmItem: %v", consts.ExecuteAlgorithm, err)
	}

	dataset, ok := payload[consts.ExecuteDataset].(string)
	if !ok || dataset == "" {
		return nil, fmt.Errorf(message, consts.ExecuteDataset)
	}

	envVars, err := utils.ConvertToType[map[string]string](payload[consts.ExecuteEnvVars])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to map[string]string: %v", consts.ExecuteEnvVars, err)
	}

	return &executionPayload{
		algorithm: algorithm,
		dataset:   dataset,
		envVars:   envVars,
	}, nil
}

func createAlgoJob(ctx context.Context, jobNamespace, jobName, image string, annotations map[string]string, labels map[string]string, payload *executionPayload, container *database.Container, record *dto.DatasetItemWithID) error {
	return tracing.WithSpan(ctx, func(ctx context.Context) error {
		span := trace.SpanFromContext(ctx)
		restartPolicy := corev1.RestartPolicyNever
		backoffLimit := int32(2)
		parallelism := int32(1)
		completions := int32(1)

		jobEnvVars, err := getAlgoJobEnvVars(payload, container, record)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to get job environment variables")
			return err
		}

		return k8s.CreateJob(ctx, k8s.JobConfig{
			Namespace:     jobNamespace,
			JobName:       jobName,
			Image:         image,
			Command:       strings.Split(container.Command, " "),
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

func getAlgoJobEnvVars(payload *executionPayload, container *database.Container, record *dto.DatasetItemWithID) ([]corev1.EnvVar, error) {
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
		{Name: "INPUT_PATH", Value: fmt.Sprintf("/data/%s", payload.dataset)},
		{Name: "OUTPUT_PATH", Value: fmt.Sprintf("/data/%s", payload.dataset)},
	}

	envNameIndexMap := make(map[string]int, len(jobEnvVars))
	for index, jobEnvVar := range jobEnvVars {
		envNameIndexMap[jobEnvVar.Name] = index
	}

	if payload.envVars != nil {
		extraEnvVarMap := make(map[string]struct{}, len(payload.envVars))
		for name, value := range payload.envVars {
			if index, exists := envNameIndexMap[name]; exists {
				jobEnvVars[index].Value = value
			} else {
				jobEnvVars = append(jobEnvVars, corev1.EnvVar{
					Name:  name,
					Value: value,
				})
				extraEnvVarMap[name] = struct{}{}
			}
		}

		// Check if all required environment variables are provided
		if container.EnvVars != "" {
			envVarsArray := strings.Split(container.EnvVars, ",")
			for _, envVar := range envVarsArray {
				if _, exists := extraEnvVarMap[envVar]; !exists {
					return nil, fmt.Errorf("environment variable %s is required but not provided in algorithm exeuciton payload", envVar)
				}
			}
		}
	} else {
		if container.EnvVars != "" {
			return nil, fmt.Errorf("environment variables %s are required but not provided in algorithm execution payload", container.EnvVars)
		}
	}

	return jobEnvVars, nil
}
