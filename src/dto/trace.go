package dto

import (
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

type TraceReq struct {
	TraceID string `uri:"trace_id" binding:"required"`
}

type TraceStreamReq struct {
	LastID string `bind:"last_event_id"`
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
