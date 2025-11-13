package dto

import (
	"time"
)

// HealthCheckResp represents system health check response
type HealthCheckResp struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Services  map[string]ServiceInfo `json:"services" swaggertype:"object"`
}

// ServiceInfo represents individual service health information
type ServiceInfo struct {
	Status       string    `json:"status"`
	LastChecked  time.Time `json:"last_checked"`
	ResponseTime string    `json:"response_time"`
	Error        string    `json:"error,omitempty"`
	Details      any       `json:"details,omitempty"`
}

type NamespaceMonitorItem struct {
	LockedBy string    `json:"locked_by"`
	EndTime  time.Time `json:"end_time"`
}

type ListNamespaceLockResp struct {
	Items map[string]NamespaceMonitorItem `json:"items" swaggertype:"object"`
}

// SystemInfo represents system information
type SystemInfo struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	LoadAverage string  `json:"load_average"`
}

// MonitoringQueryReq represents monitoring query request
type MonitoringQueryReq struct {
	Query     string    `json:"query" binding:"required"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Step      string    `json:"step,omitempty"`
}

// MetricValue represents a single metric value
type MetricValue struct {
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	Unit      string    `json:"unit,omitempty"`
}

// MonitoringMetricsResp represents monitoring metrics response
type MonitoringMetricsResp struct {
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]MetricValue `json:"metrics"`
	Labels    map[string]string      `json:"labels,omitempty"`
}
