package metric

import (
	"fmt"
	"time"
)

// GetMetricsReq represents the request to get metrics with time range and filters.
type GetMetricsReq struct {
	StartTime   *time.Time `form:"start_time" binding:"omitempty"`
	EndTime     *time.Time `form:"end_time" binding:"omitempty"`
	FaultType   *string    `form:"fault_type" binding:"omitempty"`
	AlgorithmID *int       `form:"algorithm_id" binding:"omitempty"`
}

func (req *GetMetricsReq) Validate() error {
	if req.StartTime != nil && req.EndTime != nil && req.EndTime.Before(*req.StartTime) {
		return fmt.Errorf("end_time must be after start_time")
	}
	if req.AlgorithmID != nil && *req.AlgorithmID <= 0 {
		return fmt.Errorf("algorithm_id must be positive")
	}
	return nil
}

// InjectionMetrics represents aggregated metrics for injections.
type InjectionMetrics struct {
	TotalCount       int            `json:"total_count"`
	SuccessCount     int            `json:"success_count"`
	FailedCount      int            `json:"failed_count"`
	SuccessRate      float64        `json:"success_rate"`
	AvgDuration      float64        `json:"avg_duration"`
	MinDuration      float64        `json:"min_duration"`
	MaxDuration      float64        `json:"max_duration"`
	StateDistrib     map[string]int `json:"state_distribution" swaggertype:"object"`
	FaultTypeDistrib map[string]int `json:"fault_type_distribution" swaggertype:"object"`
}

// ExecutionMetrics represents aggregated metrics for algorithm executions.
type ExecutionMetrics struct {
	TotalCount   int            `json:"total_count"`
	SuccessCount int            `json:"success_count"`
	FailedCount  int            `json:"failed_count"`
	SuccessRate  float64        `json:"success_rate"`
	AvgDuration  float64        `json:"avg_duration"`
	MinDuration  float64        `json:"min_duration"`
	MaxDuration  float64        `json:"max_duration"`
	StateDistrib map[string]int `json:"state_distribution" swaggertype:"object"`
}

// AlgorithmMetrics represents comparative metrics across different algorithms.
type AlgorithmMetrics struct {
	Algorithms []AlgorithmMetricItem `json:"algorithms"`
}

// AlgorithmMetricItem represents metrics for a single algorithm.
type AlgorithmMetricItem struct {
	AlgorithmID    int     `json:"algorithm_id"`
	AlgorithmName  string  `json:"algorithm_name"`
	ExecutionCount int     `json:"execution_count"`
	SuccessCount   int     `json:"success_count"`
	FailedCount    int     `json:"failed_count"`
	SuccessRate    float64 `json:"success_rate"`
	AvgDuration    float64 `json:"avg_duration"`
}
