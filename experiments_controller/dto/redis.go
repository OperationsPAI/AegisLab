package dto

import "github.com/CUHK-SE-Group/rcabench/consts"

type RdbMsg struct {
	Status string          `json:"status"`
	Error  string          `json:"error"`
	TaskID string          `json:"task_id"`
	Type   consts.TaskType `json:"task_type"`
}
