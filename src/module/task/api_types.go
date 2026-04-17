package taskmodule

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	"aegis/utils"
)

// BatchDeleteTaskReq represents the request to batch delete tasks.
type BatchDeleteTaskReq struct {
	IDs []string `json:"ids" binding:"required"`
}

func (req *BatchDeleteTaskReq) Validate() error {
	for i, id := range req.IDs {
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("empty id at index %d", i)
		}
		if !utils.IsValidUUID(id) {
			return fmt.Errorf("invalid UUID format for id at index %d: %s", i, id)
		}
	}
	return nil
}

// ListTaskFilters represents the filters for listing tasks.
type ListTaskFilters struct {
	TaskType  *consts.TaskType
	Immediate *bool
	TraceID   string
	GroupID   string
	ProjectID int
	State     *consts.TaskState
	Status    *consts.StatusType
}

// ListTaskReq represents the request to list tasks.
type ListTaskReq struct {
	dto.PaginationReq
	TaskType  *consts.TaskType   `form:"task_type" binding:"omitempty"`
	Immediate *bool              `form:"immediate" binding:"omitempty"`
	TraceID   string             `form:"trace_id" binding:"omitempty"`
	GroupID   string             `form:"group_id" binding:"omitempty"`
	ProjectID int                `form:"project_id" binding:"omitempty"`
	State     *consts.TaskState  `form:"state" binding:"omitempty"`
	Status    *consts.StatusType `form:"status" binding:"omitempty"`
}

func (req *ListTaskReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if err := validateTaskType(req.TaskType); err != nil {
		return err
	}
	if err := validateUUID(req.TraceID); err != nil {
		return err
	}
	if err := validateUUID(req.GroupID); err != nil {
		return err
	}
	if req.ProjectID < 0 {
		return fmt.Errorf("invalid project ID: %d", req.ProjectID)
	}
	if err := validateState(req.State); err != nil {
		return err
	}
	return validateStatus(req.Status)
}

func (req *ListTaskReq) ToFilterOptions() *ListTaskFilters {
	return &ListTaskFilters{
		Immediate: req.Immediate,
		TaskType:  req.TaskType,
		TraceID:   req.TraceID,
		GroupID:   req.GroupID,
		ProjectID: req.ProjectID,
		State:     req.State,
		Status:    req.Status,
	}
}

// TaskResp represents the response for a task.
type TaskResp struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Immediate   bool   `json:"immediate"`
	ExecuteTime int64  `json:"execute_time"`
	CronExpr    string `json:"cron_expr,omitempty"`
	TraceID     string `json:"trace_id"`
	GroupID     string `json:"group_id"`

	State       string    `json:"state"`
	Status      string    `json:"status"`
	ProjectID   int       `json:"project_id,omitempty"`
	ProjectName string    `json:"project_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewTaskResp(task *model.Task) *TaskResp {
	return &TaskResp{
		ID:          task.ID,
		Type:        consts.GetTaskTypeName(task.Type),
		Immediate:   task.Immediate,
		ExecuteTime: task.ExecuteTime,
		CronExpr:    task.CronExpr,
		TraceID:     task.TraceID,
		State:       consts.GetTaskStateName(task.State),
		Status:      consts.GetStatusTypeName(task.Status),
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

// TaskDetailResp represents a task with payload and logs.
type TaskDetailResp struct {
	TaskResp

	Payload map[string]any `json:"payload,omitempty" swaggertype:"object"`
	Logs    []string       `json:"logs"`
}

func NewTaskDetailResp(task *model.Task, logs []string) *TaskDetailResp {
	resp := &TaskDetailResp{
		TaskResp: *NewTaskResp(task),
		Logs:     logs,
	}

	if task.Payload != "" {
		var payload map[string]any
		if err := json.Unmarshal([]byte(task.Payload), &payload); err == nil {
			resp.Payload = payload
		}
	}
	return resp
}

// QueuedTasksResp represents ready and delayed queued tasks.
type QueuedTasksResp struct {
	ReadyTasks   []TaskResp `json:"ready_tasks"`
	DelayedTasks []TaskResp `json:"delayed_tasks"`
}

func validateState(state *consts.TaskState) error {
	if state != nil {
		if _, exists := consts.ValidTaskStates[*state]; !exists {
			return fmt.Errorf("invalid task state: %d", *state)
		}
	}
	return nil
}

func validateTaskType(taskType *consts.TaskType) error {
	if taskType != nil {
		if _, exists := consts.ValidTaskTypes[*taskType]; !exists {
			return fmt.Errorf("invalid task type: %d", *taskType)
		}
	}
	return nil
}

func validateUUID(id string) error {
	if id == "" {
		return nil
	}
	if !utils.IsValidUUID(id) {
		return fmt.Errorf("invalid UUID format: %s", id)
	}
	return nil
}

func validateStatus(status *consts.StatusType) error {
	if status != nil {
		if _, exists := consts.ValidStatuses[*status]; !exists {
			return fmt.Errorf("invalid status value: %d", *status)
		}
	}
	return nil
}
