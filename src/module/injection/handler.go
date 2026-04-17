package injectionmodule

import (
	"aegis/httpx"
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"aegis/consts"
	"aegis/dto"
	"aegis/middleware"
	"aegis/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListProjectInjections lists all fault injections for a project
//
//	@Summary		List project fault injections
//	@Description	Get paginated list of fault injections for a specific project
//	@Tags			Projects
//	@ID				list_project_injections
//	@Produce		json
//	@Security		BearerAuth
//	@Param			project_id	path		int														true	"Project ID"
//	@Param			page		query		int														false	"Page number"	default(1)
//	@Param			size		query		int														false	"Page size"		default(20)
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[InjectionResp]]	"Fault injections retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]								"Invalid project ID or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]								"Project not found"
//	@Failure		500			{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/projects/{project_id}/injections [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) ListProjectInjections(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req ListInjectionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.ListProjectInjections(c.Request.Context(), &req, projectID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// SearchProjectInjections searches fault injections within a specific project
//
//	@Summary		Search project fault injections
//	@Description	Advanced search for injections within a project with complex filtering
//	@Tags			Projects
//	@ID				search_project_injections
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			project_id	path		int																true	"Project ID"
//	@Param			search		body		SearchInjectionReq											true	"Search criteria"
//	@Success		200			{object}	dto.GenericResponse[dto.SearchResp[InjectionDetailResp]]	"Search results"
//	@Failure		400			{object}	dto.GenericResponse[any]										"Invalid project ID or request"
//	@Failure		401			{object}	dto.GenericResponse[any]										"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]										"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]										"Project not found"
//	@Failure		500			{object}	dto.GenericResponse[any]										"Internal server error"
//	@Router			/api/v2/projects/{project_id}/injections/search [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) SearchProjectInjections(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	h.searchInjections(c, &projectID)
}

// ListProjectFaultInjectionNoIssues lists fault injections without issues for a project
//
//	@Summary		List project fault injections without issues
//	@Description	Query fault injection records without issues within a project based on time range
//	@Tags			Projects
//	@ID				list_project_injections_no_issues
//	@Produce		json
//	@Security		BearerAuth
//	@Param			project_id			path		int													true	"Project ID"
//	@Param			labels				query		[]string											false	"Filter by labels"
//	@Param			lookback			query		string												false	"Time range query"
//	@Param			custom_start_time	query		string												false	"Custom start time"
//	@Param			custom_end_time		query		string												false	"Custom end time"
//	@Success		200					{object}	dto.GenericResponse[[]InjectionNoIssuesResp]	"Injections retrieved successfully"
//	@Failure		400					{object}	dto.GenericResponse[any]							"Invalid parameters"
//	@Failure		401					{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403					{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404					{object}	dto.GenericResponse[any]							"Project not found"
//	@Failure		500					{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/projects/{project_id}/injections/analysis/no-issues [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) ListProjectFaultInjectionNoIssues(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	h.listFaultInjectionNoIssues(c, &projectID)
}

// ListProjectFaultInjectionWithIssues lists fault injections with issues for a project
//
//	@Summary		List project fault injections with issues
//	@Description	Query fault injection records with issues within a project based on time range
//	@Tags			Projects
//	@ID				list_project_injections_with_issues
//	@Produce		json
//	@Security		BearerAuth
//	@Param			project_id			path		int													true	"Project ID"
//	@Param			labels				query		[]string											false	"Filter by labels"
//	@Param			lookback			query		string												false	"Time range query"
//	@Param			custom_start_time	query		string												false	"Custom start time"
//	@Param			custom_end_time		query		string												false	"Custom end time"
//	@Success		200					{object}	dto.GenericResponse[[]InjectionWithIssuesResp]	"Injections retrieved successfully"
//	@Failure		400					{object}	dto.GenericResponse[any]							"Invalid parameters"
//	@Failure		401					{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403					{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404					{object}	dto.GenericResponse[any]							"Project not found"
//	@Failure		500					{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/projects/{project_id}/injections/analysis/with-issues [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) ListProjectFaultInjectionWithIssues(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	h.listFaultInjectionWithIssues(c, &projectID)
}

