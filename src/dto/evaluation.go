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
		return fmt.Errorf("One of the following must be provided: pairs, (algorithms and datasets), or execution_ids")
	}

	if modeCount > 1 {
		return fmt.Errorf("Only one query mode can be used at a time: pairs, (algorithms and datasets), or execution_ids")
	}

	if req.HasPairsMode() {
		for i, pair := range req.Pairs {
			if pair.Algorithm == "" {
				return fmt.Errorf("Algorithm cannot be empty in pair at index %d", i)
			}

			if pair.Algorithm == config.GetString("algo.detector") {
				return fmt.Errorf("Algorithm '%s' is reserved and cannot be used in pairs", config.GetString("algo.detector"))
			}

			if pair.Dataset == "" {
				return fmt.Errorf("Dataset cannot be empty in pair at index %d", i)
			}
		}
	}

	if req.HasExecutionMode() {
		for i, id := range req.ExecutionIDs {
			if id <= 0 {
				return fmt.Errorf("Execution ID must be greater than 0 at index %d", i)
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
