package producer

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"errors"
	"fmt"
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
