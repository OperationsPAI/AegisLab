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
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/repository"
	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"

	"gorm.io/gorm"
)

type ExecutionPayload struct {
	Image   string
	Tag     string
	Dataset string
	EnvVars map[string]string
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

	algorithm := payload.Image
	if payload.EnvVars != nil {
		if algo, ok := payload.EnvVars[consts.ExecuteEnvVarAlgorithm]; ok {
			algorithm = algo
		}
	}

	jobEnvVars := getAlgoJobEnvVars(payload.Dataset, record, payload.EnvVars)

	executionID, err := repository.CreateExecutionResult(algorithm, task.TaskID, record.ID)
	if err != nil {
		return fmt.Errorf("failed to create execution result: %v", err)
	}

	jobName := uuid.New().String()
	image := fmt.Sprintf("%s/%s:%s", config.GetString("harbor.repository"), payload.Image, payload.Tag)
	labels := map[string]string{
		consts.LabelTaskID:      task.TaskID,
		consts.LabelTraceID:     task.TraceID,
		consts.LabelGroupID:     task.GroupID,
		consts.LabelTaskType:    string(consts.TaskTypeRunAlgorithm),
		consts.LabelAlgorithm:   algorithm,
		consts.LabelDataset:     payload.Dataset,
		consts.LabelExecutionID: fmt.Sprint(executionID),
	}
	return createAlgoJob(ctx, config.GetString("k8s.namespace"), jobName, image, labels, jobEnvVars)
}

// 解析算法执行任务的 Payload
func parseExecutionPayload(payload map[string]any) (*ExecutionPayload, error) {
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

	executionPayload := &ExecutionPayload{
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

		executionPayload.EnvVars = envVars
	}

	return executionPayload, nil
}

func createAlgoJob(ctx context.Context, jobNamespace, jobName, image string, labels map[string]string, jobEnvVars []corev1.EnvVar) error {
	restartPolicy := corev1.RestartPolicyNever
	backoffLimit := int32(2)
	parallelism := int32(1)
	completions := int32(1)
	command := []string{"bash", "/entrypoint.sh"}

	return k8s.CreateJob(ctx, k8s.JobConfig{
		Namespace:     jobNamespace,
		JobName:       jobName,
		Image:         image,
		Command:       command,
		RestartPolicy: restartPolicy,
		BackoffLimit:  backoffLimit,
		Parallelism:   parallelism,
		Completions:   completions,
		Labels:        labels,
		EnvVars:       jobEnvVars,
	})
}

func getAlgoJobEnvVars(dataset string, record *dto.DatasetItemWithID, payloadEnvVars map[string]string) []corev1.EnvVar {
	tz := config.GetString("system.timezone")
	if tz == "" {
		tz = "Asia/Shanghai"
	}

	jobEnvVars := []corev1.EnvVar{
		{Name: "TIMEZONE", Value: tz},
		{Name: "NORMAL_START", Value: strconv.FormatInt(record.StartTime.Add(-20*time.Minute).Unix(), 10)},
		{Name: "NORMAL_END", Value: strconv.FormatInt(record.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_START", Value: strconv.FormatInt(record.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_END", Value: strconv.FormatInt(record.EndTime.Unix(), 10)},
		{Name: "WORKSPACE", Value: "/app"},
		{Name: "INPUT_PATH", Value: fmt.Sprintf("/data/%s", dataset)},
		{Name: "OUTPUT_PATH", Value: fmt.Sprintf("/data/%s", dataset)},
	}

	envNameIndexMap := make(map[string]int, len(jobEnvVars))
	for index, jobEnvVar := range jobEnvVars {
		envNameIndexMap[jobEnvVar.Name] = index
	}

	if payloadEnvVars != nil {
		for name, value := range payloadEnvVars {
			if index, ok := envNameIndexMap[name]; ok {
				jobEnvVars[index].Value = value
			} else {
				jobEnvVars = append(jobEnvVars, corev1.EnvVar{
					Name:  name,
					Value: value,
				})
			}
		}
	}

	return jobEnvVars
}
