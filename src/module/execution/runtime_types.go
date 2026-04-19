package execution

import (
	"aegis/consts"
	"aegis/dto"
)

// RuntimeCreateExecutionReq captures execution writes initiated by runtime-worker-service.
type RuntimeCreateExecutionReq struct {
	TaskID             string          `json:"task_id"`
	AlgorithmVersionID int             `json:"algorithm_version_id"`
	DatapackID         int             `json:"datapack_id"`
	DatasetVersionID   *int            `json:"dataset_version_id,omitempty"`
	Labels             []dto.LabelItem `json:"labels,omitempty"`
}

// RuntimeUpdateExecutionStateReq captures execution state mutations initiated by runtime-worker-service.
type RuntimeUpdateExecutionStateReq struct {
	ExecutionID int                   `json:"execution_id"`
	State       consts.ExecutionState `json:"state"`
}
