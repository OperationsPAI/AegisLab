package executor

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/sirupsen/logrus"
)

type JobLabel struct {
	TaskID      string
	TraceID     string
	GroupID     string
	Type        TaskType
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
	taskID, ok := labels[LabelTaskID]
	if !ok || taskID == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", LabelTaskID)
	}

	traceID, ok := labels[LabelTraceID]
	if !ok || traceID == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", LabelTraceID)
	}

	groupID, ok := labels[LabelGroupID]
	if !ok || groupID == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", LabelGroupID)
	}

	taskType, ok := labels[LabelTaskType]
	if !ok || taskType == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", LabelTaskType)
	}

	dataset, ok := labels[LabelDataset]
	if !ok || dataset == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", LabelDataset)
	}

	var algorithm *string
	if algo, ok := labels[LabelAlgorithm]; ok {
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
		TaskID:      taskID,
		TraceID:     traceID,
		GroupID:     groupID,
		Type:        TaskType(taskType),
		Algorithm:   algorithm,
		Dataset:     dataset,
		ExecutionID: executionID,
		StartTime:   startTime,
		EndTime:     endTime,
	}, nil
}

func (e *Executor) HandleJobAdd(labels map[string]string) {
	jobLabel, err := parseJobLabel(labels)
	if err != nil {
		logrus.Error(err)
		return
	}

	var message string
	switch jobLabel.Type {
	case TaskTypeBuildDataset:
		message = fmt.Sprintf("Building dataset for task %s", jobLabel.TaskID)
	case TaskTypeRunAlgorithm:
		message = fmt.Sprintf("Running algorithm for task %s", jobLabel.TaskID)
	}

	updateTaskStatus(jobLabel.TaskID, jobLabel.TraceID,
		message,
		map[string]any{
			RdbMsgStatus:   TaskStatusRunning,
			RdbMsgTaskType: jobLabel.Type,
		})
}

func (e *Executor) HandleJobUpdate(labels map[string]string, status string) {
	jobLabel, err := parseJobLabel(labels)
	if err != nil {
		logrus.Error(err)
		return
	}

	if status == TaskStatusCompleted {
		if jobLabel.Type == TaskTypeBuildDataset {
			logrus.Infof(fmt.Sprintf("Dataset %s built", jobLabel.Dataset))

			updateTaskStatus(jobLabel.TaskID, jobLabel.TraceID,
				fmt.Sprintf(TaskMsgCompleted, jobLabel.TaskID),
				map[string]any{
					RdbMsgStatus:   TaskStatusCompleted,
					RdbMsgTaskType: jobLabel.Type,
					RdbMsgDataset:  jobLabel.Dataset,
				})

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

		if jobLabel.Type == TaskTypeRunAlgorithm {
			algorithm := *jobLabel.Algorithm
			executionID := *jobLabel.ExecutionID
			logrus.Infof(fmt.Sprintf("Algorithm %s executed", algorithm))

			updateTaskStatus(jobLabel.TaskID, jobLabel.TraceID,
				fmt.Sprintf(TaskMsgCompleted, jobLabel.TaskID),
				map[string]any{
					RdbMsgStatus:      TaskStatusCompleted,
					RdbMsgTaskType:    jobLabel.Type,
					RdbMsgExecutionID: executionID,
				})

			payload := map[string]any{
				CollectAlgorithm:   algorithm,
				CollectDataset:     jobLabel.Dataset,
				CollectExecutionID: executionID,
			}
			if _, _, err := SubmitTask(context.Background(), &UnifiedTask{
				Type:      TaskTypeCollectResult,
				Payload:   payload,
				Immediate: true,
				TraceID:   jobLabel.TraceID,
				GroupID:   jobLabel.GroupID,
			}); err != nil {
				logrus.Error(err)
				return
			}

			return
		}
	}

	if status == TaskStatusError {
	}
}

func (e *Executor) HandlePodUpdate() {

}
