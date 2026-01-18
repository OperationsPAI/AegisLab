package v2

import (
	"aegis/dto"
	"aegis/handlers"
	producer "aegis/service/producer"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetInjectionMetrics handles retrieval of injection metrics
//
//	@Summary		Get injection metrics
//	@Description	Get aggregated metrics for injections including success rate, duration stats, and state distribution
//	@Tags			Metrics
//	@ID				get_injection_metrics
//	@Produce		json
//	@Security		BearerAuth
//	@Param			start_time	query		string								false	"Start time (RFC3339)"
//	@Param			end_time	query		string								false	"End time (RFC3339)"
//	@Param			fault_type	query		string								false	"Filter by fault type"
//	@Success		200			{object}	dto.GenericResponse[dto.InjectionMetrics]	"Injection metrics"
//	@Failure		400			{object}	dto.GenericResponse[any]			"Invalid request"
//	@Failure		401			{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		500			{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/metrics/injections [get]
//	@x-api-type		{"sdk":"true"}
func GetInjectionMetrics(c *gin.Context) {
	var req dto.GetMetricsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	metrics, err := producer.GetInjectionMetrics(&req)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			start_time	query		string								false	"Start time (RFC3339)"
//	@Param			end_time	query		string								false	"End time (RFC3339)"
//	@Param			algorithm_id	query	int									false	"Filter by algorithm ID"
//	@Success		200			{object}	dto.GenericResponse[dto.ExecutionMetrics]	"Execution metrics"
//	@Failure		400			{object}	dto.GenericResponse[any]			"Invalid request"
//	@Failure		401			{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		500			{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/metrics/executions [get]
//	@x-api-type		{"sdk":"true"}
func GetExecutionMetrics(c *gin.Context) {
	var req dto.GetMetricsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	metrics, err := producer.GetExecutionMetrics(&req)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			algorithm_ids	query	string								false	"Comma-separated algorithm IDs"
//	@Param			start_time		query	string								false	"Start time (RFC3339)"
//	@Param			end_time		query	string								false	"End time (RFC3339)"
//	@Success		200				{object}	dto.GenericResponse[dto.AlgorithmMetrics]	"Algorithm metrics"
//	@Failure		400				{object}	dto.GenericResponse[any]			"Invalid request"
//	@Failure		401				{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		500				{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/metrics/algorithms [get]
//	@x-api-type		{"sdk":"true"}
func GetAlgorithmMetrics(c *gin.Context) {
	var req dto.GetMetricsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	metrics, err := producer.GetAlgorithmMetrics(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Algorithm metrics retrieved successfully", metrics)
}
