package dto

import "time"

// HealthCheckResponse represents system health check response
type HealthCheckResponse struct {
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

// SystemStatisticsResponse represents system statistics response
type SystemStatisticsResponse struct {
	Users       UserStatistics      `json:"users"`
	Projects    ProjectStatistics   `json:"projects"`
	Tasks       TaskStatistics      `json:"tasks"`
	Containers  ContainerStatistics `json:"containers"`
	Datasets    DatasetStatistics   `json:"datasets"`
	Injections  InjectionStatistics `json:"injections"`
	Executions  ExecutionStatistics `json:"executions"`
	GeneratedAt time.Time           `json:"generated_at"`
}

// UserStatistics represents user-related statistics
type UserStatistics struct {
	Total       int `json:"total"`
	Active      int `json:"active"`
	Inactive    int `json:"inactive"`
	NewToday    int `json:"new_today"`
	NewThisWeek int `json:"new_this_week"`
}

// ProjectStatistics represents project-related statistics
type ProjectStatistics struct {
	Total    int `json:"total"`
	Active   int `json:"active"`
	Inactive int `json:"inactive"`
	NewToday int `json:"new_today"`
}

// TaskStatistics represents task-related statistics
type TaskStatistics struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Running   int `json:"running"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

// ContainerStatistics represents container-related statistics
type ContainerStatistics struct {
	Total   int `json:"total"`
	Active  int `json:"active"`
	Deleted int `json:"deleted"`
}

// DatasetStatistics represents dataset-related statistics
type DatasetStatistics struct {
	Total     int   `json:"total"`
	Public    int   `json:"public"`
	Private   int   `json:"private"`
	TotalSize int64 `json:"total_size"`
}

// InjectionStatistics represents injection-related statistics
type InjectionStatistics struct {
	Total     int `json:"total"`
	Scheduled int `json:"scheduled"`
	Running   int `json:"running"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

// ExecutionStatistics represents execution-related statistics
type ExecutionStatistics struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Failed     int `json:"failed"`
}

// SystemInfo represents system information
type SystemInfo struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	LoadAverage string  `json:"load_average"`
}

// AuditLogResponse represents audit log response
type AuditLogResponse struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Username   string    `json:"username"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id"`
	Details    string    `json:"details"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// AuditLogRequest represents audit log creation request
type AuditLogRequest struct {
	Action     string `json:"action" binding:"required"`
	Resource   string `json:"resource" binding:"required"`
	ResourceID *int   `json:"resource_id"`
	Details    string `json:"details"`
}

// AuditLogListRequest represents audit log list query parameters
type AuditLogListRequest struct {
	Page      int       `form:"page,default=1" binding:"min=1"`
	Size      int       `form:"size,default=20" binding:"min=1,max=100"`
	UserID    *int      `form:"user_id"`
	Action    string    `form:"action"`
	Resource  string    `form:"resource"`
	Success   *bool     `form:"success"`
	StartDate time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate   time.Time `form:"end_date" time_format:"2006-01-02"`
}

// AuditLogListResponse represents paginated audit log list response
type AuditLogListResponse struct {
	Items      []AuditLogResponse `json:"items"`
	Pagination PaginationInfo     `json:"pagination"`
}

// MonitoringMetricsResponse represents monitoring metrics response
type MonitoringMetricsResponse struct {
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]MetricValue `json:"metrics"`
	Labels    map[string]string      `json:"labels,omitempty"`
}

// MetricValue represents a single metric value
type MetricValue struct {
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	Unit      string    `json:"unit,omitempty"`
}

// MonitoringQueryRequest represents monitoring query request
type MonitoringQueryRequest struct {
	Query     string    `json:"query" binding:"required"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Step      string    `json:"step,omitempty"`
}
