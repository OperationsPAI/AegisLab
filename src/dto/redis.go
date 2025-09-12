package dto

import "github.com/LGU-SE-Internal/rcabench/consts"

type RdbMsg struct {
	Status string          `json:"status"`
	Error  string          `json:"error"`
	TaskID string          `json:"task_id"`
	Type   consts.TaskType `json:"task_type"`
}
