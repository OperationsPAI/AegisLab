package v2

import (
	"aegis/consts"
	"context"
	"fmt"
	"net/http"

	"aegis/dto"
	"aegis/handlers"
	"aegis/middleware"
	producer "aegis/service/prodcuer"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
)

// BatchDeleteInjections
//
//	@Summary		Batch delete injections
//	@Description	Batch delete injections by IDs or labels or tags with cascading deletion of related records
//	@Tags			Injections
//	@ID				batch_delete_injections
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			batch_delete	body		dto.BatchDeleteInjectionReq	true	"Batch delete request"
//	@Success		200				{object}	dto.GenericResponse[any]	"Injections deleted successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]	"Invalid request"
//	@Failure		401				{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		500				{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/injections/batch-delete [post]
func BatchDeleteInjections(c *gin.Context) {
	var req dto.BatchDeleteInjectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	var err error
	if len(req.IDs) > 0 {
		err = producer.BatchDeleteInjectionsByIDs(req.IDs)
	} else {
		err = producer.BatchDeleteInjectionsByLabels(req.Labels)
	}

	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Injections deleted successfully", nil)
}

// GetInjection handles getting a single injection by ID
//
//	@Summary		Get injection by ID
//	@Description	Get detailed information about a specific injection
//	@Tags			Injections
//	@ID				get_injection_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int												true	"Injection ID"
//	@Success		200	{object}	dto.GenericResponse[dto.InjectionDetailResp]	"Injection retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]						"Invalid injection ID"
//	@Failure		401	{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]						"Injection not found"
//	@Failure		500	{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/injections/{id} [get]
//	@x-api-type		{"sdk":"true"}
func GetInjection(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	logrus.WithFields(logrus.Fields{
		"idStr": idStr,
		"path":  c.Request.URL.Path,
	}).Info("GetInjection: received request")

	id, ok := handlers.ParsePositiveID(c, idStr, "injection ID")
	if !ok {
		logrus.WithField("idStr", idStr).Warn("GetInjection: invalid ID format or ID <= 0")
		return
	}

	logrus.WithField("id", id).Info("GetInjection: calling GetInjectionDetail")
	resp, err := producer.GetInjectionDetail(id)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    id,
			"error": err.Error(),
		}).Error("GetInjection: failed to get injection detail")
	}

	if handlers.HandleServiceError(c, err) {
		return
	}

	logrus.WithField("id", id).Info("GetInjection: successfully retrieved injection")
	dto.SuccessResponse(c, resp)
}

// GetInjectionMetadata
//
//	@Summary		Get Injection Metadata
//	@Description	Get injection-related metadata including configuration, field mappings, and system resources
//	@Tags			Injections
//	@ID				get_injection_metadata
//	@Produce		json
//	@Param			system	query		chaos.SystemType								true	"System for config and resources metadata"
//	@Success		200		{object}	dto.GenericResponse[dto.InjectionMetadataResp]	"Successfully returned metadata"
//	@Failure		400		{object}	dto.GenericResponse[any]						"Invalid system"
//	@Failure		401		{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]						"Resource not found"
//	@Failure		500		{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/injections/metadata [get]
//	@x-api-type		{"sdk":"true"}
func GetInjectionMetadata(c *gin.Context) {
	systemStr := c.Query("system")

	ctx := context.Background()
	system := chaos.SystemType(systemStr)

	confNode, err := chaos.StructToNode[chaos.InjectionConf](ctx, system)
	if err != nil {
		handlers.HandleServiceError(c, err)
		return
	}

	faultResourceMap, err := chaos.GetChaosTypeResourceMappings()
	if err != nil {
		handlers.HandleServiceError(c, err)
		return
	}

	resourceMap, err := chaos.GetSystemResourceMap(ctx)
	if err != nil {
		handlers.HandleServiceError(c, err)
		return
	}

	resource, exists := resourceMap[system]
	if !exists {
		dto.ErrorResponse(c, http.StatusNotFound, "Namespace resources not found")
		return
	}

	dto.SuccessResponse(c, &dto.InjectionMetadataResp{
		Config:           confNode,
		FaultTypeMap:     chaos.ChaosTypeMap,
		FaultResourceMap: faultResourceMap,
		SystemResource:   resource,
	})
}

