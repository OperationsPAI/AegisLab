package consumer

import (
	"context"
	"fmt"

	"aegis/client"
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/service/common"
	"aegis/tracing"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type collectionPayload struct {
	algorithm   dto.ContainerVersionItem
	datapack    dto.InjectionItem
	executionID int
}

func executeCollectResult(ctx context.Context, task *dto.UnifiedTask) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		logEntry := logrus.WithField("task_id", task.TaskID)
		span := trace.SpanFromContext(childCtx)

		collectPayload, err := parseCollectPayload(task.Payload)
		if err != nil {
			logEntry.Errorf("failed to parse collection payload: %v", err)
			span.AddEvent("failed to parse collection payload")
			span.RecordError(err)
			return err
		}

		stream := fmt.Sprintf(consts.StreamLogKey, task.TraceID)

		if collectPayload.algorithm.ContainerName == config.GetString(consts.DetectorKey) {
			results, err := repository.ListDetectorResultsByExecutionID(database.DB, collectPayload.executionID)
			if err != nil {
				logEntry.Errorf("failed to get detector results by execution ID: %v", err)
				span.AddEvent("failed to get detector results by execution ID")
				span.RecordError(err)
				return fmt.Errorf("failed to get detector results by execution ID: %w", err)
			}

			if len(results) == 0 {
				publishEvent(childCtx, stream, dto.StreamEvent{
					TaskID:    task.TaskID,
					TaskType:  consts.TaskTypeCollectResult,
					EventName: consts.EventDatapackNoDetectorData,
				})

				updateTaskState(
					childCtx,
					task.TraceID,
					task.TaskID,
					fmt.Sprintf(consts.TaskMsgCompleted, task.TaskID),
					consts.TaskCompleted,
					task.Type,
				)

				logEntry.Info("no detector results found for the execution ID")
				span.AddEvent("no detector results found for the execution ID")
				return nil
			}

			hasIssues := false
			for _, v := range results {
				if v.Issues != "{}" {
					hasIssues = true
				}
			}

			if !hasIssues {
				publishEvent(childCtx, stream, dto.StreamEvent{
					TaskID:    task.TaskID,
					TaskType:  consts.TaskTypeCollectResult,
					EventName: consts.EventDatapackNoAnomaly,
					Payload:   results,
				})

				span.AddEvent("the detector result has no issues")
				logEntry.Info("the detector result has no issues")
			} else {
				publishEvent(childCtx, stream, dto.StreamEvent{
					TaskID:    task.TaskID,
					TaskType:  consts.TaskTypeCollectResult,
					EventName: consts.EventDatapackResultCollection,
					Payload:   results,
				})
			}

			updateTaskState(
				childCtx,
				task.TraceID,
				task.TaskID,
				fmt.Sprintf(consts.TaskMsgCompleted, task.TaskID),
				consts.TaskCompleted,
				task.Type,
			)

			logEntry.Info("Collect detector result task completed successfully")

			if hasIssues && client.CheckCachedField(childCtx, consts.InjectionAlgorithmsKey, task.GroupID) {
				var algorithms []dto.ContainerVersionItem
				err := client.GetHashField(childCtx, consts.InjectionAlgorithmsKey, task.GroupID, &algorithms)
				if err != nil {
					span.AddEvent("failed to get algorithms from redis")
					span.RecordError(err)
					return fmt.Errorf("failed to get algorithms from redis: %w", err)
				}

				for _, algorithm := range algorithms {
					payload := map[string]any{
						consts.ExecuteAlgorithm: algorithm,
						consts.ExecuteDatapack:  collectPayload.datapack,
					}

					if err := produceAlgorithmExeuctionTask(childCtx, task, payload); err != nil {
						span.AddEvent("failed to submit algorithm execution task")
						span.RecordError(err)
						return fmt.Errorf("failed to submit algorithm execution task: %w", err)
					}
				}

				logEntry.Info("Algorithm executions tasks submitted successfully")
			}

			return nil
		}

		results, err := repository.ListGranularityResultsByExecutionID(database.DB, collectPayload.executionID)
		if err != nil {
			span.AddEvent("failed to get detector results by execution ID")
			span.RecordError(err)
			return fmt.Errorf("failed to get detector results by execution ID: %w", err)
		}

		if len(results) == 0 {
			publishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
				TaskID:    task.TaskID,
				TaskType:  consts.TaskTypeCollectResult,
				EventName: consts.EventAlgoNoResultData,
			})
			span.AddEvent("no granularity results found for the execution ID")
			logEntry.Info("no granularity results found for the execution ID")
			return nil
		}

		publishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  consts.TaskTypeCollectResult,
			EventName: consts.EventAlgoResultCollection,
			Payload:   results,
		})

		updateTaskState(
			childCtx,
			task.TraceID,
			task.TaskID,
			fmt.Sprintf(consts.TaskMsgCompleted, task.TaskID),
			consts.TaskCompleted,
			task.Type,
		)

		logEntry.Info("Collect algorithm result task completed successfully")
		return nil
	})
}

// parseCollectPayload parses the payload for collect result tasks
func parseCollectPayload(payload map[string]any) (*collectionPayload, error) {
	algorithm, err := utils.ConvertToType[dto.ContainerVersionItem](payload[consts.CollectAlgorithm])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to ContainerVersionItem: %w", consts.CollectAlgorithm, err)
	}

	datapack, err := utils.ConvertToType[dto.InjectionItem](payload[consts.CollectDatapack])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to InjectionItem: %w", consts.CollectDatapack, err)
	}

	executionIDFloat, ok := payload[consts.CollectExecutionID].(float64)
	if !ok || executionIDFloat <= consts.DefaultInvalidID {
		return nil, fmt.Errorf("missing or invalid '%s' in collection payload: %w", consts.CollectExecutionID, err)
	}
	executionID := int(executionIDFloat)

	return &collectionPayload{
		algorithm:   algorithm,
		datapack:    datapack,
		executionID: executionID,
	}, nil
}

// produceAlgorithmExeuctionTask produces an algorithm execution task into Redis
func produceAlgorithmExeuctionTask(ctx context.Context, task *dto.UnifiedTask, payload map[string]any) error {
	newTask := &dto.UnifiedTask{
		Type:         consts.TaskTypeRunAlgorithm,
		Immediate:    true,
		Payload:      payload,
		TraceID:      task.TraceID,
		GroupID:      task.GroupID,
		ProjectID:    task.ProjectID,
		UserID:       task.UserID,
		State:        consts.TaskPending,
		TraceCarrier: task.TraceCarrier,
	}
	err := common.SubmitTask(ctx, newTask)
	if err != nil {
		return fmt.Errorf("failed to submit algorithm exectuion task: %w", err)
	}
	return nil
}
