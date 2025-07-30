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
