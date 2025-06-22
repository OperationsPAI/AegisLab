package dto

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/google/uuid"
)

type InjectCancelResp struct{}

type InjectionConfReq struct {
	Namespace string `form:"namespace" binding:"required"`
	Mode      string `form:"mode" binding:"oneof=display engine"`
}

type InjectionItem struct {
	ID          int            `json:"id"`
	TaskID      string         `json:"task_id"`
	FaultType   string         `json:"fault_type"`
	Status      string         `json:"status"`
	Spec        map[string]any `json:"spec" swaggertype:"object"`
	PreDuration int            `json:"pre_duration"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     time.Time      `json:"end_time"`
}

func (i *InjectionItem) Convert(record database.FaultInjectionSchedule) error {
	var config map[string]any
	if err := json.Unmarshal([]byte(record.DisplayConfig), &config); err != nil {
		return err
	}

	i.ID = record.ID
	i.TaskID = record.TaskID
	i.FaultType = chaos.ChaosTypeMap[chaos.ChaosType(record.FaultType)]
	i.Status = DatasetStatusMap[record.Status]
	i.Spec = config
	i.PreDuration = record.PreDuration
	i.StartTime = record.StartTime
	i.EndTime = record.EndTime

	return nil
}

type InjectionConfigListReq struct {
	TraceIDs []string `form:"trace_ids" binding:"required"`
}

func (req *InjectionConfigListReq) Validate() error {
	filteredIDs := make([]string, 0, len(req.TraceIDs))
	for _, id := range req.TraceIDs {
		if strings.TrimSpace(id) != "" {
			filteredIDs = append(filteredIDs, strings.TrimSpace(id))
		}
	}

	req.TraceIDs = filteredIDs
	if len(req.TraceIDs) == 0 {
		return fmt.Errorf("trace_ids must not be blank")
	}

	for _, id := range req.TraceIDs {
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("Invalid trace_id: %s", id)
		}
	}

	return nil
}

type InjectionListReq struct {
	PaginationReq
}

type InjectionNamespaceInfoResp struct {
	NamespaceInfo map[string][]string `json:"namespace_info" swaggertype:"object"`
}

type InjectionParaResp struct {
	Specification map[string][]chaos.ActionSpace `json:"specification" swaggertype:"object"`
	KeyMap        map[chaos.ChaosType]string     `json:"keymap" swaggertype:"object"`
}

type LabelItem struct {
	Key   string `json:"key" binding:"required,oneof=env batch"`
	Value string `json:"value" binding:"required"`
}

type InjectionSubmitReq struct {
	Interval     int          `json:"interval"`
	PreDuration  int          `json:"pre_duration"`
	Specs        []chaos.Node `json:"specs"`
	Benchmark    string       `json:"benchmark"`
	Algorithms   []string     `json:"algorithms"`
	Labels       []LabelItem  `json:"labels" binding:"omitempty,dive"`
	DirectInject bool         `json:"direct" binding:"omitempty"`
}

type InjectionSubmitResp struct {
	SubmitResp
	DuplicatedCount int `json:"duplicated_count"`
	OriginalCount   int `json:"original_count"`
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

type QueryInjectionReq struct {
	Name   string `form:"name" binding:"omitempty,max=64"`
	TaskID string `form:"task_id" binding:"omitempty,max=64"`
}

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

type FaultInjectionWithIssuesReq struct {
	TimeRangeQuery
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