// SubmitProjectFaultInjection submits fault injections for a specific project
//
//	@Summary		Submit project fault injections
//	@Description	Submit multiple fault injection tasks for a specific project
//	@Tags			Projects
//	@ID				submit_project_fault_injection
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			project_id	path		int												true	"Project ID"
//	@Param			body		body		SubmitInjectionReq							true	"Fault injection request"
//	@Success		200			{object}	dto.GenericResponse[SubmitInjectionResp]	"Injections submitted successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]						"Invalid request"
//	@Failure		401			{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]						"Project not found"
//	@Failure		500			{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/projects/{project_id}/injections/inject [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) SubmitProjectFaultInjection(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	h.submitFaultInjection(c, &projectID)
}

// SubmitProjectDatapackBuilding submits datapack building tasks for a specific project
//
//	@Summary		Submit project datapack buildings
//	@Description	Submit multiple datapack building tasks for a specific project
//	@Tags			Projects
//	@ID				submit_project_datapack_building
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			project_id	path		int													true	"Project ID"
//	@Param			body		body		SubmitDatapackBuildingReq						true	"Datapack building request"
//	@Success		202			{object}	dto.GenericResponse[SubmitDatapackBuildingResp]	"Datapack buildings submitted successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]							"Invalid request"
//	@Failure		401			{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]							"Project not found"
//	@Failure		500			{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/projects/{project_id}/injections/build [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) SubmitProjectDatapackBuilding(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	h.submitDatapackBuilding(c, &projectID)
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
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[InjectionResp]]	"Injections retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]								"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/injections [get]
//	@x-api-type		{}
func (h *Handler) ListInjections(c *gin.Context) {
	var req ListInjectionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.ListInjections(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
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
//	@Param			search	body		SearchInjectionReq											true	"Search criteria"
//	@Success		200		{object}	dto.GenericResponse[dto.SearchResp[InjectionDetailResp]]	"Search results"
//	@Failure		400		{object}	dto.GenericResponse[any]										"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]										"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]										"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]										"Internal server error"
//	@Router			/api/v2/injections/search [post]
//	@x-api-type		{}
func (h *Handler) SearchInjections(c *gin.Context) { h.searchInjections(c, nil) }

// SubmitFaultInjection submits batch fault injections
//
//	@Summary		Submit batch fault injections
//	@Description	Submit multiple fault injection tasks in batch
//	@Tags			Injections
//	@ID				inject_fault
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		SubmitInjectionReq							true	"Fault injection request body"
//	@Success		200		{object}	dto.GenericResponse[SubmitInjectionResp]	"Fault injection submitted successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]						"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]						"Resource not found"
//	@Failure		500		{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/injections/inject [post]
//	@x-api-type		{}
func (h *Handler) SubmitFaultInjection(c *gin.Context) { h.submitFaultInjection(c, nil) }

// SubmitDatapackBuilding submits batch datapack buildings
//
//	@Summary		Submit batch datapack buildings
//	@Description.	Submit multiple datapack building tasks in batch
//	@Tags			Injections
//	@ID				build_datapack
//	@Accept			json
//	@Produce		json
//	@Param			body	body		SubmitDatapackBuildingReq						true	"Datapack building request body"
//	@Success		202		{object}	dto.GenericResponse[SubmitDatapackBuildingResp]	"Datapack building submitted successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]							"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]							"Resource not found"
//	@Failure		500		{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/injections/build [post]
//	@x-api-type		{}
func (h *Handler) SubmitDatapackBuilding(c *gin.Context) { h.submitDatapackBuilding(c, nil) }

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
//	@Success		200					{object}	dto.GenericResponse[[]InjectionNoIssuesResp]	"Successfully returned fault injection records without issues"
//	@Failure		400					{object}	dto.GenericResponse[any]							"Request parameter error, such as incorrect time format or parameter validation failure, etc."
//	@Failure		500					{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/injections/analysis/no-issues [get]
//	@x-api-type		{}
func (h *Handler) ListFaultInjectionNoIssues(c *gin.Context) { h.listFaultInjectionNoIssues(c, nil) }

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
//	@Success		200					{object}	dto.GenericResponse[[]InjectionWithIssuesResp]
//	@Failure		400					{object}	dto.GenericResponse[any]	"Request parameter error, such as incorrect time format or parameter validation failure, etc."
//	@Failure		500					{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/injections/analysis/with-issues [get]
//	@x-api-type		{}
func (h *Handler) ListFaultInjectionWithIssues(c *gin.Context) {
	h.listFaultInjectionWithIssues(c, nil)
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
//	@Success		200	{object}	dto.GenericResponse[InjectionDetailResp]	"Injection retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]						"Invalid injection ID"
//	@Failure		401	{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]						"Injection not found"
//	@Failure		500	{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/injections/{id} [get]
//	@x-api-type		{}
func (h *Handler) GetInjection(c *gin.Context) {
	id, ok := parsePositiveID(c, consts.URLPathID, "injection ID")
	if !ok {
		return
	}
	resp, err := h.service.GetInjection(c.Request.Context(), id)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// GetInjectionMetadata
//
//	@Summary		Get Injection Metadata
//	@Description	Get injection-related metadata including configuration, field mappings, and system resources
//	@Tags			Injections
//	@ID				get_injection_metadata
//	@Produce		json
//	@Security		BearerAuth
//	@Param			system	query		chaos.SystemType								true	"System for config and resources metadata"
//	@Success		200		{object}	dto.GenericResponse[InjectionMetadataResp]	"Successfully returned metadata"
//	@Failure		400		{object}	dto.GenericResponse[any]						"Invalid system"
//	@Failure		401		{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]						"Resource not found"
//	@Failure		500		{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/injections/metadata [get]
//	@x-api-type		{}
func (h *Handler) GetInjectionMetadata(c *gin.Context) {
	systemStr := c.Query("system")
	ctx := c.Request.Context()
	system := chaos.SystemType(systemStr)

	confNode, err := chaos.StructToNode[chaos.InjectionConf](string(system))
	if err != nil {
		logrus.Warnf("Failed to build injection config node: %v, continuing with nil config", err)
	}

	faultResourceMap, err := chaos.GetChaosTypeResourceMappings()
	if err != nil {
		httpx.HandleServiceError(c, err)
		return
	}

	resourceMap, err := chaos.GetSystemResourceMap(ctx)
	if err != nil {
		logrus.Warnf("Failed to get system resource map: %v, using empty map", err)
		resourceMap = make(map[chaos.SystemType]chaos.SystemResource)
	}

	dto.SuccessResponse(c, &InjectionMetadataResp{
		Config:           confNode,
		FaultTypeMap:     chaos.ChaosTypeMap,
		FaultResourceMap: faultResourceMap,
		SystemResource:   resourceMap[system],
	})
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
//	@Param			manage	body		ManageInjectionLabelReq				true	"Custom label management request"
//	@Success		200		{object}	dto.GenericResponse[InjectionResp]	"Custom labels managed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid injection ID or request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]				"Injection not found"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/injections/{id}/labels [patch]
//	@x-api-type		{}
func (h *Handler) ManageInjectionCustomLabels(c *gin.Context) {
	id, ok := parsePositiveID(c, consts.URLPathID, "injection ID")
	if !ok {
		return
	}
	var req ManageInjectionLabelReq
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

// BatchManageInjectionLabels
//
//	@Summary		Batch manage injection labels
//	@Description	Add or remove labels from multiple injections by IDs with success/failure tracking
//	@Tags			Injections
//	@ID				batch_manage_injection_labels
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			batch_manage	body		BatchManageInjectionLabelReq						true	"Batch manage label request"
//	@Success		200				{object}	dto.GenericResponse[BatchManageInjectionLabelResp]	"Injection labels managed successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]								"Invalid request"
//	@Failure		401				{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		500				{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/injections/labels/batch [patch]
//	@x-api-type		{}
func (h *Handler) BatchManageInjectionLabels(c *gin.Context) {
	var req BatchManageInjectionLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}
	resp, err := h.service.BatchManageLabels(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// BatchDeleteInjections
//
//	@Summary		Batch delete injections
//	@Description	Batch delete injections by IDs or labels or tags with cascading deletion of related records
//	@Tags			Injections
//	@ID				batch_delete_injections
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			batch_delete	body		BatchDeleteInjectionReq	true	"Batch delete request"
//	@Success		200				{object}	dto.GenericResponse[any]	"Injections deleted successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]	"Invalid request"
//	@Failure		401				{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		500				{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/injections/batch-delete [post]
//	@x-api-type		{}
func (h *Handler) BatchDeleteInjections(c *gin.Context) {
	var req BatchDeleteInjectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}
	if httpx.HandleServiceError(c, h.service.BatchDelete(c.Request.Context(), &req)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusNoContent, "Injections deleted successfully", nil)
}

// CloneInjection handles cloning an injection configuration
//
//	@Summary		Clone injection
//	@Description	Clone an existing injection configuration for reuse
//	@Tags			Injections
//	@ID				clone_injection
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int												true	"Injection ID"
//	@Param			body	body		CloneInjectionReq							true	"Clone request"
//	@Success		201		{object}	dto.GenericResponse[InjectionDetailResp]	"Injection cloned successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]						"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		404		{object}	dto.GenericResponse[any]						"Injection not found"
//	@Failure		500		{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/injections/{id}/clone [post]
//	@x-api-type		{}
func (h *Handler) CloneInjection(c *gin.Context) {
	id, ok := parsePositiveID(c, consts.URLPathID, "injection ID")
	if !ok {
		return
	}
	var req CloneInjectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	resp, err := h.service.Clone(c.Request.Context(), id, &req)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.JSONResponse(c, http.StatusCreated, "Injection cloned successfully", resp)
}

// GetInjectionLogs handles getting injection execution logs
//
//	@Summary		Get injection logs
//	@Description	Get execution logs for a specific injection
//	@Tags			Injections
//	@ID				get_injection_logs
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int											true	"Injection ID"
//	@Success		200	{object}	dto.GenericResponse[InjectionLogsResp]	"Logs retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]					"Invalid injection ID"
//	@Failure		401	{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		404	{object}	dto.GenericResponse[any]					"Injection not found"
//	@Failure		500	{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/injections/{id}/logs [get]
//	@x-api-type		{}
func (h *Handler) GetInjectionLogs(c *gin.Context) {
	id, ok := parsePositiveID(c, consts.URLPathID, "injection ID")
	if !ok {
		return
	}
	resp, err := h.service.GetLogs(c.Request.Context(), id)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.JSONResponse(c, http.StatusOK, "Logs retrieved successfully", resp)
}

// DownloadDatapack handles datapack file download
//
//	@Summary		Download datapack
//	@Description	Download datapack file by injection ID
//	@Tags			Injections
//	@ID				download_datapack
//	@Produce		application/zip
//	@Security		BearerAuth
//	@Param			id	path		int							true	"Injection ID"
//	@Success		200	{file}		binary						"Datapack zip file"
//	@Failure		400	{object}	dto.GenericResponse[any]	"Invalid injection ID"
//	@Failure		403	{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]	"Injection not found"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/injections/{id}/download [get]
//	@x-api-type		{}
func (h *Handler) DownloadDatapack(c *gin.Context) {
	id, ok := parsePositiveID(c, consts.URLPathID, "injection ID")
	if !ok {
		return
	}
	filename, err := h.service.GetDatapackFilename(c.Request.Context(), id)
	if httpx.HandleServiceError(c, err) {
		return
	}
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", filename))
	zipWriter := zip.NewWriter(c.Writer)
	defer func() { _ = zipWriter.Close() }()
	if err := h.service.DownloadDatapack(c.Request.Context(), zipWriter, []utils.ExculdeRule{}, id); err != nil {
		delete(c.Writer.Header(), "Content-Disposition")
		c.Header("Content-Type", "application/json; charset=utf-8")
		httpx.HandleServiceError(c, err)
	}
}

// ListDatapackFiles handles getting the file structure of an injection datapack
//
//	@Summary		List datapack files
//	@Description	Get the file structure of an injection datapack
//	@Tags			Injections
//	@ID				list_datapack_files
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int											true	"Injection ID"
//	@Success		200	{object}	dto.GenericResponse[DatapackFilesResp]	"Files retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]					"Invalid injection ID"
//	@Failure		401	{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		404	{object}	dto.GenericResponse[any]					"Datapack not found or not ready"
//	@Failure		500	{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/injections/{id}/files [get]
//	@x-api-type		{}
func (h *Handler) ListDatapackFiles(c *gin.Context) {
	id, ok := parsePositiveID(c, consts.URLPathID, "datapack ID")
	if !ok {
		return
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	resp, err := h.service.GetDatapackFiles(c.Request.Context(), id, baseURL)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// DownloadDatapackFile handles downloading a specific file from a datapack.
// Supports HTTP Range requests for resumable downloads.
//
//	@Summary		Download datapack file
//	@Description	Download a specific file from a datapack. Supports Range requests for resumable download.
//	@Tags			Injections
//	@ID				download_datapack_file
//	@Produce		application/octet-stream
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Injection ID"
//	@Param			path	query		string						true	"Relative path to the file"
//	@Success		200		{file}		binary						"Complete file content"
//	@Success		206		{file}		binary						"Partial file content (Range request)"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid injection ID or file path"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]	"Datapack or file not found"
//	@Failure		416		{object}	dto.GenericResponse[any]	"Range not satisfiable"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/injections/{id}/files/download [get]
//	@x-api-type		{}
func (h *Handler) DownloadDatapackFile(c *gin.Context) {
	id, ok := parsePositiveID(c, consts.URLPathID, "datapack ID")
	if !ok {
		return
	}
	filePath := c.Query("path")
	if filePath == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "file path is required")
		return
	}
	fileName, contentType, fileSize, fileReader, err := h.service.DownloadDatapackFile(c.Request.Context(), id, filePath)
	if httpx.HandleServiceError(c, err) {
		return
	}
	defer func() { _ = fileReader.Close() }()
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Accept-Ranges", "bytes")
	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		serveRangeRequest(c, fileReader, fileSize, rangeHeader)
		return
	}
	c.Header("Content-Length", strconv.FormatInt(fileSize, 10))
	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, fileReader); err != nil {
		logrus.WithError(err).Error("failed to stream file content")
	}
}

// QueryDatapackFile handles querying the content of a specific file in the datapack.
// Returns the complete file with Content-Length for download progress tracking.
//
// NOTE: Arrow IPC is a structured stream that must be read sequentially from the
// beginning — Range requests are intentionally NOT supported here. Use
// DownloadDatapackFile for resumable downloads of raw files.
//
//	@Summary		Query datapack file content
//	@Description	Query the content of a parquet file in the datapack, returned as a complete stream. Content-Length header is provided for progress tracking.
//	@Tags			Injections
//	@ID				query_datapack_file
//	@Produce		application/vnd.apache.arrow.stream
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Injection ID"
//	@Param			path	query		string						true	"Relative path to the file"
//	@Success		200		{file}		binary						"Complete Arrow IPC stream"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid injection ID or file path"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]	"Datapack or file not found"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/injections/{id}/files/query [get]
//	@x-api-type		{}
func (h *Handler) QueryDatapackFile(c *gin.Context) {
	id, ok := parsePositiveID(c, consts.URLPathID, "datapack ID")
	if !ok {
		return
	}
	filePath := c.Query("path")
	if filePath == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "file path is required")
		return
	}
	fileName, totalRows, reader, err := h.service.QueryDatapackFile(c.Request.Context(), id, filePath)
	if err != nil && httpx.HandleServiceError(c, err) {
		return
	}
	defer func() { _ = reader.Close() }()
	c.Header("Content-Type", "application/vnd.apache.arrow.stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.arrow", fileName))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("X-Total-Rows", strconv.FormatInt(totalRows, 10))
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, reader); err != nil {
		logrus.Errorf("failed to stream file content: %v", err)
	}
}

