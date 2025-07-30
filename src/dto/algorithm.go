package dto

import (
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/utils"
)

type AlgorithmItem struct {
	Name  string `json:"name" binding:"required"`
	Image string `json:"image" binding:"omitempty"`
	Tag   string `json:"tag" binding:"omitempty"`
}

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

type ListAlgorithmsResp []database.Container

type DetectorRecord struct {
	SpanName            string   `json:"span_name"`
	Issues              string   `json:"issue"`
	AbnormalAvgDuration *float64 `json:"abnormal_avg_duration" swaggertype:"number" example:"0.5"`
	NormalAvgDuration   *float64 `json:"normal_avg_duration" swaggertype:"number" example:"0.3"`
	AbnormalSuccRate    *float64 `json:"abnormal_succ_rate" swaggertype:"number" example:"0.8"`
	NormalSuccRate      *float64 `json:"normal_succ_rate" swaggertype:"number" example:"0.95"`
	AbnormalP90         *float64 `json:"abnormal_p90" swaggertype:"number" example:"1.2"`
	NormalP90           *float64 `json:"normal_p90" swaggertype:"number" example:"0.8"`
	AbnormalP95         *float64 `json:"abnormal_p95" swaggertype:"number" example:"1.5"`
	NormalP95           *float64 `json:"normal_p95" swaggertype:"number" example:"1.0"`
	AbnormalP99         *float64 `json:"abnormal_p99" swaggertype:"number" example:"2.0"`
	NormalP99           *float64 `json:"normal_p99" swaggertype:"number" example:"1.3"`
}

type ExecutionRecord struct {
	Algorithm          string              `json:"algorithm"`
	GranularityRecords []GranularityRecord `json:"granularity_records"`
}

type ExecutionRecordWithDatasetID struct {
	DatasetID int
	ExecutionRecord
}

type GranularityRecord struct {
	Level      string  `json:"level"`
	Result     string  `json:"result"`
	Rank       int     `json:"rank"`
	Confidence float64 `json:"confidence"`
}

func (g *GranularityRecord) Convert(result database.GranularityResult) {
	g.Level = result.Level
	g.Result = result.Result
	g.Rank = result.Rank
	g.Confidence = result.Confidence
}
