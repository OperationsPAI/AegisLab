package dto

import "time"

// HealthCheckResponse represents system health check response
type HealthCheckResponse struct {
	Status    string                 `json:"status" example:"healthy"`
	Timestamp time.Time              `json:"timestamp" example:"2024-01-01T12:00:00Z"`
	Version   string                 `json:"version" example:"1.0.0"`
	Uptime    string                 `json:"uptime" example:"72h30m15s"`
	Services  map[string]ServiceInfo `json:"services"`
}

// ServiceInfo represents individual service health information
type ServiceInfo struct {
	Status       string      `json:"status" example:"healthy"`
	LastChecked  time.Time   `json:"last_checked" example:"2024-01-01T12:00:00Z"`
	ResponseTime string      `json:"response_time" example:"15ms"`
	Error        string      `json:"error,omitempty"`
	Details      interface{} `json:"details,omitempty"`
}

// SystemStatisticsResponse represents system statistics response
type SystemStatisticsResponse struct {
	Users       UserStatistics      `json:"users"`
	Projects    ProjectStatistics   `json:"projects"`
	Tasks       TaskStatistics      `json:"tasks"`
	Containers  ContainerStatistics `json:"containers"`
	Datasets    DatasetStatistics   `json:"datasets"`
	Injections  InjectionStatistics `json:"injections"`
	Executions  ExecutionStatistics `json:"executions"`
	GeneratedAt time.Time           `json:"generated_at" example:"2024-01-01T12:00:00Z"`
}

// UserStatistics represents user-related statistics
type UserStatistics struct {
	Total       int `json:"total" example:"150"`
	Active      int `json:"active" example:"120"`
	Inactive    int `json:"inactive" example:"30"`
	NewToday    int `json:"new_today" example:"5"`
	NewThisWeek int `json:"new_this_week" example:"20"`
}

// ProjectStatistics represents project-related statistics
type ProjectStatistics struct {
	Total    int `json:"total" example:"50"`
	Active   int `json:"active" example:"45"`
	Inactive int `json:"inactive" example:"5"`
	NewToday int `json:"new_today" example:"2"`
}

// TaskStatistics represents task-related statistics
type TaskStatistics struct {
	Total     int `json:"total" example:"1000"`
	Pending   int `json:"pending" example:"50"`
	Running   int `json:"running" example:"20"`
	Completed int `json:"completed" example:"900"`
	Failed    int `json:"failed" example:"30"`
}

// ContainerStatistics represents container-related statistics
type ContainerStatistics struct {
	Total   int `json:"total" example:"25"`
	Active  int `json:"active" example:"20"`
	Deleted int `json:"deleted" example:"5"`
}

// DatasetStatistics represents dataset-related statistics
type DatasetStatistics struct {
	Total     int   `json:"total" example:"100"`
	Public    int   `json:"public" example:"30"`
	Private   int   `json:"private" example:"70"`
	TotalSize int64 `json:"total_size" example:"1073741824"`
}

// InjectionStatistics represents injection-related statistics
type InjectionStatistics struct {
	Total     int `json:"total" example:"500"`
	Scheduled int `json:"scheduled" example:"10"`
	Running   int `json:"running" example:"5"`
	Completed int `json:"completed" example:"480"`
	Failed    int `json:"failed" example:"5"`
}

// ExecutionStatistics represents execution-related statistics
type ExecutionStatistics struct {
	Total      int `json:"total" example:"800"`
	Successful int `json:"successful" example:"750"`
	Failed     int `json:"failed" example:"50"`
}

// SystemInfo represents system information
type SystemInfo struct {
	CPUUsage    float64 `json:"cpu_usage" example:"25.5"`
	MemoryUsage float64 `json:"memory_usage" example:"60.2"`
	DiskUsage   float64 `json:"disk_usage" example:"45.8"`
	LoadAverage string  `json:"load_average" example:"1.2, 1.5, 1.8"`
}

// AuditLogResponse represents audit log response
type AuditLogResponse struct {
	ID         int       `json:"id" example:"1"`
	UserID     int       `json:"user_id" example:"1"`
	Username   string    `json:"username" example:"admin"`
	Action     string    `json:"action" example:"CREATE_USER"`
	Resource   string    `json:"resource" example:"users"`
	ResourceID string    `json:"resource_id" example:"123"`
	Details    string    `json:"details" example:"{\"username\":\"newuser\"}"`
	IPAddress  string    `json:"ip_address" example:"192.168.1.100"`
	UserAgent  string    `json:"user_agent" example:"Mozilla/5.0..."`
	Success    bool      `json:"success" example:"true"`
	Error      string    `json:"error,omitempty"`
	Timestamp  time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
}

// AuditLogRequest represents audit log creation request
type AuditLogRequest struct {
	Action     string `json:"action" binding:"required" example:"CREATE_USER"`
	Resource   string `json:"resource" binding:"required" example:"users"`
	ResourceID *int   `json:"resource_id" example:"123"`
	Details    string `json:"details" example:"{\"username\":\"newuser\"}"`
}

// AuditLogListRequest represents audit log list query parameters
type AuditLogListRequest struct {
	Page      int       `form:"page,default=1" binding:"min=1" example:"1"`
	Size      int       `form:"size,default=20" binding:"min=1,max=100" example:"20"`
	UserID    *int      `form:"user_id" example:"1"`
	Action    string    `form:"action" example:"CREATE_USER"`
	Resource  string    `form:"resource" example:"users"`
	Success   *bool     `form:"success" example:"true"`
	StartDate time.Time `form:"start_date" time_format:"2006-01-02" example:"2024-01-01"`
	EndDate   time.Time `form:"end_date" time_format:"2006-01-02" example:"2024-12-31"`
}

// AuditLogListResponse represents paginated audit log list response
type AuditLogListResponse struct {
	Items      []AuditLogResponse `json:"items"`
	Pagination PaginationInfo     `json:"pagination"`
}

// MonitoringMetricsResponse represents monitoring metrics response
type MonitoringMetricsResponse struct {
	Timestamp time.Time              `json:"timestamp" example:"2024-01-01T12:00:00Z"`
	Metrics   map[string]MetricValue `json:"metrics"`
	Labels    map[string]string      `json:"labels,omitempty"`
}

// MetricValue represents a single metric value
type MetricValue struct {
	Value     float64   `json:"value" example:"42.5"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
	Unit      string    `json:"unit,omitempty" example:"percent"`
}

// MonitoringQueryRequest represents monitoring query request
type MonitoringQueryRequest struct {
	Query     string    `json:"query" binding:"required" example:"cpu_usage"`
	StartTime time.Time `json:"start_time" example:"2024-01-01T00:00:00Z"`
	EndTime   time.Time `json:"end_time" example:"2024-01-01T23:59:59Z"`
	Step      string    `json:"step,omitempty" example:"1m"`
}