// UpdateGroundtruth handles updating ground truth for a datapack
//
//	@Summary		Update datapack ground truth
//	@Description	Update or set ground truth labels for a datapack (fault injection)
//	@Tags			Injections
//	@ID				update_groundtruth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int								true	"Injection ID"
//	@Param			request	body		UpdateGroundtruthReq		true	"Ground truth data"
//	@Success		200		{object}	dto.GenericResponse[any]		"Ground truth updated"
//	@Failure		400		{object}	dto.GenericResponse[any]		"Invalid request"
//	@Failure		404		{object}	dto.GenericResponse[any]		"Injection not found"
//	@Router			/api/v2/injections/{id}/groundtruth [put]
//	@x-api-type		{}
func (h *Handler) UpdateGroundtruth(c *gin.Context) {
	id, ok := parsePositiveID(c, consts.URLPathID, "injection ID")
	if !ok {
		return
	}
	var req UpdateGroundtruthReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	if httpx.HandleServiceError(c, h.service.UpdateGroundtruth(c.Request.Context(), id, &req)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusOK, "Groundtruth updated successfully", nil)
}

// UploadDatapack handles manual datapack upload
//
//	@Summary		Upload a manual datapack
//	@Description	Upload a zip archive as a manual datapack data source
//	@Tags			Injections
//	@ID				upload_datapack
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			name			formData	string						true	"Datapack name"
//	@Param			description		formData	string						false	"Description"
//	@Param			category		formData	string						false	"Category"
//	@Param			labels			formData	string						false	"JSON-encoded labels"
//	@Param			ground_truths	formData	string						false	"JSON-encoded ground truths"
//	@Param			file			formData	file						true	"Zip archive file"
//	@Success		201				{object}	dto.GenericResponse[UploadDatapackResp]	"Datapack uploaded successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]	"Invalid request"
//	@Failure		401				{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		500				{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/injections/upload [post]
//	@x-api-type		{}
func (h *Handler) UploadDatapack(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "file is required: "+err.Error())
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "failed to open uploaded file: "+err.Error())
		return
	}
	defer func() { _ = file.Close() }()

	req := &UploadDatapackReq{
		Name:         c.PostForm("name"),
		Description:  c.PostForm("description"),
		Category:     c.PostForm("category"),
		Labels:       c.PostForm("labels"),
		Groundtruths: c.PostForm("groundtruths"),
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.UploadDatapack(c.Request.Context(), req, file, fileHeader.Size)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.JSONResponse(c, http.StatusCreated, "Datapack uploaded successfully", resp)
}

