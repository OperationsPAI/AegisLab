package dto

import (
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/utils"
)

// AlgorithmItem represents an algorithm configuration
type AlgorithmItem struct {
	Name  string `json:"name" binding:"required"`
	Image string `json:"image" binding:"omitempty"`
	Tag   string `json:"tag" binding:"omitempty"`
}

// ExecutionPayload represents algorithm execution payload
type ExecutionPayload struct {
	Algorithm AlgorithmItem     `json:"algorithm" binding:"required"`
	Dataset   string            `json:"dataset" binding:"required"`
	EnvVars   map[string]string `json:"env_vars" binding:"omitempty" swaggertype:"object"`
}

func (p *ExecutionPayload) Validate() error {
	for key := range p.EnvVars {
		if err := utils.IsValidEnvVar(key); err != nil {
			return fmt.Errorf("invalid environment variable key %s: %v", key, err)
		}
	}

	return nil
}

// SubmitExecutionReq represents algorithm execution submission request
type SubmitExecutionReq struct {
	ProjectName string             `json:"project_name" binding:"required"`
	Payloads    []ExecutionPayload `json:"payloads" binding:"required,dive,required"`
}

func (req *SubmitExecutionReq) Validate() error {
	if req.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}
	if len(req.Payloads) == 0 {
		return fmt.Errorf("at least one execution payload is required")
	}
	for _, payload := range req.Payloads {
		if err := payload.Validate(); err != nil {
			return fmt.Errorf("invalid execution payload: %v", err)
		}
	}
	return nil
}

// ListAlgorithmsResp represents algorithm list response
type ListAlgorithmsResp []database.Container

// DatasetExecutionPayload 表示使用数据集执行算法的负载
type DatasetExecutionPayload struct {
	Algorithm AlgorithmItem     `json:"algorithm" binding:"required"`
	Dataset   string            `json:"dataset" binding:"required"`
	EnvVars   map[string]string `json:"env_vars" binding:"omitempty" swaggertype:"object"`
}

func (p *DatasetExecutionPayload) Validate() error {
	// 检查环境变量
	for key := range p.EnvVars {
		if err := utils.IsValidEnvVar(key); err != nil {
			return fmt.Errorf("invalid environment variable key %s: %v", key, err)
		}
	}

	if p.Dataset == "" {
		return fmt.Errorf("dataset is required")
	}

	return nil
}

// DatapackExecutionPayload 表示使用数据包执行算法的负载
type DatapackExecutionPayload struct {
	Algorithm AlgorithmItem     `json:"algorithm" binding:"required"`
	Datapack  string            `json:"datapack" binding:"required"`
	EnvVars   map[string]string `json:"env_vars" binding:"omitempty" swaggertype:"object"`
}

func (p *DatapackExecutionPayload) Validate() error {
	// 检查环境变量
	for key := range p.EnvVars {
		if err := utils.IsValidEnvVar(key); err != nil {
			return fmt.Errorf("invalid environment variable key %s: %v", key, err)
		}
	}

	if p.Datapack == "" {
		return fmt.Errorf("datapack is required")
	}

	return nil
}

// SubmitDatasetExecutionReq 表示使用数据集执行算法的请求
type SubmitDatasetExecutionReq struct {
	ProjectName string                  `json:"project_name" binding:"required"`
	Payload     DatasetExecutionPayload `json:"payload" binding:"required"`
}

func (req *SubmitDatasetExecutionReq) Validate() error {
	if req.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}

	if err := req.Payload.Validate(); err != nil {
		return fmt.Errorf("invalid execution payload: %v", err)
	}

	return nil
}

// SubmitDatapackExecutionReq 表示使用数据包执行算法的请求
type SubmitDatapackExecutionReq struct {
	ProjectName string                   `json:"project_name" binding:"required"`
	Payload     DatapackExecutionPayload `json:"payload" binding:"required"`
}

func (req *SubmitDatapackExecutionReq) Validate() error {
	if req.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}

	if err := req.Payload.Validate(); err != nil {
		return fmt.Errorf("invalid execution payload: %v", err)
	}

	return nil
}

