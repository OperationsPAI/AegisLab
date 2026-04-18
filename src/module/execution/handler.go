package executionmodule

import (
	"aegis/httpx"
	"context"
	"net/http"
	"strconv"

	"aegis/consts"
	"aegis/dto"
	"aegis/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	service HandlerService
}

func NewHandler(service HandlerService) *Handler {
	return &Handler{service: service}
}

// ListProjectExecutions lists all algorithm executions for a project
//
//	@Summary		List project executions
//	@Description	Get paginated list of algorithm executions for a specific project
//	@Tags			Projects
//	@ID				list_project_executions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			project_id	path		int														true	"Project ID"
//	@Param			page		query		int														false	"Page number"	default(1)
//	@Param			size		query		int														false	"Page size"		default(20)
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[ExecutionResp]]	"Executions retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]								"Invalid project ID or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]								"Project not found"
//	@Failure		500			{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/projects/{project_id}/executions [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) ListProjectExecutions(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req ListExecutionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.ListProjectExecutions(c.Request.Context(), &req, projectID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// SubmitAlgorithmExecution submits batch algorithm execution for multiple datapacks or datasets
//
//	@Summary		Submit batch algorithm execution
//	@Description	Submit multiple algorithm execution tasks in batch. Supports mixing datapack (v1 compatible) and dataset (v2 feature) executions.
//	@Tags			Executions
//	@ID				run_algorithm
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			project_id	path		int												true	"Project ID"
//	@Param			request	body		SubmitExecutionReq							true	"Algorithm execution request"
//	@Success		200		{object}	dto.GenericResponse[SubmitExecutionResp]	"Algorithm execution submitted successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]						"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]						"Project, algorithm, datapack or dataset not found"
//	@Failure		500		{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/projects/{project_id}/executions/execute [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) SubmitAlgorithmExecution(c *gin.Context) {
	groupID := c.GetString("groupID")
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	spanCtx, span, ok := spanFromGin(c)
	if !ok {
		return
	}

	var req SubmitExecutionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		span.SetStatus(codes.Error, "validation error in SubmitAlgorithmExecution: "+err.Error())
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		span.SetStatus(codes.Error, "validation error in SubmitAlgorithmExecution: "+err.Error())
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.SubmitAlgorithmExecution(spanCtx, &req, groupID, userID)
	if err != nil {
		span.SetStatus(codes.Error, "service error in SubmitAlgorithmExecution: "+err.Error())
		logrus.Errorf("Failed to submit algorithm execution: %v", err)
		httpx.HandleServiceError(c, err)
		return
	}

	span.SetStatus(codes.Ok, "Successfully submitted algorithm execution")
	dto.SuccessResponse(c, resp)
}

