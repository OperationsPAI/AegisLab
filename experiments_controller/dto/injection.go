package dto

import (
	"fmt"
	"strconv"
	"time"

	chaos "github.com/CUHK-SE-Group/chaos-experiment/handler"
)

type InjectCancelResp struct {
}

type InjectionDetailResp struct {
	Task InjectionTask `json:"task"`
	Logs []string      `json:"logs"`
}

type InjectionItem struct {
	ID         int            `json:"id"`
	TaskID     string         `json:"task_id"`
	FaultType  string         `json:"fault_type"`
	Name       string         `gorm:"column:injection_name" json:"name"`
	Status     string         `json:"status"`
	InjectTime time.Time      `gorm:"column:start_time" json:"inject_time"`
	Duration   int            `json:"duration"`
	Payload    map[string]any `json:"payload"`
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
	InjectionSpec map[string]any `json:"spec"`
	Benchmark     string         `json:"benchmark"`
	ExecutionTime *time.Time     `json:"execution_time,omitempty"`
	PreDuration   int            `json:"pre_duration"`
}

type timeRange struct {
	Start time.Time
	End   time.Time
}

type InjectionSubmitReq struct {
	Interval    int              `json:"interval"`
	PreDuration int              `json:"pre_duration"`
	Specs       []map[string]any `json:"specs"`
	Benchmark   string           `json:"benchmark"`
}

func (r *InjectionSubmitReq) GetExecutionTimes() ([]time.Time, error) {
	if len(r.Specs) == 0 {
		return nil, nil
	}

	executionTimes := make([]time.Time, 0, len(r.Specs))
	timeRanges := make([]timeRange, 0, len(r.Specs))

	currentTime := time.Now()
	for i, spec := range r.Specs {
		faultDuration, err := extractFaultDuration(spec)
		if err != nil {
			return nil, fmt.Errorf("spec[%d]: %w", i, err)
		}

		execTime := currentTime.Add(time.Duration(i*r.Interval) * time.Minute)
		start := execTime.Add(-time.Duration(r.PreDuration) * time.Minute)
		end := execTime.Add(time.Duration(faultDuration) * time.Minute)

		executionTimes = append(executionTimes, execTime)
		timeRanges = append(timeRanges, timeRange{Start: start, End: end})

		if i > 0 && timeRanges[i-1].End.After(start) {
			return nil, fmt.Errorf("spec[%d]: time range overlaps with previous", i)
		}
	}

	return executionTimes, nil
}

// 提取故障持续时间的辅助函数
func extractFaultDuration(spec map[string]any) (int, error) {
	node, err := chaos.MapToNode(spec)
	if err != nil {
		return 0, fmt.Errorf("convert spec to node failed: %w", err)
	}

	if _, err := chaos.NodeToStruct[chaos.InjectionConf](node); err != nil {
		return 0, fmt.Errorf(err.Error())
	}

	var key string
	for key = range node.Children {
	}

	subNode := node.Children[key]
	faultDuration := subNode.Children[strconv.Itoa(0)].Value

	return faultDuration, nil
}

type InjectionTask struct {
	ID        string           `json:"id"`
	Type      string           `json:"type"`
	Payload   InjectionPayload `json:"payload"`
	Status    string           `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
}
