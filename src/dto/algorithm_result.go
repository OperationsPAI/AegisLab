package dto

import (
	"aegis/database"
	"fmt"
	"time"
)

// DetectorResultItem Single detector result item
type DetectorResultItem struct {
	SpanName            string   `json:"span_name" binding:"required"`
	Issues              string   `json:"issues" binding:"required"`
	AbnormalAvgDuration *float64 `json:"abnormal_avg_duration,omitempty"`
	NormalAvgDuration   *float64 `json:"normal_avg_duration,omitempty"`
	AbnormalSuccRate    *float64 `json:"abnormal_succ_rate,omitempty"`
	NormalSuccRate      *float64 `json:"normal_succ_rate,omitempty"`
	AbnormalP90         *float64 `json:"abnormal_p90,omitempty"`
	NormalP90           *float64 `json:"normal_p90,omitempty"`
	AbnormalP95         *float64 `json:"abnormal_p95,omitempty"`
	NormalP95           *float64 `json:"normal_p95,omitempty"`
	AbnormalP99         *float64 `json:"abnormal_p99,omitempty"`
	NormalP99           *float64 `json:"normal_p99,omitempty"`
}

func (item *DetectorResultItem) Validate() error {
	if item.SpanName == "" {
		return fmt.Errorf("span_name cannot be empty")
	}
	if item.Issues == "" {
		return fmt.Errorf("issues cannot be empty")
	}

	if item.AbnormalSuccRate != nil && (*item.AbnormalSuccRate < 0 || *item.AbnormalSuccRate > 1) {
		return fmt.Errorf("abnormal_succ_rate must be between 0-1")
	}
	if item.NormalSuccRate != nil && (*item.NormalSuccRate < 0 || *item.NormalSuccRate > 1) {
		return fmt.Errorf("normal_succ_rate must be between 0-1")
	}

	if item.AbnormalAvgDuration != nil && *item.AbnormalAvgDuration < 0 {
		return fmt.Errorf("abnormal_avg_duration cannot be negative")
	}
	if item.NormalAvgDuration != nil && *item.NormalAvgDuration < 0 {
		return fmt.Errorf("normal_avg_duration cannot be negative")
	}

	return nil
}

func (item DetectorResultItem) ConvertToDetectorResult(executionID int) *database.DetectorResult {
	return &database.DetectorResult{
		SpanName:            item.SpanName,
		Issues:              item.Issues,
		AbnormalAvgDuration: item.AbnormalAvgDuration,
		NormalAvgDuration:   item.NormalAvgDuration,
		AbnormalSuccRate:    item.AbnormalSuccRate,
		NormalSuccRate:      item.NormalSuccRate,
		AbnormalP90:         item.AbnormalP90,
		NormalP90:           item.NormalP90,
		AbnormalP95:         item.AbnormalP95,
		NormalP95:           item.NormalP95,
		AbnormalP99:         item.AbnormalP99,
		NormalP99:           item.NormalP99,
		ExecutionID:         executionID,
	}
}

func NewDetectorResultItem(result *database.DetectorResult) DetectorResultItem {
	return DetectorResultItem{
		SpanName:            result.SpanName,
		Issues:              result.Issues,
		AbnormalAvgDuration: result.AbnormalAvgDuration,
		NormalAvgDuration:   result.NormalAvgDuration,
		AbnormalSuccRate:    result.AbnormalSuccRate,
		NormalSuccRate:      result.NormalSuccRate,
		AbnormalP90:         result.AbnormalP90,
		NormalP90:           result.NormalP90,
		AbnormalP95:         result.AbnormalP95,
		NormalP95:           result.NormalP95,
		AbnormalP99:         result.AbnormalP99,
		NormalP99:           result.NormalP99,
	}
}

// GranularityResultItem Single granularity result item
type GranularityResultItem struct {
	Level      string  `json:"level" binding:"required"`
	Result     string  `json:"result" binding:"required"`
	Rank       int     `json:"rank" binding:"required"`
	Confidence float64 `json:"confidence" binding:"omitempty"`
}

func (item *GranularityResultItem) Valiate() error {
	if item.Level == "" {
		return fmt.Errorf("level cannot be empty")
	}
	if item.Result == "" {
		return fmt.Errorf("result cannot be empty")
	}
	if item.Rank <= 0 {
		return fmt.Errorf("rank must be positive")
	}
	if item.Confidence != 0 && (item.Confidence < 0 || item.Confidence > 1) {
		return fmt.Errorf("confidence must be between 0-1")
	}
	return nil
}

func (item *GranularityResultItem) ConvertToGranularityResult(executionID int) *database.GranularityResult {
	return &database.GranularityResult{
		Level:       item.Level,
		Result:      item.Result,
		Rank:        item.Rank,
		Confidence:  item.Confidence,
		ExecutionID: executionID,
	}
}

func NewGranularityResultItem(result *database.GranularityResult) GranularityResultItem {
	return GranularityResultItem{
		Level:      result.Level,
		Result:     result.Result,
		Rank:       result.Rank,
		Confidence: result.Confidence,
	}
}

// DetectorResultRequest Detector result upload request
type UploadDetectorResultReq struct {
	Duration float64              `json:"duration" binding:"required"` // Execution duration in seconds
	Results  []DetectorResultItem `json:"results" binding:"required"`
}

func (req *UploadDetectorResultReq) Validate() error {
	if len(req.Results) == 0 {
		return fmt.Errorf("at least one detection result is required")
	}

	for i, result := range req.Results {
		if err := result.Validate(); err != nil {
			return fmt.Errorf("validation failed for result %d: %w", i+1, err)
		}
	}

	return nil
}

// HasAnomalies checks if detector results contain anomalies
func (req *UploadDetectorResultReq) HasAnomalies() bool {
	for _, result := range req.Results {
		if result.Issues != "{}" && result.Issues != "" {
			return true
		}
	}
	return false
}

// GranularityResultRequest Granularity result upload request
type UploadGranularityResultReq struct {
	Duration float64                 `json:"duration" binding:"required"` // Execution duration in seconds
	Results  []GranularityResultItem `json:"results" binding:"required,dive,required"`
}

func (req *UploadGranularityResultReq) Validate() error {
	if len(req.Results) == 0 {
		return fmt.Errorf("at least one granularity result is required")
	}

	rankMap := make(map[int]bool)
	for i, result := range req.Results {
		if err := result.Valiate(); err != nil {
			return fmt.Errorf("validation failed for result %d: %w", i+1, err)
		}

		if rankMap[result.Rank] {
			return fmt.Errorf("rank %d appeared repeatedly", result.Rank)
		}
		rankMap[result.Rank] = true
	}

	return nil
}

// UploadExecutionResultResp Execution result upload response
type UploadExecutionResultResp struct {
	ResultCount  int       `json:"result_count"`
	UploadedAt   time.Time `json:"uploaded_at"`
	HasAnomalies bool      `json:"has_anomalies,omitempty"` // Only included for detector results
}
