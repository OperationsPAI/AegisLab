package dto

import (
	"fmt"
	"time"
)

// DetectorResultRequest Detector result upload request
type DetectorResultRequest struct {
	Results []DetectorResultItem `json:"results" binding:"required,dive,required"`
}

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

// GranularityResultRequest Granularity result upload request
type GranularityResultRequest struct {
	Results []GranularityResultItem `json:"results" binding:"required,dive,required"`
}

// GranularityResultItem Single granularity result item
type GranularityResultItem struct {
	Level      string  `json:"level" binding:"required"`
	Result     string  `json:"result" binding:"required"`
	Rank       int     `json:"rank" binding:"required,min=1"`
	Confidence float64 `json:"confidence" binding:"required,min=0,max=1"`
}

// AlgorithmResultUploadResponse Algorithm result upload response
type AlgorithmResultUploadResponse struct {
	ExecutionID  int       `json:"execution_id"`
	AlgorithmID  int       `json:"algorithm_id"`
	ResultCount  int       `json:"result_count"`
	UploadedAt   time.Time `json:"uploaded_at"`
	HasAnomalies bool      `json:"has_anomalies,omitempty"` // Only included for detector results
	Message      string    `json:"message"`
}

// Validate validates the detector result request
func (req *DetectorResultRequest) Validate() error {
	if len(req.Results) == 0 {
		return fmt.Errorf("at least one detection result is required")
	}

	for i, result := range req.Results {
		if result.SpanName == "" {
			return fmt.Errorf("span_name cannot be empty for result %d", i+1)
		}
		if result.Issues == "" {
			return fmt.Errorf("issues cannot be empty for result %d", i+1)
		}

		// Validate percentage value range
		if result.AbnormalSuccRate != nil && (*result.AbnormalSuccRate < 0 || *result.AbnormalSuccRate > 1) {
			return fmt.Errorf("abnormal_succ_rate must be between 0-1 for result %d", i+1)
		}
		if result.NormalSuccRate != nil && (*result.NormalSuccRate < 0 || *result.NormalSuccRate > 1) {
			return fmt.Errorf("normal_succ_rate must be between 0-1 for result %d", i+1)
		}

		// Validate non-negative duration
		if result.AbnormalAvgDuration != nil && *result.AbnormalAvgDuration < 0 {
			return fmt.Errorf("abnormal_avg_duration cannot be negative for result %d", i+1)
		}
		if result.NormalAvgDuration != nil && *result.NormalAvgDuration < 0 {
			return fmt.Errorf("normal_avg_duration cannot be negative for result %d", i+1)
		}
	}

	return nil
}

// Validate validates the granularity result request
func (req *GranularityResultRequest) Validate() error {
	if len(req.Results) == 0 {
		return fmt.Errorf("at least one granularity result is required")
	}

	rankMap := make(map[int]bool)
	for i, result := range req.Results {
		if result.Level == "" {
			return fmt.Errorf("level cannot be empty for result %d", i+1)
		}
		if result.Result == "" {
			return fmt.Errorf("result cannot be empty for result %d", i+1)
		}
		if result.Rank <= 0 {
			return fmt.Errorf("rank must be greater than 0 for result %d", i+1)
		}
		if result.Confidence < 0 || result.Confidence > 1 {
			return fmt.Errorf("confidence must be between 0-1 for result %d", i+1)
		}

		// Check for duplicate ranks
		if rankMap[result.Rank] {
			return fmt.Errorf("rank %d appeared repeatedly", result.Rank)
		}
		rankMap[result.Rank] = true
	}

	return nil
}

// HasAnomalies checks if detector results contain anomalies
func (req *DetectorResultRequest) HasAnomalies() bool {
	for _, result := range req.Results {
		if result.Issues != "{}" && result.Issues != "" {
			return true
		}
	}
	return false
}

// GranularityResultEnhancedRequest Enhanced granularity result upload request
type GranularityResultEnhancedRequest struct {
	DatapackID int                     `json:"datapack_id,omitempty"` // Required if no execution_id
	Results    []GranularityResultItem `json:"results" binding:"required,dive,required"`
}

// Validate validates the enhanced granularity result request
func (req *GranularityResultEnhancedRequest) Validate() error {
	if len(req.Results) == 0 {
		return fmt.Errorf("at least one granularity result is required")
	}

	rankMap := make(map[int]bool)
	for i, result := range req.Results {
		if result.Level == "" {
			return fmt.Errorf("level cannot be empty for result %d", i+1)
		}
		if result.Result == "" {
			return fmt.Errorf("result cannot be empty for result %d", i+1)
		}
		if result.Rank <= 0 {
			return fmt.Errorf("rank must be greater than 0 for result %d", i+1)
		}
		if result.Confidence < 0 || result.Confidence > 1 {
			return fmt.Errorf("confidence must be between 0-1 for result %d", i+1)
		}

		// Check for duplicate ranks
		if rankMap[result.Rank] {
			return fmt.Errorf("rank %d appeared repeatedly", result.Rank)
		}
		rankMap[result.Rank] = true
	}

	return nil
}
