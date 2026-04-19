package systemmetric

import "time"

type NsMonitorItem struct {
	LockedBy string    `json:"locked_by"`
	EndTime  time.Time `json:"end_time"`
	Status   string    `json:"status"`
}

type ListNamespaceLockResp struct {
	Items map[string]NsMonitorItem `json:"items" swaggertype:"object"`
}

// MetricValue represents a single metric value.
type MetricValue struct {
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	Unit      string    `json:"unit,omitempty"`
}

// SystemMetricsResp represents current system metrics.
type SystemMetricsResp struct {
	CPU    MetricValue `json:"cpu"`
	Memory MetricValue `json:"memory"`
	Disk   MetricValue `json:"disk"`
}

// SystemMetricsHistoryResp represents historical system metrics.
type SystemMetricsHistoryResp struct {
	CPU    []MetricValue `json:"cpu"`
	Memory []MetricValue `json:"memory"`
}