func (h *Handler) searchInjections(c *gin.Context, projectID *int) {
	var req SearchInjectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.service.Search(c.Request.Context(), &req, projectID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

func (h *Handler) listFaultInjectionNoIssues(c *gin.Context, projectID *int) {
	var req ListInjectionNoIssuesReq
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

	items, err := h.service.ListNoIssues(c.Request.Context(), &req, projectID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, items)
}

func (h *Handler) listFaultInjectionWithIssues(c *gin.Context, projectID *int) {
	var req ListInjectionWithIssuesReq
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

	items, err := h.service.ListWithIssues(c.Request.Context(), &req, projectID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, items)
}

func (h *Handler) submitFaultInjection(c *gin.Context, projectID *int) {
	groupID := c.GetString("groupID")
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	spanCtx, span, ok := spanFromGin(c, "SubmitFaultInjection")
	if !ok {
		return
	}

	var req SubmitInjectionReq
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

	if req.ProjectName == "" && projectID == nil {
		span.SetStatus(codes.Error, "validation error in SubmitFaultInjection: project name is required")
		dto.ErrorResponse(c, http.StatusBadRequest, "Project name or ID is required")
		return
	}

	resp, err := h.service.SubmitFaultInjection(spanCtx, &req, groupID, userID, projectID)
	if err != nil {
		span.SetStatus(codes.Error, "service error in SubmitFaultInjection: "+err.Error())
		logrus.Errorf("Failed to submit fault injection: %v", err)
		httpx.HandleServiceError(c, err)
		return
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("Successfully submitted %d fault injections with groupID: %s", len(resp.Items), groupID))
	dto.SuccessResponse(c, resp)
}

func (h *Handler) submitDatapackBuilding(c *gin.Context, projectID *int) {
	groupID := c.GetString("groupID")
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	spanCtx, span, ok := spanFromGin(c, "SubmitDatapackBuilding")
	if !ok {
		return
	}

	var req SubmitDatapackBuildingReq
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

	if req.ProjectName == "" && projectID == nil {
		span.SetStatus(codes.Error, "validation error in SubmitFaultInjection: project name is required")
		dto.ErrorResponse(c, http.StatusBadRequest, "Project name or ID is required")
		return
	}

	resp, err := h.service.SubmitDatapackBuilding(spanCtx, &req, groupID, userID, projectID)
	if err != nil {
		span.SetStatus(codes.Error, "service error in SubmitDatapackBuilding: "+err.Error())
		logrus.Errorf("Failed to submit datapack building: %v", err)
		httpx.HandleServiceError(c, err)
		return
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("Successfully submitted %d datapack buildings with groupID: %s", len(resp.Items), groupID))
	dto.SuccessResponse(c, resp)
}

func spanFromGin(c *gin.Context, operation string) (context.Context, trace.Span, bool) {
	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		logrus.Errorf("Failed to get span context from gin.Context in %s", operation)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get span context")
		return nil, nil, false
	}

	spanCtx := ctx.(context.Context)
	return spanCtx, trace.SpanFromContext(spanCtx), true
}

func parsePositiveID(c *gin.Context, key, label string) (int, bool) {
	id, ok := httpx.ParsePositiveID(c, c.Param(key), label)
	return id, ok
}

func serveRangeRequest(c *gin.Context, reader io.ReadSeeker, fileSize int64, rangeHeader string) {
	const prefix = "bytes="
	if !strings.HasPrefix(rangeHeader, prefix) {
		dto.ErrorResponse(c, http.StatusRequestedRangeNotSatisfiable, "invalid range format")
		return
	}
	rangeSpec := strings.TrimPrefix(rangeHeader, prefix)
	if strings.Contains(rangeSpec, ",") {
		dto.ErrorResponse(c, http.StatusRequestedRangeNotSatisfiable, "multi-range not supported")
		return
	}
	parts := strings.SplitN(rangeSpec, "-", 2)
	if len(parts) != 2 {
		dto.ErrorResponse(c, http.StatusRequestedRangeNotSatisfiable, "invalid range format")
		return
	}
	var start, end int64
	var err error
	if parts[0] == "" {
		suffix, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || suffix <= 0 || suffix > fileSize {
			c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
			dto.ErrorResponse(c, http.StatusRequestedRangeNotSatisfiable, "invalid range")
			return
		}
		start = fileSize - suffix
		end = fileSize - 1
	} else {
		start, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil || start < 0 || start >= fileSize {
			c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
			dto.ErrorResponse(c, http.StatusRequestedRangeNotSatisfiable, "invalid range start")
			return
		}
		if parts[1] == "" {
			end = fileSize - 1
		} else {
			end, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil || end < start || end >= fileSize {
				c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
				dto.ErrorResponse(c, http.StatusRequestedRangeNotSatisfiable, "invalid range end")
				return
			}
		}
	}
	contentLength := end - start + 1
	if _, err := reader.Seek(start, io.SeekStart); err != nil {
		logrus.Errorf("failed to seek to range start: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to seek to range start")
		return
	}
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
	c.Status(http.StatusPartialContent)
	if _, err := io.CopyN(c.Writer, reader, contentLength); err != nil {
		logrus.Errorf("failed to stream partial content: %v", err)
	}
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
