package dto

import (
	"aegis/config"
	"aegis/database"
	"fmt"
	"time"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
)

// =====================================================================
// Evaluation CRUD DTOs
// =====================================================================

// ListEvaluationReq represents the request for listing evaluations
type ListEvaluationReq struct {
	PaginationReq
}

// EvaluationResp represents an evaluation in API responses
type EvaluationResp struct {
	ID               int       `json:"id"`
	ProjectID        *int      `json:"project_id,omitempty"`
	AlgorithmName    string    `json:"algorithm_name"`
	AlgorithmVersion string    `json:"algorithm_version"`
	DatapackName     string    `json:"datapack_name,omitempty"`
	DatasetName      string    `json:"dataset_name,omitempty"`
	DatasetVersion   string    `json:"dataset_version,omitempty"`
	EvalType         string    `json:"eval_type"`
	Precision        float64   `json:"precision"`
	Recall           float64   `json:"recall"`
	F1Score          float64   `json:"f1_score"`
	Accuracy         float64   `json:"accuracy"`
	ResultJSON       string    `json:"result_json,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// NewEvaluationResp creates an EvaluationResp from a database Evaluation
func NewEvaluationResp(eval *database.Evaluation) *EvaluationResp {
	return &EvaluationResp{
		ID:               eval.ID,
		ProjectID:        eval.ProjectID,
		AlgorithmName:    eval.AlgorithmName,
		AlgorithmVersion: eval.AlgorithmVersion,
		DatapackName:     eval.DatapackName,
		DatasetName:      eval.DatasetName,
		DatasetVersion:   eval.DatasetVersion,
		EvalType:         eval.EvalType,
		Precision:        eval.Precision,
		Recall:           eval.Recall,
		F1Score:          eval.F1Score,
		Accuracy:         eval.Accuracy,
		ResultJSON:       eval.ResultJSON,
		CreatedAt:        eval.CreatedAt,
		UpdatedAt:        eval.UpdatedAt,
	}
}

// Execution represents execution data for evaluation
type Execution struct {
	Items []GranularityResultItem `json:"items"`
}

// Conclusion represents evaluation conclusion
type Conclusion struct {
	Level  string  `json:"level"`  // For example service level
	Metric string  `json:"metric"` // For example topk
	Rate   float64 `json:"rate"`
}

// EvaluateMetric represents evaluation metric function type
type EvaluateMetric func([]Execution) ([]Conclusion, error)

// =====================================================================
// Batch Evaluate Datapack DTOs
// =====================================================================

type EvaluateDatapackSpec struct {
	Algorithm    ContainerRef `json:"algorithm" binding:"required"`
	Datapack     string       `json:"datapack" binding:"required"`
	FilterLabels []LabelItem  `json:"filter_labels" binding:"omitempty"`
}

func (spec *EvaluateDatapackSpec) Validate() error {
	if err := spec.Algorithm.Validate(); err != nil {
		return fmt.Errorf("invalid algorithm: %w", err)
	}
	if spec.Algorithm.Name == config.GetDetectorName() {
		return fmt.Errorf("detector algorithm cannot be used for evaluation")
	}

	if spec.Datapack == "" {
		return fmt.Errorf("datapack cannot be empty")
	}

	return validateLabelItemsFiled(spec.FilterLabels)
}

type BatchEvaluateDatapackReq struct {
	Specs []EvaluateDatapackSpec `json:"specs" binding:"required"`
}

func (req *BatchEvaluateDatapackReq) Validate() error {
	if len(req.Specs) == 0 {
		return fmt.Errorf("at least one evaluation spec is required")
	}
	for i, spec := range req.Specs {
		if err := spec.Validate(); err != nil {
			return fmt.Errorf("invalid spec at index %d: %w", i, err)
		}
	}
	return nil
}

type EvaluateDatapackRef struct {
	Datapack      string              `json:"datapack"`
	Groundtruths  []chaos.Groundtruth `json:"groundtruths"`
	ExecutionRefs []ExecutionRef      `json:"execution_refs"`
}

type EvaluateDatapackItem struct {
	Algorithm        string `json:"algorithm"`
	AlgorithmVersion string `json:"algorithm_version"`
	EvaluateDatapackRef
}

type BatchEvaluateDatapackResp struct {
	FailedCount  int                    `json:"failed_count"`
	FailedItems  []string               `json:"failed_items"`
	SuccessCount int                    `json:"success_count"`
	SuccessItems []EvaluateDatapackItem `json:"success_items"`
}

// =====================================================================
// Batch Evaluate Dataset DTOs
// =====================================================================

type EvaluateDatasetSpec struct {
	Algorithm    ContainerRef `json:"algorithm" binding:"required"`
	Dataset      DatasetRef   `json:"dataset" binding:"required"`
	FilterLabels []LabelItem  `json:"filter_labels" binding:"omitempty"`
}

func (spec *EvaluateDatasetSpec) Validate() error {
	if err := spec.Algorithm.Validate(); err != nil {
		return fmt.Errorf("invalid algorithm: %w", err)
	}
	if spec.Algorithm.Name == config.GetDetectorName() {
		return fmt.Errorf("detector algorithm cannot be used for evaluation")
	}

	if err := spec.Dataset.Validate(); err != nil {
		return fmt.Errorf("invalid dataset: %w", err)
	}

	return validateLabelItemsFiled(spec.FilterLabels)
}

type BatchEvaluateDatasetReq struct {
	Specs []EvaluateDatasetSpec `json:"specs" binding:"required"`
}

func (req *BatchEvaluateDatasetReq) Validate() error {
	if len(req.Specs) == 0 {
		return fmt.Errorf("at least one evaluation spec is required")
	}
	for i, spec := range req.Specs {
		if err := spec.Validate(); err != nil {
			return fmt.Errorf("invalid spec at index %d: %w", i, err)
		}
	}
	return nil
}

type EvaluateDatasetItem struct {
	Algorithm            string                `json:"algorithm"`              // Algorithm name
	AlgorithmVersion     string                `json:"algorithm_version"`      // Algorithm version
	Dataset              string                `json:"dataset"`                // Dataset name
	DatasetVersion       string                `json:"dataset_version"`        // Dataset version
	TotalCount           int                   `json:"total_count"`            // Total number of datapacks in dataset
	EvaluateRefs         []EvaluateDatapackRef `json:"evalaute_refs"`          // Evaluation refs for each dataset
	NotExecutedDatapacks []string              `json:"not_executed_datapacks"` // Datapacks that were not executed
}

type BatchEvaluateDatasetResp struct {
	FailedCount  int                   `json:"failed_count"`
	FailedItems  []string              `json:"failed_items"`
	SuccessCount int                   `json:"success_count"`
	SuccessItems []EvaluateDatasetItem `json:"success_items"`
}
