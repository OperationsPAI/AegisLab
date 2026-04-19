package consumer

import (
	"context"
	"fmt"

	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	redis "aegis/infra/redis"
	execution "aegis/module/execution"
	"aegis/service/common"
	"aegis/tracing"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type collectionPayload struct {
	algorithm   dto.ContainerVersionItem
	datapack    dto.InjectionItem
	executionID int
}

func executeCollectResult(ctx context.Context, task *dto.UnifiedTask, deps RuntimeDeps) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		db := deps.DB
		if db == nil {
			return fmt.Errorf("consumer runtime db is nil")
		}
		redisGateway := deps.RedisGateway
		if redisGateway == nil {
			return fmt.Errorf("consumer redis gateway is nil")
		}

		logEntry := logrus.WithField("task_id", task.TaskID)
		span := trace.SpanFromContext(childCtx)

		collectPayload, err := parseCollectPayload(task.Payload)
		if err != nil {
			logEntry.Errorf("failed to parse collection payload: %v", err)
			span.AddEvent("failed to parse collection payload")
			span.RecordError(err)
			return err
		}

		if collectPayload.algorithm.ContainerName == config.GetDetectorName() {
			results, err := loadDetectorResults(childCtx, deps, db, collectPayload.executionID)
			if err != nil {
				logEntry.Errorf("failed to get detector results by execution ID: %v", err)
				span.AddEvent("failed to get detector results by execution ID")
				span.RecordError(err)
				return fmt.Errorf("failed to get detector results by execution ID: %w", err)
			}

			hasResults := len(results) > 0 // Fixed: should be > 0, not >= 0

			eventName := consts.EventDatapackResultCollection
			if !hasResults {
				eventName = consts.EventDatapackNoDetectorData
				message := fmt.Sprintf("no detector results found for the execution ID %d", collectPayload.executionID)
				logEntry.Warn(message)
				span.AddEvent(message)
			}

			hasIssues := false
			if hasResults {
				for _, v := range results {
					if v.Issues != "{}" {
						hasIssues = true
					}
				}
			}

			if !hasIssues {
				eventName = consts.EventDatapackNoAnomaly
				message := "the detector result has no issues"
				logEntry.Warn(message)
				span.AddEvent(message)
			}

			updateTaskState(childCtx, taskCompletedWithEvent(task, eventName, results).withDB(db).withRedis(redisGateway))

			logEntry.Info("Collect detector result task completed successfully")

			if hasIssues {
				algorithms, cached, err := loadCachedInjectionAlgorithms(redisGateway, childCtx, task.GroupID)
				if err != nil {
					span.AddEvent("failed to get algorithms from redis")
					span.RecordError(err)
					return fmt.Errorf("failed to get algorithms from redis: %w", err)
				}
				if !cached {
					return nil
				}

				for idx, algorithm := range algorithms {
					payload := map[string]any{
						consts.ExecuteAlgorithm: algorithm,
						consts.ExecuteDatapack:  collectPayload.datapack,
					}

					if err := produceAlgorithmExeuctionTask(childCtx, db, deps.RedisGateway, task, payload, idx); err != nil {
						span.AddEvent("failed to submit algorithm execution task")
						span.RecordError(err)
						return fmt.Errorf("failed to submit algorithm execution task: %w", err)
					}
				}

				logEntry.Info("Algorithm executions tasks submitted successfully")
			}

			return nil
		}

		results, err := loadGranularityResults(childCtx, deps, db, collectPayload.executionID)
		if err != nil {
			span.AddEvent("failed to get detector results by execution ID")
			span.RecordError(err)
			return fmt.Errorf("failed to get detector results by execution ID: %w", err)
		}

		eventName := consts.EventAlgoResultCollection
		if len(results) == 0 {
			eventName = consts.EventAlgoNoResultData
			message := fmt.Sprintf("no granularity results found for the execution ID %d", collectPayload.executionID)
			logEntry.Warn(message)
			span.AddEvent(message)
		}

		updateTaskState(childCtx, taskCompletedWithEvent(task, eventName, results).withDB(db).withRedis(redisGateway))

		logEntry.Info("Collect algorithm result task completed successfully")
		return nil
	})
}

func loadDetectorResults(ctx context.Context, deps RuntimeDeps, _ *gorm.DB, executionID int) ([]execution.DetectorResultItem, error) {
	if deps.ExecutionOwner == nil {
		return nil, fmt.Errorf("execution owner service is nil")
	}
	resp, err := deps.ExecutionOwner.GetExecution(ctx, executionID)
	if err != nil {
		return nil, err
	}
	return resp.DetectorResults, nil
}

func loadGranularityResults(ctx context.Context, deps RuntimeDeps, _ *gorm.DB, executionID int) ([]execution.GranularityResultItem, error) {
	if deps.ExecutionOwner == nil {
		return nil, fmt.Errorf("execution owner service is nil")
	}
	resp, err := deps.ExecutionOwner.GetExecution(ctx, executionID)
	if err != nil {
		return nil, err
	}
	return resp.GranularityResults, nil
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
func produceAlgorithmExeuctionTask(ctx context.Context, db *gorm.DB, redisGateway *redis.Gateway, task *dto.UnifiedTask, payload map[string]any, index int) error {
	newTask := &dto.UnifiedTask{
		Type:         consts.TaskTypeRunAlgorithm,
		Immediate:    true,
		Payload:      payload,
		Sequence:     index,
		ParentTaskID: utils.StringPtr(task.TaskID),
		TraceID:      task.TraceID,
		GroupID:      task.GroupID,
		ProjectID:    task.ProjectID,
		UserID:       task.UserID,
		State:        consts.TaskPending,
		TraceCarrier: task.TraceCarrier,
	}
	err := common.SubmitTaskWithDB(ctx, db, redisGateway, newTask)
	if err != nil {
		return fmt.Errorf("failed to submit algorithm exectuion task: %w", err)
	}
	return nil
}
