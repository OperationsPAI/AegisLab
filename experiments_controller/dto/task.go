package dto

import (
	"encoding/json"
	"time"

	"github.com/CUHK-SE-Group/rcabench/database"
)

type TaskDetailResp struct {
	Task TaskItem `json:"task"`
	Logs []string `json:"logs"`
}

type TaskItem struct {
	ID        string         `json:"id"`
	TraceID   string         `json:"trace_id"`
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload"`
	Status    string         `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
}

func (t *TaskItem) Convert(task database.Task) error {
	var payload map[string]any
	if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
		return err
	}

	t.ID = task.ID
	t.TraceID = task.TraceID
	t.Type = task.Type
	t.Payload = payload
	t.Status = task.Status
	t.CreatedAt = task.CreatedAt

	return nil
}

type TaskReq struct {
	TaskID string `uri:"task_id" binding:"required"`
}

type TaskStreamItem struct {
	Type    string
	TraceID string
}

func (t *TaskStreamItem) Convert(task database.Task) {
	t.Type = task.Type
	t.TraceID = task.TraceID
}
