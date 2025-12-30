package producer

import (
	"aegis/client"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// InspectLock retrieves the current lock status of all namespaces
func InspectLock(ctx context.Context) (*dto.ListNamespaceLockResp, error) {
	redisClient := client.GetRedisClient()

	// Get all namespaces
	namespaces, err := redisClient.SMembers(ctx, consts.NamespacesKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces from Redis: %v", err)
	}

	nsMap := make(map[string]dto.NsMonitorItem, len(namespaces))

	// Get data for each namespace
	for _, ns := range namespaces {
		nsKey := fmt.Sprintf(consts.NamespaceKeyPattern, ns)
		values, err := redisClient.HGetAll(ctx, nsKey).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get data for namespace %s: %v", ns, err)
		}

		endTimeUnix, err := strconv.ParseInt(values["end_time"], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time format for namespace %s: %v", ns, err)
		}

		// Get status, default to enabled for backward compatibility
		status := consts.CommonEnabled
		if statusStr, ok := values["status"]; ok {
			statusInt, err := strconv.Atoi(statusStr)
			if err == nil {
				status = consts.StatusType(statusInt)
			}
		}

		nsMap[ns] = dto.NsMonitorItem{
			LockedBy: values["trace_id"],
			EndTime:  time.Unix(endTimeUnix, 0),
			Status:   consts.GetStatusTypeName(status),
		}
	}

	resp := &dto.ListNamespaceLockResp{
		Items: nsMap,
	}
	return resp, nil
}

// ListQueuedTasks lists tasks currently in the ready and delayed queues
func ListQueuedTasks(ctx context.Context) (*dto.QueuedTasksResp, error) {
	readyTaskDatas, err := repository.ListReadyTasks(ctx)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("%w: no ready tasks found", consts.ErrNotFound)
		}
		return nil, err
	}

	readyTask := make([]dto.TaskResp, 0, len(readyTaskDatas))
	for _, taskData := range readyTaskDatas {
		var task database.Task
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			return nil, err
		}

		readyTask = append(readyTask, *dto.NewTaskResp(&task))
	}

	delayedTaskDatas, err := repository.ListDelayedTasks(ctx, 1000)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("%w: no delayed tasks found", consts.ErrNotFound)
		}
		return nil, err
	}

	delayedTask := make([]dto.TaskResp, 0, len(delayedTaskDatas))
	for _, taskData := range delayedTaskDatas {
		var task database.Task
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			return nil, err
		}

		delayedTask = append(delayedTask, *dto.NewTaskResp(&task))
	}

	resp := &dto.QueuedTasksResp{
		ReadyTasks:   readyTask,
		DelayedTasks: delayedTask,
	}
	return resp, nil
}