// ListExecutions handles listing executions with pagination and filtering
//
//	@Summary		List executions
//	@Description	Get a paginated list of executions with pagination and filtering
//	@Tags			Executions
//	@ID				list_executions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int														false	"Page number"	default(1)
//	@Param			size	query		int														false	"Page size"		default(20)
//	@Param			state	query		consts.ExecutionState									false	"Filter by execution state"
//	@Param			status	query		consts.StatusType										false	"Filter by status"
//	@Param			labels	query		[]string												false	"Filter by labels (array of key:value strings, e.g., 'type:test')"
//	@Success		200		{object}	dto.GenericResponse[dto.ListResp[ExecutionResp]]	"Executions retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]								"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/executions [get]
//	@x-api-type		{}
func (h *Handler) ListExecutions(c *gin.Context) {
	var req ListExecutionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.ListExecutions(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// GetExecution handles getting a single execution by ID
//
//	@Summary		Get execution by ID
//	@Description	Get detailed information about a specific execution
//	@Tags			Executions
//	@ID				get_execution_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int												true	"Execution ID"
//	@Success		200	{object}	dto.GenericResponse[ExecutionDetailResp]	"Execution retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]						"Invalid execution ID"
//	@Failure		401	{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]						"Execution not found"
//	@Failure		500	{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/executions/{id} [get]
//	@x-api-type		{}
func (h *Handler) GetExecution(c *gin.Context) {
	id, ok := httpx.ParsePositiveID(c, c.Param(consts.URLPathExecutionID), "execution ID")
	if !ok {
		return
	}
	resp, err := h.service.GetExecution(c.Request.Context(), id)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// ListExecutionLabels handles listing available execution labels
//
//	@Summary		List execution labels
//	@Description	List all available label keys for executions
//	@Tags			Executions
//	@ID				list_execution_labels
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	dto.GenericResponse[[]dto.LabelItem]	"Available label keys"
//	@Failure		401	{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		500	{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/executions/labels [get]
//	@x-api-type		{}
func (h *Handler) ListAvailableExecutionLabels(c *gin.Context) {
	labels, err := h.service.ListAvailableLabels(c.Request.Context())
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, labels)
}

// ManageExecutionCustomLabels manages execution custom labels (key-value pairs)
//
//	@Summary		Manage execution custom labels
//	@Description	Add or remove custom labels (key-value pairs) for an execution
//	@Tags			Executions
//	@ID				update_execution_labels
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int										true	"Execution ID"
//	@Param			manage	body		ManageExecutionLabelReq				true	"Custom label management request"
//	@Success		200		{object}	dto.GenericResponse[ExecutionResp]	"Custom labels managed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid execution ID or request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]				"Execution not found"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/executions/{id}/labels [patch]
//	@x-api-type		{}
func (h *Handler) ManageExecutionCustomLabels(c *gin.Context) {
	id, ok := httpx.ParsePositiveID(c, c.Param(consts.URLPathExecutionID), "execution ID")
	if !ok {
		return
	}
	var req ManageExecutionLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.ManageLabels(c.Request.Context(), &req, id)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// BatchDeleteExecutions handles batch deletion of executions
//
//	@Summary		Batch delete executions
//	@Description	Batch delete executions by IDs or labels with cascading deletion of related records
//	@Tags			Executions
//	@ID				batch_delete_executions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		BatchDeleteExecutionReq	true	"Batch delete request"
//	@Success		200		{object}	dto.GenericResponse[any]	"Executions deleted successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/executions/batch-delete [post]
//	@x-api-type		{}
func (h *Handler) BatchDeleteExecutions(c *gin.Context) {
	var req BatchDeleteExecutionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	if httpx.HandleServiceError(c, h.service.BatchDelete(c.Request.Context(), &req)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusNoContent, "Executions deleted successfully", nil)
}

// UploadDetectorResults uploads detector results
//
//	@Summary		Upload detector results
//	@Description	Upload detection results for detector algorithm via API instead of file collection
//	@Tags			Executions
//	@ID				upload_detection_results
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			execution_id	path		int													true	"Execution ID"
//	@Param			request			body		UploadDetectorResultReq								true	"Detector results"
//	@Success		200				{object}	dto.GenericResponse[UploadExecutionResultResp]		"Results uploaded successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]							"Invalid executionID or invalid request format or parameters"
//	@Failure		401				{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]							"Execution not found"
//	@Failure		500				{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/executions/{execution_id}/detector_results [post]
//	@x-api-type		{"runtime":"true"}
func (h *Handler) UploadDetectorResults(c *gin.Context) {
	executionID, ok := httpx.ParsePositiveID(c, c.Param(consts.URLPathExecutionID), "execution ID")
	if !ok {
		return
	}
	var req UploadDetectorResultReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.UploadDetectorResults(c.Request.Context(), &req, executionID)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// UploadGranularityResults uploads granularity results
//
//	@Summary		Upload granularity results
//	@Description	Upload granularity results for regular algorithms via API instead of file collection
//	@Tags			Executions
//	@ID				upload_localization_results
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			execution_id	path		int													true	"Execution ID"
//	@Param			request			body		UploadGranularityResultReq							true	"Granularity results"
//	@Success		200				{object}	dto.GenericResponse[UploadExecutionResultResp]		"Results uploaded successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]							"Invalid exeuction ID or invalid request form or parameters"
//	@Failure		401				{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]							"Execution not found"
//	@Failure		500				{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/executions/{execution_id}/granularity_results [post]
//	@x-api-type		{"runtime":"true"}
func (h *Handler) UploadGranularityResults(c *gin.Context) {
	executionID, ok := httpx.ParsePositiveID(c, c.Param(consts.URLPathExecutionID), "execution ID")
	if !ok {
		return
	}
	var req UploadGranularityResultReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.UploadGranularityResults(c.Request.Context(), &req, executionID)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

func spanFromGin(c *gin.Context) (context.Context, trace.Span, bool) {
	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		logrus.Error("failed to get span context from gin.Context")
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get span context")
		return nil, nil, false
	}

	spanCtx := ctx.(context.Context)
	return spanCtx, trace.SpanFromContext(spanCtx), true
}

func parseProjectID(c *gin.Context) (int, bool) {
	projectIDStr := c.Param(consts.URLPathProjectID)
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil || projectID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return 0, false
	}
	return projectID, true
}
