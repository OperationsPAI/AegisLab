package metricmodule

import (
	"aegis/httpx"
	"net/http"

	"aegis/dto"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetInjectionMetrics handles retrieval of injection metrics
//
//	@Summary		Get injection metrics
//	@Description	Get aggregated metrics for injections including success rate, duration stats, and state distribution
//	@Tags			Metrics
//	@ID				get_injection_metrics
//	@Produce		json
//	@Security		BearerAuth
//	@Param			start_time	query		string										false	"Start time (RFC3339)"
//	@Param			end_time	query		string										false	"End time (RFC3339)"
//	@Param			fault_type	query		string										false	"Filter by fault type"
//	@Success		200			{object}	dto.GenericResponse[InjectionMetrics]		"Injection metrics"
//	@Failure		400			{object}	dto.GenericResponse[any]					"Invalid request"
//	@Failure		401			{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		500			{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/metrics/injections [get]
//	@x-api-type		{}
func (h *Handler) GetInjectionMetrics(c *gin.Context) {
	var req GetMetricsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	metrics, err := h.service.GetInjectionMetrics(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Injection metrics retrieved successfully", metrics)
}

// GetExecutionMetrics handles retrieval of execution metrics
//
//	@Summary		Get execution metrics
//	@Description	Get aggregated metrics for algorithm executions including performance stats and state distribution
//	@Tags			Metrics
//	@ID				get_execution_metrics
//	@Produce		json
//	@Security		BearerAuth
//	@Param			start_time		query		string										false	"Start time (RFC3339)"
//	@Param			end_time		query		string										false	"End time (RFC3339)"
//	@Param			algorithm_id	query		int											false	"Filter by algorithm ID"
//	@Success		200				{object}	dto.GenericResponse[ExecutionMetrics]		"Execution metrics"
//	@Failure		400				{object}	dto.GenericResponse[any]					"Invalid request"
//	@Failure		401				{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		500				{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/metrics/executions [get]
//	@x-api-type		{}
func (h *Handler) GetExecutionMetrics(c *gin.Context) {
	var req GetMetricsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	metrics, err := h.service.GetExecutionMetrics(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Execution metrics retrieved successfully", metrics)
}

// GetAlgorithmMetrics handles retrieval of algorithm comparison metrics
//
//	@Summary		Get algorithm comparison metrics
//	@Description	Get comparative metrics across different algorithms for performance analysis
//	@Tags			Metrics
//	@ID				get_algorithm_metrics
//	@Produce		json
//	@Security		BearerAuth
//	@Param			algorithm_ids	query		string										false	"Comma-separated algorithm IDs"
//	@Param			start_time		query		string										false	"Start time (RFC3339)"
//	@Param			end_time		query		string										false	"End time (RFC3339)"
//	@Success		200				{object}	dto.GenericResponse[AlgorithmMetrics]		"Algorithm metrics"
//	@Failure		400				{object}	dto.GenericResponse[any]					"Invalid request"
//	@Failure		401				{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		500				{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/metrics/algorithms [get]
//	@x-api-type		{}
func (h *Handler) GetAlgorithmMetrics(c *gin.Context) {
	var req GetMetricsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	metrics, err := h.service.GetAlgorithmMetrics(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Algorithm metrics retrieved successfully", metrics)
}
