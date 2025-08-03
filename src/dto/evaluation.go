package dto

import (
	"fmt"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/LGU-SE-Internal/rcabench/config"
)

// GroundTruthReq represents ground truth request
type GroundTruthReq struct {
	Datasets []string `json:"datasets" binding:"required"`
}

// GroundTruthResp represents ground truth response
type GroundTruthResp map[string]chaos.Groundtruth

// AlgorithmDatasetPair represents algorithm and dataset pair
type AlgorithmDatasetPair struct {
	Algorithm string
	Dataset   string
}

// RawDataReq represents raw data request
type RawDataReq struct {
	Pairs        []AlgorithmDatasetPair `json:"pairs" binding:"omitempty"`
	ExecutionIDs []int                  `json:"execution_ids" binding:"omitempty"`

	TimeRangeQuery
}

func (req *RawDataReq) HasPairsMode() bool {
	return len(req.Pairs) > 0
}

func (req *RawDataReq) HasExecutionMode() bool {
	return len(req.ExecutionIDs) > 0
}

func (req *RawDataReq) Validate() error {
	modeCount := 0
	if req.HasPairsMode() {
		modeCount++
	}
	if req.HasExecutionMode() {
		modeCount++
	}

	if modeCount == 0 {
		return fmt.Errorf("one of the following must be provided: pairs, (algorithms and datasets), or execution_ids")
	}

	if modeCount > 1 {
		return fmt.Errorf("only one query mode can be used at a time: pairs, (algorithms and datasets), or execution_ids")
	}

	if req.HasPairsMode() {
		for i, pair := range req.Pairs {
			if pair.Algorithm == "" {
				return fmt.Errorf("algorithm cannot be empty in pair at index %d", i)
			}

			if pair.Algorithm == config.GetString("algo.detector") {
				return fmt.Errorf("algorithm '%s' is reserved and cannot be used in pairs", config.GetString("algo.detector"))
			}

			if pair.Dataset == "" {
				return fmt.Errorf("dataset cannot be empty in pair at index %d", i)
			}
		}
	}

	if req.HasExecutionMode() {
		for i, id := range req.ExecutionIDs {
			if id <= 0 {
				return fmt.Errorf("execution ID must be greater than 0 at index %d", i)
			}
		}
	}

	return req.TimeRangeQuery.Validate()
}

// RawDataItem represents raw data item
type RawDataItem struct {
	Algorithm   string              `json:"algorithm"`
	Dataset     string              `json:"dataset"`
	ExecutionID int                 `json:"execution_id,omitempty"`
	Groundtruth chaos.Groundtruth   `json:"groundtruth"`
	Entries     []GranularityRecord `json:"entries,omitempty"`
}

// RawDataResp represents raw data response
type RawDataResp []RawDataItem

// Execution represents execution data for evaluation
type Execution struct {
	Dataset            DatasetItem         `json:"dataset"`
	GranularityRecords []GranularityRecord `json:"granularity_records"`
}

// Conclusion represents evaluation conclusion
type Conclusion struct {
	Level  string  `json:"level"`  // For example service level
	Metric string  `json:"metric"` // For example topk
	Rate   float64 `json:"rate"`
}

// EvaluateMetric represents evaluation metric function type
type EvaluateMetric func([]Execution) ([]Conclusion, error)

// AlgorithmDatasetEvaluationReq represents request for algorithm evaluation on a dataset
type AlgorithmDatasetEvaluationReq struct {
	Algorithm      string `json:"algorithm" binding:"required"`
	Dataset        string `json:"dataset" binding:"required"`
	DatasetVersion string `json:"dataset_version,omitempty" form:"dataset_version"` // Dataset version (optional, defaults to "v1.0")
	Tag            string `json:"tag,omitempty" form:"tag"`                         // Tag filter for filtering execution results
}

// DatapackEvaluationItem represents evaluation item for a single datapack
type DatapackEvaluationItem struct {
	DatapackName string              `json:"datapack_name"` // Datapack name (from FaultInjectionSchedule)
	ExecutionID  int                 `json:"execution_id"`  // Execution ID
	Groundtruth  chaos.Groundtruth   `json:"groundtruth"`   // Ground truth for this datapack
	Predictions  []GranularityRecord `json:"predictions"`   // Algorithm predictions
	ExecutedAt   string              `json:"executed_at"`   // Execution time
}

// AlgorithmDatasetEvaluationResp represents response for algorithm evaluation on a dataset
type AlgorithmDatasetEvaluationResp struct {
	Algorithm      string                   `json:"algorithm"`       // Algorithm name
	Dataset        string                   `json:"dataset"`         // Dataset name
	DatasetVersion string                   `json:"dataset_version"` // Dataset version
	TotalCount     int                      `json:"total_count"`     // Total number of datapacks in dataset
	ExecutedCount  int                      `json:"executed_count"`  // Number of successfully executed datapacks
	Items          []DatapackEvaluationItem `json:"items"`           // Evaluation items for each datapack
}

// AlgorithmDatapackEvaluationReq represents request for algorithm evaluation on a single datapack
type AlgorithmDatapackEvaluationReq struct {
	Algorithm string `json:"algorithm" binding:"required"`
	Datapack  string `json:"datapack" binding:"required"`
	Tag       string `json:"tag,omitempty" form:"tag"` // Tag filter for filtering execution results
}

// AlgorithmDatapackEvaluationResp represents response for algorithm evaluation on a single datapack
type AlgorithmDatapackEvaluationResp struct {
	Algorithm   string              `json:"algorithm"`    // Algorithm name
	Datapack    string              `json:"datapack"`     // Datapack name
	ExecutionID int                 `json:"execution_id"` // Execution ID (0 if no execution found)
	Groundtruth chaos.Groundtruth   `json:"groundtruth"`  // Ground truth for this datapack
	Predictions []GranularityRecord `json:"predictions"`  // Algorithm predictions
	ExecutedAt  string              `json:"executed_at"`  // Execution time
	Found       bool                `json:"found"`        // Whether execution result was found
}
