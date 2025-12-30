package system

import (
	"aegis/dto"
	"aegis/handlers"
	producer "aegis/service/prodcuer"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// GetMetrics handles monitoring metrics query
//
//	@Summary		Get monitoring metrics
//	@Description	Query monitoring metrics for system performance
//	@Tags			System
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.MonitoringQueryReq							true	"Metrics query request"
//	@Success		200		{object}	dto.GenericResponse[dto.MonitoringMetricsResp]	"Metrics retrieved successfully"
//	@Success		400		{object}	dto.GenericResponse[any]						"Invalid request format"
//	@Failure		401		{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/system/monitor/metrics [post]
func GetMetrics(c *gin.Context) {
	var req dto.MonitoringQueryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// TODO: Implement real metrics querying from monitoring system
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

	response := dto.MonitoringMetricsResp{
		Timestamp: time.Now(),
		Metrics:   metrics,
		Labels:    labels,
	}

	dto.SuccessResponse(c, response)
}

// GetSystemInfo handles basic system information
//
//	@Summary		Get system information
//	@Description	Get basic system information and status
//	@Tags			System
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GenericResponse[dto.SystemInfo]	"System info retrieved successfully"
//	@Failure		401	{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		500	{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/system/monitor/info [get]
func GetSystemInfo(c *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// TODO : Implement real system info retrieval
	info := dto.SystemInfo{
		CPUUsage:    25.5,
		MemoryUsage: float64(memStats.Alloc) / float64(memStats.Sys) * 100,
		DiskUsage:   45.8,
		LoadAverage: "1.2, 1.5, 1.8",
	}

	dto.SuccessResponse(c, info)
}

// ListNamespaceLocks handles listing of namespace locks
//
//	@Summary		List namespace locks
//	@Description	Retrieve the list of currently locked namespaces
//	@Tags			System
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GenericResponse[dto.ListNamespaceLockResp]	"Successfully retrieved the list of locks"
//	@Failure		401	{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		500	{object}	dto.GenericResponse[any]						"Internal Server Error"
//	@Router			/system/monitor/namespaces/locks [get]
func ListNamespaceLocks(c *gin.Context) {
	items, err := producer.InspectLock(c.Request.Context())
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Successfully retrieved the list of locks", items)
}

// ListQueuedTasks handles listing of queued tasks
//
//	@Summary		List queued tasks
//	@Description	List tasks in queue (ready and delayed)
//	@Tags			System
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GenericResponse[dto.QueuedTasksResp]	"Queued tasks retrieved successfully"
//	@Failure		401	{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]					"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]					"No queued tasks found"
//	@Failure		500	{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/system/monitor/tasks/queue [post]
func ListQueuedTasks(c *gin.Context) {
	ctx := c.Request.Context()
	resp, err := producer.ListQueuedTasks(ctx)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Queued tasks retrieved successfully", resp)
}
