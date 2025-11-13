package producer

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// BatchDeleteTasks deletes multiple tasks by their IDs
func BatchDeleteTasks(taskIDs []string) error {
	if len(taskIDs) == 0 {
		return nil
	}

	if err := repository.BatchDeleteTasks(database.DB, taskIDs); err != nil {
		return err
	}
	return nil
}

// GetTaskDetail retrieves detailed information about a specific task
func GetTaskDetail(taskID string) (*dto.TaskDetailResp, error) {
	task, err := repository.GetTaskByID(database.DB, taskID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: task id: %s", consts.ErrNotFound, taskID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// TODO logs retrieval can be added later
	resp := dto.NewTaskDetailResp(task, []string{})
	return resp, nil
}

// ListTasks lists tasks based on filter options and pagination
func ListTasks(req *dto.ListTaskReq) (*dto.ListResp[dto.TaskResp], error) {
	if req == nil {
		return nil, fmt.Errorf("list tasks request is nil")
	}

	limit, offset := req.ToGormParams()
	fitlerOptions := req.ToFilterOptions()

	tasks, total, err := repository.ListTasks(database.DB, limit, offset, fitlerOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	taskResps := make([]dto.TaskResp, 0, len(tasks))
	for _, task := range tasks {
		taskResps = append(taskResps, *dto.NewTaskResp(&task))
	}

	resp := dto.ListResp[dto.TaskResp]{
		Items:      taskResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

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
