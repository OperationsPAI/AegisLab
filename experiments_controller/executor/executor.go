package executor

import (
	"fmt"
	"strconv"
	"time"

	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/sirupsen/logrus"
)

type JobLabel struct {
	JobType     string
	TaskID      string
	Algorithm   *string
	Dataset     string
	ExecutionID *int
	StartTime   *time.Time
	EndTime     *time.Time
}

type Executor struct {
}

var Exec *Executor

func parseJobLabel(labels map[string]string) (*JobLabel, error) {
	jobType, ok := labels[LabelJobType]
	if !ok || jobType == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", LabelJobType)
	}

	taskID, ok := labels[LabelTaskID]
	if !ok || taskID == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", LabelTaskID)
	}

	dataset, ok := labels[LabelDataset]
	if !ok || dataset == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", LabelDataset)
	}

	var algorithm *string
	if algo, ok := labels[LabelAlgo]; ok {
		algorithm = &algo
	}

	var executionID *int
	executionIDStr, ok := labels[LabelExecutionID]
	if ok && executionIDStr != "" {
		id, err := strconv.Atoi(executionIDStr)
		if err != nil {
			return nil, fmt.Errorf("missing or invalid '%s' key in payload", LabelExecutionID)
		}
		executionID = &id
	}

	var startTime, endTime *time.Time
	startTimeStr, ok := labels[LabelStartTime]
	if ok && startTimeStr != "" {
		timestamp, err := strconv.Atoi(startTimeStr)
		if err != nil {
			return nil, fmt.Errorf("missing or invalid '%s' key in payload", LabelStartTime)
		}

		parsedTime := time.Unix(int64(timestamp), 0).In(time.FixedZone("HKT", 8*3600))
		startTime = &parsedTime
	}
	endTimeStr, ok := labels[LabelEndTime]
	if ok && endTimeStr != "" {
		timestamp, err := strconv.Atoi(endTimeStr)
		if err != nil {
			return nil, fmt.Errorf("missing or invalid '%s' key in payload", LabelEndTime)
		}

		parsedTime := time.Unix(int64(timestamp), 0).In(time.FixedZone("HKT", 8*3600))
		endTime = &parsedTime
	}

	return &JobLabel{
		JobType:     jobType,
		TaskID:      taskID,
		Algorithm:   algorithm,
		Dataset:     dataset,
		ExecutionID: executionID,
		StartTime:   startTime,
		EndTime:     endTime,
	}, nil
}

func (e *Executor) AddFunc(labels map[string]string) {
	jobLabel, err := parseJobLabel(labels)
	if err != nil {
		logrus.Error(err)
		return
	}

	var message string
	switch jobLabel.JobType {
	case string(TaskTypeBuildDataset):
		message = fmt.Sprintf("Building dataset for task %s", jobLabel.TaskID)
	case string(TaskTypeRunAlgorithm):
		message = fmt.Sprintf("Running algorithm for task %s", jobLabel.TaskID)
	}

	updateTaskStatus(jobLabel.TaskID, TaskStatusRunning, message)
}

func (e *Executor) UpdateFunc(labels map[string]string) {
	jobLabel, err := parseJobLabel(labels)
	if err != nil {
		logrus.Error(err)
		return
	}

	updateTaskStatus(jobLabel.TaskID, TaskStatusCompleted, fmt.Sprintf("Task %s completed", jobLabel.TaskID))

	if jobLabel.JobType == string(TaskTypeBuildDataset) {
		logrus.Infof("Dataset %s built", jobLabel.Dataset)

		if jobLabel.StartTime == nil || jobLabel.EndTime == nil {
			logrus.Errorf("Failed to update record for dataset %s: %v", jobLabel.Dataset, err)
			return
		}

		var faultRecord database.FaultInjectionSchedule
		if err := database.DB.
			Model(&faultRecord).
			Where("injection_name = ?", jobLabel.Dataset).
			Updates(map[string]any{
				"start_time": *jobLabel.StartTime,
				"end_time":   *jobLabel.EndTime,
				"status":     DatasetSuccess,
			}).Error; err != nil {
			logrus.Errorf("Failed to update record for dataset %s: %v", jobLabel.Dataset, err)
		}
		return
	}

	if jobLabel.JobType == string(TaskTypeRunAlgorithm) {
		payload := map[string]any{
			CollectAlgorithm:   *jobLabel.Algorithm,
			CollectDataset:     jobLabel.Dataset,
			CollectExecutionID: *jobLabel.ExecutionID,
		}
		if err := collectResult(jobLabel.TaskID, payload); err != nil {
			logrus.Error(err)
			return
		}

		logrus.Infof("Result of dataset %s collected", jobLabel.Dataset)
		return
	}
}
