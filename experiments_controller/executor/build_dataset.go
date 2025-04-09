package executor

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"

	"github.com/CUHK-SE-Group/rcabench/client/k8s"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
)

type DatasetMeta struct {
	Benchmark   string
	Namespace   string
	Name        string
	PreDuration int
	Service     string
	StartTime   time.Time
	EndTime     time.Time
}

func executeBuildDataset(ctx context.Context, task *UnifiedTask) error {
	datasetMeta, err := parseDatasetPayload(task.Payload)
	if err != nil {
		return err
	}

	jobName := fmt.Sprintf("%s-%s", consts.DatasetJobName, datasetMeta.Name)
	image := fmt.Sprintf("%s/%s_dataset:%s", config.GetString("harbor.repository"), datasetMeta.Benchmark, config.GetString("image.tag"))
	labels := map[string]string{
		consts.LabelTaskID:   task.TaskID,
		consts.LabelTraceID:  task.TraceID,
		consts.LabelGroupID:  task.GroupID,
		consts.LabelTaskType: string(consts.TaskTypeBuildDataset),
		consts.LabelDataset:  datasetMeta.Name,
	}
	jobEnv := &k8s.JobEnv{
		Namespace:   datasetMeta.Namespace,
		Service:     datasetMeta.Service,
		PreDuration: datasetMeta.PreDuration,
		StartTime:   datasetMeta.StartTime,
		EndTime:     datasetMeta.EndTime,
	}

	return createDatasetJob(ctx, datasetMeta.Name, jobName, config.GetString("k8s.namespace"), image, []string{"python", "prepare_inputs.py"}, labels, jobEnv)
}

func parseDatasetPayload(payload map[string]any) (*DatasetMeta, error) {
	message := "missing or invalid '%s' key in payload"

	benchmark, ok := payload[consts.BuildBenchmark].(string)
	if !ok || benchmark == "" {
		return nil, fmt.Errorf(message, consts.BuildBenchmark)
	}

	datasetName, ok := payload[consts.BuildDataset].(string)
	if !ok || datasetName == "" {
		return nil, fmt.Errorf(message, consts.BuildDataset)
	}

	namespace, ok := payload[consts.BuildNamespace].(string)
	if !ok || namespace == "" {
		return nil, fmt.Errorf(message, consts.BuildNamespace)
	}

	preDurationFloat, ok := payload[consts.BuildPreDuration].(float64)
	if !ok || preDurationFloat <= 0 {
		return nil, fmt.Errorf(message, consts.BuildPreDuration)
	}
	preDuration := int(preDurationFloat)

	service, ok := payload[consts.BuildService].(string)
	if !ok || namespace == "" {
		return nil, fmt.Errorf(message, consts.BuildService)
	}

	startTimePtr, err := parseTimeFromPayload(payload, consts.BuildStartTime)
	if err != nil {
		return nil, fmt.Errorf(message, consts.BuildStartTime)
	}

	endTimePtr, err := parseTimeFromPayload(payload, consts.BuildEndTime)
	if err != nil {
		return nil, fmt.Errorf(message, consts.BuildEndTime)
	}

	var startTime, endTime time.Time
	if startTimePtr != nil && endTimePtr != nil {
		startTime = *startTimePtr
		endTime = *endTimePtr
	} else {
		var fiRecord database.FaultInjectionSchedule
		if err := database.DB.
			Where("injection_name = ? and status = ?", datasetName, consts.DatasetInjectSuccess).
			First(&fiRecord).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("find matching fault injection record found failed")
			}

			return nil, fmt.Errorf("query database for dataset failed: %v", err)
		}

		startTime = fiRecord.StartTime
		endTime = fiRecord.EndTime
	}

	return &DatasetMeta{
		Benchmark:   benchmark,
		Namespace:   namespace,
		Name:        datasetName,
		PreDuration: preDuration,
		Service:     service,
		StartTime:   startTime,
		EndTime:     endTime,
	}, nil
}

func parseTimeFromPayload(payload map[string]any, key string) (*time.Time, error) {
	timeStr, ok := payload[key].(string)
	if !ok {
		return nil, fmt.Errorf("%s must be a string", key)
	}

	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid %s format: %v", key, err)
	}

	return &t, nil
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
		{Name: "NORMAL_START", Value: strconv.FormatInt(jobEnv.StartTime.Add(-time.Duration(jobEnv.PreDuration)*time.Minute).Unix(), 10)},
		{Name: "NORMAL_END", Value: strconv.FormatInt(jobEnv.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_START", Value: strconv.FormatInt(jobEnv.StartTime.Unix(), 10)},
		{Name: "ABNORMAL_END", Value: strconv.FormatInt(jobEnv.EndTime.Unix(), 10)},
		{Name: "INPUT_PATH", Value: fmt.Sprintf("/data/%s", datasetName)},
		{Name: "OUTPUT_PATH", Value: fmt.Sprintf("/data/%s", datasetName)},
		{Name: "NAMESPACE", Value: jobEnv.Namespace},
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