// ListInjections handles listing injections with pagination and filtering
//
//	@Summary		List injections
//	@Description	Get a paginated list of injections with pagination and filtering
//	@Tags			Injections
//	@ID				list_injections
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int														false	"Page number"	default(1)
//	@Param			size		query		int														false	"Page size"		default(20)
//	@Param			type		query		chaos.ChaosType											false	"Filter by fault type"
//	@Param			benchmark	query		string													false	"Filter by benchmark"
//	@Param			state		query		consts.DatapackState									false	"Filter by injection state"
//	@Param			status		query		int														false	"Filter by status"
//	@Param			labels		query		[]string												false	"Filter by labels (array of key:value strings, e.g., 'type:chaos')"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[dto.InjectionResp]]	"Injections retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]								"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/injections [get]
//	@x-api-type		{"sdk":"true"}
func ListInjections(c *gin.Context) {
	var req dto.ListInjectionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListInjections(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// SearchInjections
//
//	@Summary		Search injections
//	@Description	Advanced search for injections with complex filtering including name search, custom labels, tags, and time ranges
//	@Tags			Injections
//	@ID				search_injections
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			search	body		dto.SearchInjectionReq											true	"Search criteria"
//	@Success		200		{object}	dto.GenericResponse[dto.SearchResp[dto.InjectionDetailResp]]	"Search results"
//	@Failure		400		{object}	dto.GenericResponse[any]										"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]										"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]										"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]										"Internal server error"
//	@Router			/api/v2/injections/search [post]
//	@x-api-type		{"sdk":"true"}
func SearchInjections(c *gin.Context) {
	var req dto.SearchInjectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := producer.SearchInjections(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// ListFaultInjectionNoIssues
//
//	@Summary		Query Fault Injection Records Without Issues
//	@Description	Query all fault injection records without issues based on time range, returning detailed records including configuration information
//	@Tags			Injections
//	@ID				list_failed_injections
//	@Produce		json
//	@Param			labels				query		[]string											false	"Filter by labels (array of key:value strings, e.g., 'type:chaos')"
//	@Param			lookback			query		string												false	"Time range query, supports custom relative time (1h/24h/7d) or custom, default not set"
//	@Param			custom_start_time	query		string												false	"Custom start time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Param			custom_end_time		query		string												false	"Custom end time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Success		200					{object}	dto.GenericResponse[[]dto.InjectionNoIssuesResp]	"Successfully returned fault injection records without issues"
//	@Failure		400					{object}	dto.GenericResponse[any]							"Request parameter error, such as incorrect time format or parameter validation failure, etc."
//	@Failure		500					{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/injections/analysis/no-issues [get]
//	@x-api-type		{"sdk":"true"}
func ListFaultInjectionNoIssues(c *gin.Context) {
	var req dto.ListInjectionNoIssuesReq
	if err := c.BindQuery(&req); err != nil {
		logrus.Errorf("failed to bind query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if err := req.Validate(); err != nil {
		logrus.Errorf("invalid query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	items, err := producer.ListInjectionsNoIssues(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, items)
}

// ListFaultInjectionWithIssues
//
//	@Summary		Query Fault Injection Records With Issues
//	@Description	Query all fault injection records with issues based on time range
//	@Tags			Injections
//	@ID				list_successful_injections
//	@Produce		json
//	@Param			labels				query		[]string	false	"Filter by labels (array of key:value strings, e.g., 'type:chaos')"
//	@Param			lookback			query		string		false	"Time range query, supports custom relative time (1h/24h/7d) or custom, default not set"
//	@Param			custom_start_time	query		string		false	"Custom start time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Param			custom_end_time		query		string		false	"Custom end time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Success		200					{object}	dto.GenericResponse[[]dto.InjectionWithIssuesResp]
//	@Failure		400					{object}	dto.GenericResponse[any]	"Request parameter error, such as incorrect time format or parameter validation failure, etc."
//	@Failure		500					{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/injections/analysis/with-issues [get]
//	@x-api-type		{"sdk":"true"}
func ListFaultInjectionWithIssues(c *gin.Context) {
	var req dto.ListInjectionWithIssuesReq
	if err := c.BindQuery(&req); err != nil {
		logrus.Errorf("failed to bind query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if err := req.Validate(); err != nil {
		logrus.Errorf("invalid query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	items, err := producer.ListInjectionsWithIssues(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, items)
}

// ManageInjectionCustomLabels manages injection custom labels (key-value pairs)
//
//	@Summary		Manage injection custom labels
//	@Description	Add or remove custom labels (key-value pairs) for an injection
//	@Tags			Injections
//	@ID				manage_injection_labels
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int										true	"Injection ID"
//	@Param			manage	body		dto.ManageInjectionLabelReq				true	"Custom label management request"
//	@Success		200		{object}	dto.GenericResponse[dto.InjectionResp]	"Custom labels managed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid injection ID or request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]				"Injection not found"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/injections/{id}/labels [patch]
//	@x-api-type		{"sdk":"true"}
func ManageInjectionCustomLabels(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, ok := handlers.ParsePositiveID(c, idStr, "injection ID")
	if !ok {
		return
	}

	var req dto.ManageInjectionLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ManageInjectionLabels(&req, id)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// BatchManageInjectionLabels
//
//	@Summary		Batch manage injection labels
//	@Description	Add or remove labels from multiple injections by IDs with success/failure tracking
//	@Tags			Injections
//	@ID				batch_manage_injection_labels
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			batch_manage	body		dto.BatchManageInjectionLabelReq						true	"Batch manage label request"
//	@Success		200				{object}	dto.GenericResponse[dto.BatchManageInjectionLabelResp]	"Injection labels managed successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]								"Invalid request"
//	@Failure		401				{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		500				{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/injections/labels/batch [patch]
//	@x-api-type		{"sdk":"true"}
func BatchManageInjectionLabels(c *gin.Context) {
	var req dto.BatchManageInjectionLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := producer.BatchManageInjectionLabels(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// SubmitFaultInjection submits batch fault injections
//
//	@Summary		Submit batch fault injections
//	@Description	Submit multiple fault injection tasks in batch
//	@Tags			Injections
//	@ID				inject_fault
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.SubmitInjectionReq							true	"Fault injection request body"
//	@Success		200		{object}	dto.GenericResponse[dto.SubmitInjectionResp]	"Fault injection submitted successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]						"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]						"Resource not found"
//	@Failure		500		{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/injections/inject [post]
//	@x-api-type		{"sdk":"true"}
func SubmitFaultInjection(c *gin.Context) {
	groupID := c.GetString("groupID")
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		logrus.Error("Failed to get span context from gin.Context in SubmitFaultInjection")
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get span context")
		return
	}

	spanCtx := ctx.(context.Context)
	span := trace.SpanFromContext(spanCtx)

	var req dto.SubmitInjectionReq
	if err := c.BindJSON(&req); err != nil {
		span.SetStatus(codes.Error, "validation error in SubmitFaultInjection: "+err.Error())
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		span.SetStatus(codes.Error, "validation error in SubmitFaultInjection: "+err.Error())
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ProduceRestartPedestalTasks(spanCtx, &req, groupID, userID)
	if err != nil {
		span.SetStatus(codes.Error, "service error in SubmitFaultInjection: "+err.Error())
		logrus.Errorf("Failed to submit fault injection: %v", err)
		handlers.HandleServiceError(c, err)
		return
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("Successfully submitted %d fault injections with groupID: %s", len(resp.Items), groupID))
	dto.SuccessResponse(c, resp)
}

// SubmitDatapackBuilding submits batch datapack buildings
//
//	@Summary		Submit batch datapack buildings
//	@Description.	Submit multiple datapack building tasks in batch
//	@Tags			Injections
//	@ID				build_datapack
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.SubmitDatapackBuildingReq						true	"Datapack building request body"
//	@Success		202		{object}	dto.GenericResponse[dto.SubmitDatapackBuildingResp]	"Datapack building submitted successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]							"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]							"Resource not found"
//	@Failure		500		{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/injections/build [post]
//	@x-api-type		{"sdk":"true"}
func SubmitDatapackBuilding(c *gin.Context) {
	groupID := c.GetString("groupID")
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		logrus.Error("Failed to get span context from gin.Context in SubmitDatapackBuilding")
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get span context")
		return
	}

	spanCtx := ctx.(context.Context)
	span := trace.SpanFromContext(spanCtx)

	var req dto.SubmitDatapackBuildingReq
	if err := c.BindJSON(&req); err != nil {
		span.SetStatus(codes.Error, "validation error in SubmitDatapackBuilding: "+err.Error())
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		span.SetStatus(codes.Error, "validation error in SubmitDatapackBuilding: "+err.Error())
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ProduceDatapackBuildingTasks(spanCtx, &req, groupID, userID)
	if err != nil {
		span.SetStatus(codes.Error, "service error in SubmitDatapackBuilding: "+err.Error())
		logrus.Errorf("Failed to submit datapack building: %v", err)
		handlers.HandleServiceError(c, err)
		return
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("Successfully submitted %d datapack buildings with groupID: %s", len(resp.Items), groupID))
	dto.SuccessResponse(c, resp)
}
