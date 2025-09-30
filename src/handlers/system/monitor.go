package system

import (
	"aegis/dto"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// GetSystemInfo handles basic system information
//
//	@Summary Get system information
//	@Description Get basic system information and status
//	@Tags System
//	@Produce json
//	@Success 200 {object} dto.GenericResponse[dto.SystemInfo] "System info retrieved successfully"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /system/monitor/info [get]
func GetSystemInfo(c *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	info := dto.SystemInfo{
		CPUUsage:    25.5, // Mock value - should get from system
		MemoryUsage: float64(memStats.Alloc) / float64(memStats.Sys) * 100,
		DiskUsage:   45.8,            // Mock value - should get from system
		LoadAverage: "1.2, 1.5, 1.8", // Mock value - should get from system
	}

	dto.SuccessResponse(c, info)
}

// GetMetrics handles monitoring metrics query
//
//	@Summary Get monitoring metrics
//	@Description Query monitoring metrics for system performance
//	@Tags System
//	@Accept json
//	@Produce json
//	@Param request body dto.MonitoringQueryRequest true "Metrics query request"
//	@Success 200 {object} dto.GenericResponse[dto.MonitoringMetricsResponse] "Metrics retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /system/monitor/metrics [post]
func GetMetrics(c *gin.Context) {
	var req dto.MonitoringQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// TODO: Implement real metrics querying from monitoring system
	// For now, return mock data
	metrics := map[string]dto.MetricValue{
		"cpu_usage": {
			Value:     25.5,
			Timestamp: time.Now(),
			Unit:      "percent",
		},
		"memory_usage": {
			Value:     60.2,
			Timestamp: time.Now(),
			Unit:      "percent",
		},
		"disk_usage": {
			Value:     45.8,
			Timestamp: time.Now(),
			Unit:      "percent",
		},
		"active_connections": {
			Value:     142,
			Timestamp: time.Now(),
			Unit:      "count",
		},
	}

	labels := map[string]string{
		"instance": "rcabench-01",
		"version":  "1.0.0",
	}

	response := dto.MonitoringMetricsResponse{
		Timestamp: time.Now(),
		Metrics:   metrics,
		Labels:    labels,
	}

	dto.SuccessResponse(c, response)
}
