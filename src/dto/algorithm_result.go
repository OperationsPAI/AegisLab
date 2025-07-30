package dto

import (
	"fmt"
	"time"
)

// DetectorResultRequest 检测器结果上传请求
type DetectorResultRequest struct {
	Results []DetectorResultItem `json:"results" binding:"required,dive,required"`
}

// DetectorResultItem 单个检测器结果项
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

// GranularityResultRequest 粒度结果上传请求
type GranularityResultRequest struct {
	Results []GranularityResultItem `json:"results" binding:"required,dive,required"`
}

// GranularityResultItem 单个粒度结果项
type GranularityResultItem struct {
	Level      string  `json:"level" binding:"required"`
	Result     string  `json:"result" binding:"required"`
	Rank       int     `json:"rank" binding:"required,min=1"`
	Confidence float64 `json:"confidence" binding:"required,min=0,max=1"`
}

// AlgorithmResultUploadResponse 算法结果上传响应
type AlgorithmResultUploadResponse struct {
	ExecutionID  int       `json:"execution_id"`
	AlgorithmID  int       `json:"algorithm_id"`
	ResultCount  int       `json:"result_count"`
	UploadedAt   time.Time `json:"uploaded_at"`
	HasAnomalies bool      `json:"has_anomalies,omitempty"` // 仅检测器结果包含
	Message      string    `json:"message"`
}

// Validate 验证检测器结果请求
func (req *DetectorResultRequest) Validate() error {
	if len(req.Results) == 0 {
		return fmt.Errorf("至少需要一个检测结果")
	}

	for i, result := range req.Results {
		if result.SpanName == "" {
			return fmt.Errorf("第%d个结果的span_name不能为空", i+1)
		}
		if result.Issues == "" {
			return fmt.Errorf("第%d个结果的issues不能为空", i+1)
		}

		// 验证百分比值范围
		if result.AbnormalSuccRate != nil && (*result.AbnormalSuccRate < 0 || *result.AbnormalSuccRate > 1) {
			return fmt.Errorf("第%d个结果的abnormal_succ_rate必须在0-1之间", i+1)
		}
		if result.NormalSuccRate != nil && (*result.NormalSuccRate < 0 || *result.NormalSuccRate > 1) {
			return fmt.Errorf("第%d个结果的normal_succ_rate必须在0-1之间", i+1)
		}

		// 验证持续时间为非负数
		if result.AbnormalAvgDuration != nil && *result.AbnormalAvgDuration < 0 {
			return fmt.Errorf("第%d个结果的abnormal_avg_duration不能为负数", i+1)
		}
		if result.NormalAvgDuration != nil && *result.NormalAvgDuration < 0 {
			return fmt.Errorf("第%d个结果的normal_avg_duration不能为负数", i+1)
		}
	}

	return nil
}

// Validate 验证粒度结果请求
func (req *GranularityResultRequest) Validate() error {
	if len(req.Results) == 0 {
		return fmt.Errorf("至少需要一个粒度结果")
	}

	rankMap := make(map[int]bool)
	for i, result := range req.Results {
		if result.Level == "" {
			return fmt.Errorf("第%d个结果的level不能为空", i+1)
		}
		if result.Result == "" {
			return fmt.Errorf("第%d个结果的result不能为空", i+1)
		}
		if result.Rank <= 0 {
			return fmt.Errorf("第%d个结果的rank必须大于0", i+1)
		}
		if result.Confidence < 0 || result.Confidence > 1 {
			return fmt.Errorf("第%d个结果的confidence必须在0-1之间", i+1)
		}

		// 检查rank是否重复
		if rankMap[result.Rank] {
			return fmt.Errorf("rank %d 重复出现", result.Rank)
		}
		rankMap[result.Rank] = true
	}

	return nil
}

// HasAnomalies 检查检测器结果是否包含异常
func (req *DetectorResultRequest) HasAnomalies() bool {
	for _, result := range req.Results {
		if result.Issues != "{}" && result.Issues != "" {
			return true
		}
	}
	return false
}

// GranularityResultEnhancedRequest 增强版粒度结果上传请求
type GranularityResultEnhancedRequest struct {
	DatapackID int                     `json:"datapack_id,omitempty"` // 当没有execution_id时必需
	Results    []GranularityResultItem `json:"results" binding:"required,dive,required"`
}

// Validate 验证增强版粒度结果请求
func (req *GranularityResultEnhancedRequest) Validate() error {
	if len(req.Results) == 0 {
		return fmt.Errorf("至少需要一个粒度结果")
	}

	rankMap := make(map[int]bool)
	for i, result := range req.Results {
		if result.Level == "" {
			return fmt.Errorf("第%d个结果的level不能为空", i+1)
		}
		if result.Result == "" {
			return fmt.Errorf("第%d个结果的result不能为空", i+1)
		}
		if result.Rank <= 0 {
			return fmt.Errorf("第%d个结果的rank必须大于0", i+1)
		}
		if result.Confidence < 0 || result.Confidence > 1 {
			return fmt.Errorf("第%d个结果的confidence必须在0-1之间", i+1)
		}

		// 检查rank是否重复
		if rankMap[result.Rank] {
			return fmt.Errorf("rank %d 重复出现", result.Rank)
		}
		rankMap[result.Rank] = true
	}

	return nil
}
