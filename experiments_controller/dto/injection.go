package dto

import (
	"time"

	chaos "github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/executor"
)

type InjectCancelResp struct {
}

type InjectionDetailResp struct {
	Task InjectionTask `json:"task"`
	Logs []string      `json:"logs"`
}

type InjectionItem struct {
	ID              int                    `json:"id"`
	TaskID          string                 `json:"task_id"`
	FaultType       string                 `json:"fault_type"`
	Name            string                 `gorm:"column:injection_name" json:"name"`
	Status          string                 `json:"status"`
	InjectTime      time.Time              `gorm:"column:start_time" json:"inject_time"`
	ProposedEndTime time.Time              `json:"proposed_end_time"`
	Duration        int                    `json:"duration"`
	Payload         executor.InjectionMeta `gorm:"-" json:"payload"`
}

type InjectionListReq struct {
	PaginationReq
}

type InjectionNamespaceInfoResp struct {
	NamespaceInfo map[string][]string `json:"namespace_info"`
}

type InjectionParaResp struct {
	Specification map[string][]chaos.ActionSpace `json:"specification"`
	KeyMap        map[chaos.ChaosType]string     `json:"keymap"`
}

type InjectionPayload struct {
	FaultDuration int            `json:"fault_duration"`
	FaultType     int            `json:"fault_type"`
	Namespace     string         `json:"inject_namespace"`
	Pod           string         `json:"inject_pod"`
	InjectSpec    map[string]int `json:"spec"`

	Benchmark     string     `json:"benchmark"`
	ExecutionTime *time.Time `json:"execution_time,omitempty"`
	PreDuration   int        `json:"pre_duration"`
}

type timeRange struct {
	Start time.Time
	End   time.Time
}

// 检查两个时间段是否重叠
func isOverlap(a, b timeRange) bool {
	return a.End.After(b.Start)
}

func (i *InjectionPayload) GetTimeRange() timeRange {
	executionTime := i.ExecutionTime
	preStart := executionTime.Add(-time.Duration(i.PreDuration) * time.Minute)
	faultEnd := executionTime.Add(time.Duration(i.FaultDuration) * time.Minute)
	return timeRange{Start: preStart, End: faultEnd}
}

type InjectionSubmitReq struct {
	IsCroned bool               `json:"is_croned"`
	Interval int                `json:"interval"`
	Payloads []InjectionPayload `json:"payloads"`
}

// 检查所有任务的时间冲突
func (r *InjectionSubmitReq) CheckConflicts() bool {
	if len(r.Payloads) <= 1 {
		return false
	}

	var allRanges []timeRange
	for _, payload := range r.Payloads {
		allRanges = append(allRanges, payload.GetTimeRange())
	}

	// 检查时间段是否重叠
	for i := range len(allRanges) - 1 {
		if isOverlap(allRanges[i], allRanges[i+1]) {
			return true
		}
	}
	return false
}

type InjectionTask struct {
	ID        string           `json:"id"`
	Type      string           `json:"type"`
	Payload   InjectionPayload `json:"payload"`
	Status    string           `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
}