// SubmitBatchDatasetExecutionReq 表示批量使用数据集执行算法的请求
type SubmitBatchDatasetExecutionReq struct {
	ProjectName string                    `json:"project_name" binding:"required"`
	Payloads    []DatasetExecutionPayload `json:"payloads" binding:"required,dive,required"`
}

func (req *SubmitBatchDatasetExecutionReq) Validate() error {
	if req.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}

	if len(req.Payloads) == 0 {
		return fmt.Errorf("at least one execution payload is required")
	}

	for _, payload := range req.Payloads {
		if err := payload.Validate(); err != nil {
			return fmt.Errorf("invalid execution payload: %v", err)
		}
	}

	return nil
}

// SubmitBatchDatapackExecutionReq 表示批量使用数据包执行算法的请求
type SubmitBatchDatapackExecutionReq struct {
	ProjectName string                     `json:"project_name" binding:"required"`
	Payloads    []DatapackExecutionPayload `json:"payloads" binding:"required,dive,required"`
}

func (req *SubmitBatchDatapackExecutionReq) Validate() error {
	if req.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}

	if len(req.Payloads) == 0 {
		return fmt.Errorf("at least one execution payload is required")
	}

	for _, payload := range req.Payloads {
		if err := payload.Validate(); err != nil {
			return fmt.Errorf("invalid execution payload: %v", err)
		}
	}

	return nil
}

// V2 Algorithm Execution DTOs

// AlgorithmExecutionRequest represents v2 algorithm execution request
type AlgorithmExecutionRequest struct {
	ProjectName string            `json:"project_name" binding:"required"`
	Algorithm   AlgorithmItem     `json:"algorithm" binding:"required"`
	EnvVars     map[string]string `json:"env_vars" binding:"omitempty" swaggertype:"object"`
	Datapack    *string           `json:"datapack,omitempty"`
	Dataset     *string           `json:"dataset,omitempty"`
}

func (req *AlgorithmExecutionRequest) Validate() error {
	if req.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}

	// Must specify either datapack or dataset, but not both
	if req.Datapack == nil && req.Dataset == nil {
		return fmt.Errorf("either datapack or dataset must be specified")
	}

	if req.Datapack != nil && req.Dataset != nil {
		return fmt.Errorf("cannot specify both datapack and dataset")
	}

	if req.Datapack != nil && *req.Datapack == "" {
		return fmt.Errorf("datapack name cannot be empty")
	}

	if req.Dataset != nil && *req.Dataset == "" {
		return fmt.Errorf("dataset name cannot be empty")
	}

	// Validate environment variables
	for key := range req.EnvVars {
		if err := utils.IsValidEnvVar(key); err != nil {
			return fmt.Errorf("invalid environment variable key %s: %v", key, err)
		}
	}

	return nil
}

// BatchAlgorithmExecutionRequest represents v2 batch algorithm execution request
type BatchAlgorithmExecutionRequest struct {
	ProjectName string                      `json:"project_name" binding:"required"`
	Executions  []AlgorithmExecutionRequest `json:"executions" binding:"required,dive,required"`
}

func (req *BatchAlgorithmExecutionRequest) Validate() error {
	if req.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}

	if len(req.Executions) == 0 {
		return fmt.Errorf("at least one execution is required")
	}

	for i, execution := range req.Executions {
		// Override project name to ensure consistency
		execution.ProjectName = req.ProjectName

		if err := execution.Validate(); err != nil {
			return fmt.Errorf("invalid execution at index %d: %v", i, err)
		}
	}

	return nil
}

// AlgorithmExecutionResponse represents algorithm execution response
type AlgorithmExecutionResponse struct {
	TraceID     string `json:"trace_id"`
	TaskID      string `json:"task_id"`
	AlgorithmID int    `json:"algorithm_id"`
	DatapackID  *int   `json:"datapack_id,omitempty"`
	DatasetID   *int   `json:"dataset_id,omitempty"`
	Status      string `json:"status"`
}

// BatchAlgorithmExecutionResponse represents batch algorithm execution response
type BatchAlgorithmExecutionResponse struct {
	GroupID    string                       `json:"group_id"`
	Executions []AlgorithmExecutionResponse `json:"executions"`
	Message    string                       `json:"message"`
}
