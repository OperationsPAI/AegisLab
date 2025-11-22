package v2

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/handlers"
	"aegis/middleware"
	producer "aegis/service/prodcuer"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// BatchDeleteExecutions handles batch deletion of executions
//
//	@Summary		Batch delete executions
//	@Description	Batch delete executions by IDs or labels with cascading deletion of related records
//	@Tags			Executions
//	@ID				batch_delete_executions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.BatchDeleteExecutionReq	true	"Batch delete request"
//	@Success		200		{object}	dto.GenericResponse[any]	"Executions deleted successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/executions/batch-delete [post]
func BatchDeleteExecutions(c *gin.Context) {
	var req dto.BatchDeleteExecutionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	var err error
	if len(req.IDs) > 0 {
		err = producer.BatchDeleteExecutionsByIDs(req.IDs)
	} else {
		err = producer.BatchDeleteExecutionsByLabels(req.Labels)
	}

	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Executions deleted successfully", nil)
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
//	@Success		200	{object}	dto.GenericResponse[dto.ExecutionDetailResp]	"Execution retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]						"Invalid execution ID"
//	@Failure		401	{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]						"Execution not found"
//	@Failure		500	{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/executions/{id} [get]
//	@x-api-type		{"sdk":"true"}
func GetExecution(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid execution ID")
		return
	}

	resp, err := producer.GetExecutionDetail(id)
	if handlers.HandleServiceError(c, err) {
		return
	}

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
//	@Success		200		{object}	dto.GenericResponse[dto.ListResp[dto.ExecutionResp]]	"Executions retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]								"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/executions [get]
//	@x-api-type		{"sdk":"true"}
func ListExecutions(c *gin.Context) {
	var req dto.ListExecutionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListExecutions(&req)
	if handlers.HandleServiceError(c, err) {
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
//	@x-api-type		{"sdk":"true"}
func ListAvaliableExecutionLabels(c *gin.Context) {
	labels, err := producer.ListAvaliableExecutionLabels()
	if handlers.HandleServiceError(c, err) {
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
//	@Param			manage	body		dto.ManageExecutionLabelReq				true	"Custom label management request"
//	@Success		200		{object}	dto.GenericResponse[dto.ExecutionResp]	"Custom labels managed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid execution ID or request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]				"Execution not found"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/executions/{id}/labels [patch]
func ManageExecutionCustomLabels(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid execution ID")
		return
	}

	var req dto.ManageExecutionLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ManageExecutionLabels(&req, id)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			request	body		dto.SubmitExecutionReq							true	"Algorithm execution request"
//	@Success		200		{object}	dto.GenericResponse[dto.SubmitExecutionResp]	"Algorithm execution submitted successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]						"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]						"Project, algorithm, datapack or dataset not found"
//	@Failure		500		{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/executions/execute [post]
//	@x-api-type		{"sdk":"true"}
func SubmitAlgorithmExecution(c *gin.Context) {
	groupID := c.GetString("groupID")
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		logrus.Error("failed to get span context from gin.Context")
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get span context")
		return
	}

	spanCtx := ctx.(context.Context)
	span := trace.SpanFromContext(spanCtx)

	var req dto.SubmitExecutionReq
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

	resp, err := producer.ProduceAlgorithmExeuctionTasks(spanCtx, &req, groupID, userID)
	if err != nil {
		span.SetStatus(codes.Error, "service error in SubmitAlgorithmExecution: "+err.Error())
		logrus.Errorf("Failed to submit algorithm execution: %v", err)
		handlers.HandleServiceError(c, err)
		return
	}

	span.SetStatus(codes.Ok, "Successfully submitted algorithm execution")
	dto.SuccessResponse(c, resp)
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
//	@Param			request			body		dto.UploadDetectorResultReq							true	"Detector results"
//	@Success		200				{object}	dto.GenericResponse[dto.UploadExecutionResultResp]	"Results uploaded successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]							"Invalid executionID or invalid request format or parameters"
//	@Failure		401				{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]							"Execution not found"
//	@Failure		500				{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/executions/{execution_id}/detector_results [post]
//	@x-api-type		{"sdk":"true"}
func UploadDetectorResults(c *gin.Context) {
	executionIDStr := c.Param(consts.URLPathExecutionID)
	executionID, err := strconv.Atoi(executionIDStr)
	if err != nil || executionID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid execution ID")
		return
	}

	var req dto.UploadDetectorResultReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.BatchCreateDetectorResults(&req, executionID)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			request			body		dto.UploadGranularityResultReq						true	"Granularity results"
//	@Success		200				{object}	dto.GenericResponse[dto.UploadExecutionResultResp]	"Results uploaded successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]							"Invalid exeuction ID or invalid request form or parameters"
//	@Failure		401				{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]							"Execution not found"
//	@Failure		500				{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/executions/{execution_id}/granularity_results [post]
//	@x-api-type		{"sdk":"true"}
func UploadGranularityResults(c *gin.Context) {
	executionIDStr := c.Param(consts.URLPathExecutionID)
	executionID, err := strconv.Atoi(executionIDStr)
	if err != nil || executionID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid execution ID")
		return
	}

	var req dto.UploadGranularityResultReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.BatchCreateGranularityResults(&req, executionID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}
