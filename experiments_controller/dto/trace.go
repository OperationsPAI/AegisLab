package dto

import (
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/consts"
)

var ValidTaskEventMap = map[consts.TaskType][]consts.EventType{
	consts.TaskTypeBuildDataset: {
		consts.EventDatasetBuildSucceed,
	},
	consts.TaskTypeCollectResult: {
		consts.EventDatasetResultCollection,
		consts.EventDatasetNoAnomaly,
		consts.EventDatasetNoConclusionFile,
	},
	consts.TaskTypeFaultInjection: {
		consts.EventFaultInjectionStarted,
		consts.EventFaultInjectionCompleted,
		consts.EventFaultInjectionFailed,
	},
	consts.TaskTypeRunAlgorithm: {
		consts.EventAlgoRunSucceed,
	},
	consts.TaskTypeRestartService: {
		consts.EventNoNamespaceAvailable,
		consts.EventRestartServiceStarted,
		consts.EventRestartServiceCompleted,
		consts.EventRestartServiceFailed,
	},
}

var ValidTaskTypes = map[consts.TaskType]struct{}{
	consts.TaskTypeBuildDataset:   {},
	consts.TaskTypeRestartService: {},
	consts.TaskTypeRunAlgorithm:   {},
}

type TraceReq struct {
	TraceID string `uri:"trace_id" binding:"required"`
}

type TraceStreamReq struct {
	LastID string `bind:"last_event_id"`
}

type TraceAnalyzeFilterOptions struct {
	FirstTaskType consts.TaskType
	TimeFilterOption
}

type GetCompletedMapReq struct {
	TimeRangeQuery
}

type GetCompletedMapResp struct {
	AnomalyTraces   []string `json:"has_anomaly"` // 检测到异常的链路ID列表
	NoAnomalyTraces []string `json:"no_anomaly"`  // 没有异常的链路ID列表
}

func (req *GetCompletedMapReq) Validate() error {
	return req.TimeRangeQuery.Validate()
}

func (req *GetCompletedMapReq) Convert() (*TraceAnalyzeFilterOptions, error) {
	opts, err := req.TimeRangeQuery.Convert()
	if err != nil {
		return nil, err
	}

	return &TraceAnalyzeFilterOptions{
		TimeFilterOption: *opts,
	}, nil
}

type TraceAnalyzeReq struct {
	FirstTaskType string `form:"first_task_type" binding:"omitempty"`
	TimeRangeQuery
}

func (req *TraceAnalyzeReq) Validate() error {
	if req.FirstTaskType != "" {
		if _, exists := ValidTaskTypes[consts.TaskType(req.FirstTaskType)]; !exists {
			return fmt.Errorf("Invalid event name: %s", req.FirstTaskType)
		}
	}

	return req.TimeRangeQuery.Validate()
}

func (req *TraceAnalyzeReq) Convert() (*TraceAnalyzeFilterOptions, error) {
	opts, err := req.TimeRangeQuery.Convert()
	if err != nil {
		return nil, err
	}

	return &TraceAnalyzeFilterOptions{
		FirstTaskType:    consts.TaskType(req.FirstTaskType),
		TimeFilterOption: *opts,
	}, nil
}
