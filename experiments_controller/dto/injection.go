package dto

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
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

type InjectionSubmitReq struct {
	Interval    int          `json:"interval"`
	PreDuration int          `json:"pre_duration"`
	Specs       []chaos.Node `json:"specs"`
	Benchmark   string       `json:"benchmark"`
}

type InjectionConfig struct {
	Index         int
	FaultType     int
	FaultDuration int
	DisplayData   string
	Conf          *chaos.InjectionConf
	Node          *chaos.Node
	ExecuteTime   time.Time
}

func (r *InjectionSubmitReq) ParseInjectionSpecs() ([]*InjectionConfig, error) {
	if len(r.Specs) == 0 {
		return nil, fmt.Errorf("spec must not be blank")
	}

	intervalDuration := time.Duration(r.Interval) * consts.DefaultTimeUnit
	currentTime := time.Now()
	configs := make([]*InjectionConfig, 0, len(r.Specs))
	for idx, spec := range r.Specs {

		childNode, exists := spec.Children[strconv.Itoa(spec.Value)]
		if !exists {
			return nil, fmt.Errorf("failed to find key %d in the children", spec.Value)
		}

		nsPrefixs := config.GetNsPrefixs()
		nsCountMap, err := config.GetNsCountMap()
		if err != nil {
			return nil, fmt.Errorf("failed to get namespace target map in configuration")
		}

		index := childNode.Children[consts.NamespaceNodeKey].Value
		namespaceCount := nsCountMap[nsPrefixs[index]]

		var execTime time.Time
		if idx < namespaceCount {
			execTime = currentTime.Add(time.Second * time.Duration(rand.Int()%20)) // random delay
		} else {
			execTime = currentTime.Add(intervalDuration * time.Duration(idx/namespaceCount)).Add(time.Second * time.Duration(rand.Int()%60))
		}

		faultDuration := childNode.Children[consts.DurationNodeKey].Value

		conf, err := chaos.NodeToStruct[chaos.InjectionConf](&spec)
		if err != nil {
			return nil, fmt.Errorf("failed to convert node to injecton conf: %v", err)
		}

		displayConfig, err := conf.GetDisplayConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get display config: %v", err)
		}

		displayData, err := json.Marshal(displayConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal injection spec to display config: %v", err)
		}

		configs = append(configs, &InjectionConfig{
			Index:         idx,
			FaultType:     spec.Value,
			FaultDuration: faultDuration,
			DisplayData:   string(displayData),
			Conf:          conf,
			Node:          &spec,
			ExecuteTime:   execTime,
		})
	}

	return configs, nil
}

type QueryInjectionReq struct {
	Name   string `form:"name" binding:"omitempty,max=64"`
	TaskID string `form:"task_id" binding:"omitempty,max=64"`
}

// FaultInjectionAnalysisReq 故障注入分析请求参数
type FaultInjectionAnalysisReq struct {
	PageNum  int `form:"page_num" binding:"min=1" json:"page_num"`
	PageSize int `form:"page_size" binding:"min=1,max=100" json:"page_size"`
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
