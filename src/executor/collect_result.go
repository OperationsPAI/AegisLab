package executor

import (
	"context"
	"fmt"

	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	"aegis/repository"
	"aegis/tracing"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type collectionPayload struct {
	algorithm   dto.AlgorithmItem
	dataset     string
	executionID int
	timestamp   string
}

func executeCollectResult(ctx context.Context, task *dto.UnifiedTask) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)

		logEntry := logrus.WithField("task_id", task.TaskID)

		collectPayload, err := parseCollectPayload(task.Payload)
		if err != nil {
			return err
		}

		if collectPayload.algorithm.Name == config.GetString("algo.detector") {
			results, err := repository.ListDetectorResultsByExecutionID(collectPayload.executionID)
			if err != nil {
				span.AddEvent("failed to get detector results by execution ID")
				span.RecordError(err)
				return fmt.Errorf("failed to get detector results by execution ID: %v", err)
			}

			if len(results) == 0 {
				repository.PublishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
					TaskID:    task.TaskID,
					TaskType:  consts.TaskTypeCollectResult,
					EventName: consts.EventDatapackNoDetectorData,
				})
				span.AddEvent("no detector results found for the execution ID")
				logEntry.Info("no detector results found for the execution ID")
				return nil
			}

			hasIssues := false
			for _, v := range results {
				if v.Issues != "{}" {
					hasIssues = true
				}
			}

			if !hasIssues {
				repository.PublishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
					TaskID:    task.TaskID,
					TaskType:  consts.TaskTypeCollectResult,
					EventName: consts.EventDatapackNoAnomaly,
					Payload:   results,
				})

				span.AddEvent("the detector result has no issues")
				logEntry.Info("the detector result has no issues")
			} else {
				repository.PublishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
					TaskID:    task.TaskID,
					TaskType:  consts.TaskTypeCollectResult,
					EventName: consts.EventDatapackResultCollection,
					Payload:   results,
				})
			}

			updateTaskStatus(
				childCtx,
				task.TraceID,
				task.TaskID,
				fmt.Sprintf(consts.TaskMsgCompleted, task.TaskID),
				consts.TaskStatusCompleted,
				task.Type,
			)

			logEntry.Info("Collect detector result task completed successfully")

			if hasIssues && repository.CheckCachedField(childCtx, consts.InjectionAlgorithmsKey, task.GroupID) {
				items, err := repository.GetCachedAlgorithmItemsFromRedis(childCtx, consts.InjectionAlgorithmsKey, task.GroupID)
				if err != nil {
					span.AddEvent("failed to get algorithms from redis")
					span.RecordError(err)
					return fmt.Errorf("failed to get algorithms from redis: %v", err)
				}

				for _, item := range items {
					childTask := &dto.UnifiedTask{
						Type: consts.TaskTypeRunAlgorithm,
						Payload: map[string]any{
							consts.ExecuteAlgorithm:    item.Name,
							consts.ExecuteAlgorithmTag: item.Tag,
							consts.ExecuteDataset:      collectPayload.dataset,
							consts.ExecuteEnvVars:      item.EnvVars,
						},
						Immediate:    true,
						TraceID:      task.TraceID,
						GroupID:      task.GroupID,
						ProjectID:    task.ProjectID,
						TraceCarrier: task.TraceCarrier,
					}

					if _, _, err := SubmitTask(childCtx, childTask); err != nil {
						span.AddEvent("failed to submit algorithm execution task")
						span.RecordError(err)
						return fmt.Errorf("failed to submit algorithm execution task: %v", err)
					}
				}

				logEntry.Infof("Algorithm executions submitted successfully")
			}
		} else {
			results, err := repository.ListGranularityResultsByExecutionID(collectPayload.executionID)
			if err != nil {
				span.AddEvent("failed to get detector results by execution ID")
				span.RecordError(err)
				return fmt.Errorf("failed to get detector results by execution ID: %v", err)
			}

			if len(results) == 0 {
				repository.PublishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
					TaskID:    task.TaskID,
					TaskType:  consts.TaskTypeCollectResult,
					EventName: consts.EventAlgoNoResultData,
				})
				span.AddEvent("no granularity results found for the execution ID")
				logEntry.Info("no granularity results found for the execution ID")
				return nil
			}

			repository.PublishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
				TaskID:    task.TaskID,
				TaskType:  consts.TaskTypeCollectResult,
				EventName: consts.EventAlgoResultCollection,
				Payload:   results,
			})

			updateTaskStatus(
				childCtx,
				task.TraceID,
				task.TaskID,
				fmt.Sprintf(consts.TaskMsgCompleted, task.TaskID),
				consts.TaskStatusCompleted,
				task.Type,
			)

			logEntry.Info("Collect algorithm result task completed successfully")
		}

		return nil
	})
}

func parseCollectPayload(payload map[string]any) (*collectionPayload, error) {
	algorithm, err := utils.ConvertToType[dto.AlgorithmItem](payload[consts.CollectAlgorithm])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to AlgorithmItem: %v", consts.CollectAlgorithm, err)
	}

	dataset, ok := payload[consts.CollectDataset].(string)
	if !ok || dataset == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", consts.CollectDataset)
	}

	executionIDFloat, ok := payload[consts.CollectExecutionID].(float64)
	if !ok || executionIDFloat == 0.0 {
		return nil, fmt.Errorf("missing '%s' key in payload", consts.CollectExecutionID)
	}
	executionID := int(executionIDFloat)

	timestamp, ok := payload[consts.CollectTimestamp].(string)
	if !ok || timestamp == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", consts.CollectTimestamp)
	}

	return &collectionPayload{
		algorithm:   algorithm,
		dataset:     dataset,
		executionID: executionID,
		timestamp:   timestamp,
	}, nil
}
