package dto

import (
	"fmt"
	"time"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/utils"
)

type InjectCancelResp struct{}

type InjectionConfReq struct {
	Namespace string `form:"namespace" binding:"required"`
	Mode      string `form:"mode" binding:"omitempty,oneof=display engine"`
}

func (req *InjectionConfReq) setDefaults() {
	if req.Mode == "" {
		req.Mode = "engine"
	}
}

func (req *InjectionConfReq) Validate() error {
	req.setDefaults()
	return nil
}

type ListDisplayConfigsReq struct {
	TraceIDs []string `form:"trace_ids" binding:"omitempty"`
}

func (req *ListDisplayConfigsReq) Validate() error {
	req.TraceIDs = utils.FilterEmptyStrings(req.TraceIDs)
	for _, traceID := range req.TraceIDs {
		if !utils.IsValidUUID(traceID) {
			return fmt.Errorf("Invalid trace_id format: %s", traceID)
		}
	}

	return nil
}

type ListInjectionsReq struct {
	Env       string `form:"env" binding:"omitempty"`
	Batch     string `form:"batch" binding:"omitempty"`
	Benchmark string `form:"benchmark" binding:"omitempty"`
	Status    *int   `form:"status" binding:"omitempty"`
	FaultType *int   `form:"fault_type" binding:"omitempty"`

	ListOptionsQuery
	TimeRangeQuery
}

func (req *ListInjectionsReq) Validate() error {
	if req.Benchmark != "" {
		if _, exists := config.GetValidBenchmarkMap()[req.Benchmark]; !exists {
			return fmt.Errorf("Invalid benchmark: %s", req.Benchmark)
		}
	}

	if req.Status != nil {
		status := *req.Status
		if status < 0 {
			return fmt.Errorf("Status must be a non-negative integer")
		}

		if _, exists := DatasetStatusMap[status]; !exists {
			return fmt.Errorf("Invalid status: %d", req.Status)
		}
	}

	if req.FaultType != nil {
		if _, exists := chaos.ChaosTypeMap[chaos.ChaosType(*req.FaultType)]; !exists {
			return fmt.Errorf("Invalid fault type: %d", req.FaultType)
		}
	}

	if err := req.ListOptionsQuery.Validate(); err != nil {
		return err
	}

	if err := req.TimeRangeQuery.Validate(); err != nil {
		return err
	}

	return nil
}

type QueryInjectionReq struct {
	Name   string `form:"name" binding:"omitempty"`
	TaskID string `form:"task_id" binding:"omitempty"`
}

func (req *QueryInjectionReq) Validate() error {
	if req.Name == "" && req.TaskID == "" {
		return fmt.Errorf("Either name or task_id must be provided")
	}

	if req.Name != "" && req.TaskID != "" {
		return fmt.Errorf("Only one of name or task_id should be provided")
	}

	if req.TaskID != "" {
		if !utils.IsValidUUID(req.TaskID) {
			return fmt.Errorf("Invalid task_id format: %s", req.TaskID)
		}
	}

	return nil
}

type LabelItem struct {
	Key   string `json:"key" binding:"required,oneof=env batch"`
	Value string `json:"value" binding:"required"`
}

type SubmitInjectionReq struct {
	Interval    int          `json:"interval" binding:"required,min=1"`
	PreDuration int          `json:"pre_duration" binding:"required,min=1"`
	Specs       []chaos.Node `json:"specs" binding:"required"`
	Benchmark   string       `json:"benchmark" binding:"required"`
	Algorithms  []string     `json:"algorithms" bindging:"omitempty"`
	Labels      []LabelItem  `json:"labels" binding:"omitempty"`
}

func (req *SubmitInjectionReq) Validate() error {
	req.Algorithms = utils.FilterEmptyStrings(req.Algorithms)

	if req.Labels == nil {
		req.Labels = make([]LabelItem, 0)
	}

	if req.Interval <= req.PreDuration {
		return fmt.Errorf("Interval must be greater than pre_duration")
	}

	if len(req.Specs) == 0 {
		return fmt.Errorf("Specs must not be empty")
	}

	if req.Benchmark == "" {
		return fmt.Errorf("Benchmark must not be blank")
	} else {
		if _, exists := config.GetValidBenchmarkMap()[req.Benchmark]; !exists {
			return fmt.Errorf("Invalid benchmark: %s", req.Benchmark)
		}
	}

	return nil
}

type InjectionConfig struct {
	Index         int
	FaultType     int
	FaultDuration int
	DisplayData   string
	Conf          *chaos.InjectionConf
	Node          *chaos.Node
	ExecuteTime   time.Time
	Labels        []LabelItem
}

type InjectionFieldMappingResp struct {
	StatusMap    map[int]string             `json:"status" swaggertype:"object"`
	FaultTypeMap map[chaos.ChaosType]string `json:"fault_type" swaggertype:"object"`
}

type SubmitInjectionResp struct {
	SubmitResp
	DuplicatedCount int `json:"duplicated_count"`
	OriginalCount   int `json:"original_count"`
}

// analysis

type FaultInjectionNoIssuesReq struct {
	TimeRangeQuery
}

// FaultInjectionNoIssuesResp 没有问题的故障注入响应
type FaultInjectionNoIssuesResp struct {
	DatasetID     int        `json:"dataset_id"`
	DisplayConfig string     `json:"display_config"`
	EngineConfig  chaos.Node `json:"engine_config"`
	PreDuration   int        `json:"pre_duration"`
	InjectionName string     `json:"injection_name"`
}

// FaultInjectionWithIssuesResp 有问题的故障注入响应
type FaultInjectionWithIssuesResp struct {
	DatasetID     int        `json:"dataset_id"`
	DisplayConfig string     `json:"display_config"`
	EngineConfig  chaos.Node `json:"engine_config"`
	PreDuration   int        `json:"pre_duration"`
	InjectionName string     `json:"injection_name"`
	Issues        string     `json:"issues"`
}

// FaultInjectionStatisticsResp 故障注入统计响应
type FaultInjectionStatisticsResp struct {
	NoIssuesCount   int64 `json:"no_issues_count"`
	WithIssuesCount int64 `json:"with_issues_count"`
	TotalCount      int64 `json:"total_count"`
}

type FaultInjectionInjectionResp struct {
	database.FaultInjectionSchedule
	GroundTruth chaos.Groundtruth `json:"ground_truth,omitempty"`
}
