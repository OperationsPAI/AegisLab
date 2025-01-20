package executor

import (
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"
)

type Executor struct {
}

var Exec *Executor

type JobLabel struct {
	Dataset     string
	ExecutionID int
	JobType     string
	TaskID      string
}

func parseJobLabel(labels map[string]string) (*JobLabel, error) {
	dataset := labels[LabelDataset]
	if dataset == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", LabelDataset)
	}
	executionIDStr := labels[LabelExecutionID]
	executionID, err := strconv.Atoi(executionIDStr)
	if err != nil || executionIDStr == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", LabelExecutionID)
	}
	jobType := labels[LabelJobType]
	if jobType == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", LabelJobType)
	}
	taskID := labels[LabelTaskID]
	if taskID == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", LabelTaskID)
	}
	return &JobLabel{
		Dataset:     dataset,
		ExecutionID: executionID,
		JobType:     jobType,
		TaskID:      taskID,
	}, nil
}

func (e *Executor) AddFunc(labels map[string]string) {
	jobLabel, err := parseJobLabel(labels)
	if err != nil {
		logrus.Error(err)
	}

	var message string
	switch labels[LabelJobType] {
	case string(TaskTypeRunAlgorithm):
		message = fmt.Sprintf("Running algorithm for task %s", jobLabel.TaskID)
	}
	updateTaskStatus(jobLabel.TaskID, TaskStatusRunning, message)
}

func (e *Executor) UpdateFunc(labels map[string]string) {
	jobLabel, err := parseJobLabel(labels)
	if err != nil {
		logrus.Error(err)
	}

	updateTaskStatus(jobLabel.TaskID, "Completed", fmt.Sprintf("Task %s completed", jobLabel.TaskID))

	if labels[LabelJobType] == string(TaskTypeRunAlgorithm) {
		dataset := labels[LabelDataset]
		payload := map[string]interface{}{
			CollectDataset:     dataset,
			CollectExecutionID: labels[LabelExecutionID],
		}
		if err := collectResult(jobLabel.TaskID, payload); err != nil {
			logrus.Error(err)
			return
		}

		logrus.Infof("Result of dataset %s collected", dataset)
	}
}
