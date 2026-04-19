package system

import (
	"aegis/httpx"
	"net/http"
	"strconv"

	"aegis/consts"
	"aegis/dto"
	"aegis/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service HandlerService
}

func NewHandler(service HandlerService) *Handler {
	return &Handler{service: service}
}

// GetHealth handles system health check
//
//	@Summary		System health check
//	@Description	Get system health status and service information
//	@Tags			System
//	@ID				get_system_health
//	@Produce		json
//	@Success		200	{object}	dto.GenericResponse[HealthCheckResp]	"Health check successful"
//	@Failure		500	{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/system/health [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) GetHealth(c *gin.Context) {
	resp, err := h.service.GetHealth(c.Request.Context())
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get health status: "+err.Error())
		return
	}
	dto.SuccessResponse(c, resp)
}

// GetMetrics handles monitoring metrics query
//
//	@Summary		Get monitoring metrics
//	@Description	Deprecated: This endpoint returns hardcoded/fabricated data. Use the v2 equivalent GET /api/v2/system/metrics which provides real system metrics via gopsutil.
//	@Deprecated
//	@Tags		System
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		request	body		MonitoringQueryReq							true	"Metrics query request"
//	@Success	200		{object}	dto.GenericResponse[MonitoringMetricsResp]	"Metrics retrieved successfully"
//	@Success	400		{object}	dto.GenericResponse[any]					"Invalid request format"
//	@Failure	401		{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure	403		{object}	dto.GenericResponse[any]					"Permission denied"
//	@Failure	500		{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router		/system/monitor/metrics [post]
//	@x-api-type	{"admin":"true"}
func (h *Handler) GetMetrics(c *gin.Context) {
	var req MonitoringQueryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	c.Header("Deprecation", "true")
	c.Header("Link", `</api/v2/system/metrics>; rel="successor-version"`)
	resp, err := h.service.GetMetrics(c.Request.Context())
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// GetSystemInfo handles basic system information
//
//	@Summary		Get system information
//	@Description	Deprecated: This endpoint returns partially hardcoded data. Use the v2 equivalent GET /api/v2/system/metrics which provides real system metrics via gopsutil.
//	@Deprecated
//	@Tags		System
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	dto.GenericResponse[SystemInfo]	"System info retrieved successfully"
//	@Failure	401	{object}	dto.GenericResponse[any]		"Authentication required"
//	@Failure	403	{object}	dto.GenericResponse[any]		"Permission denied"
//	@Failure	500	{object}	dto.GenericResponse[any]		"Internal server error"
//	@Router		/system/monitor/info [get]
//	@x-api-type	{"admin":"true"}
func (h *Handler) GetSystemInfo(c *gin.Context) {
	c.Header("Deprecation", "true")
	c.Header("Link", `</api/v2/system/metrics>; rel="successor-version"`)
	resp, err := h.service.GetSystemInfo(c.Request.Context())
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// ListNamespaceLocks handles listing of namespace locks
//
//	@Summary		List namespace locks
//	@Description	Retrieve the list of currently locked namespaces
//	@Tags			System
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GenericResponse[ListNamespaceLockResp]	"Successfully retrieved the list of locks"
//	@Failure		401	{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]					"Permission denied"
//	@Failure		500	{object}	dto.GenericResponse[any]					"Internal Server Error"
//	@Router			/system/monitor/namespaces/locks [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) ListNamespaceLocks(c *gin.Context) {
	resp, err := h.service.ListNamespaceLocks(c.Request.Context())
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.JSONResponse(c, http.StatusOK, "Successfully retrieved the list of locks", resp)
}

// ListQueuedTasks handles listing of queued tasks
//
//	@Summary		List queued tasks
//	@Description	List tasks in queue (ready and delayed)
//	@Tags			System
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GenericResponse[QueuedTasksResp]	"Queued tasks retrieved successfully"
//	@Failure		401	{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]				"No queued tasks found"
//	@Failure		500	{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/system/monitor/tasks/queue [post]
//	@x-api-type		{"admin":"true"}
func (h *Handler) ListQueuedTasks(c *gin.Context) {
	resp, err := h.service.ListQueuedTasks(c.Request.Context())
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.JSONResponse(c, http.StatusOK, "Queued tasks retrieved successfully", resp)
}

// GetAuditLog handles single audit log retrieval
//
//	@Summary		Get audit log by ID
//	@Description	Get a specific audit log entry by ID
//	@Tags			System
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int										true	"Audit log ID"
//	@Success		200	{object}	dto.GenericResponse[AuditLogDetailResp]	"Audit log retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]				"Invalid ID"
//	@Failure		401	{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]				"Audit log not found"
//	@Failure		500	{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/system/audit/{id} [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) GetAuditLog(c *gin.Context) {
	id, ok := parseID(c, "id", "Invalid audit log ID")
	if !ok {
		return
	}

	resp, err := h.service.GetAuditLog(c.Request.Context(), id)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// ListAuditLogs handles audit log listing
//
//	@Summary		List audit logs
//	@Description	Get paginated list of audit logs with optional filtering
//	@Tags			System
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int												false	"Page number"	default(1)
//	@Param			size		query		int												false	"Page size"		default(20)
//	@Param			action		query		string											false	"Filter by action"
//	@Param			user_id		query		int												false	"Filter by user ID"
//	@Param			resource_id	query		int												false	"Filter by resource ID"
//	@Param			state		query		int												false	"Filter by state"
//	@Param			status		query		int												false	"Filter by status"
//	@Param			start_date	query		string											false	"Filter from date (YYYY-MM-DD)"
//	@Param			end_date	query		string											false	"Filter to date (YYYY-MM-DD)"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[AuditLogResp]]	"Audit logs retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]						"Invalid request format/parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/system/audit [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) ListAuditLogs(c *gin.Context) {
	var req ListAuditLogReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	resp, err := h.service.ListAuditLogs(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.JSONResponse(c, http.StatusOK, "Audit logs retrieved successfully", resp)
}

// GetConfig retrieves a configuration by ID
//
//	@Summary		Get configuration
//	@Description	Get detailed information about a specific configuration
//	@Tags			Configurations
//	@ID				get_config_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			config_id	path		int								true	"Configuration ID"
//	@Success		200			{object}	dto.GenericResponse[ConfigResp]	"Configuration retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]		"Invalid config ID"
//	@Failure		401			{object}	dto.GenericResponse[any]		"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]		"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]		"Config not found"
//	@Failure		500			{object}	dto.GenericResponse[any]		"Internal server error"
//	@Router			/system/configs/{config_id} [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) GetConfig(c *gin.Context) {
	configID, ok := parseID(c, consts.URLPathConfigID, "Invalid config ID")
	if !ok {
		return
	}

	resp, err := h.service.GetConfig(c.Request.Context(), configID)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// ListConfigs lists configurations with pagination and filtering
//
//	@Summary		List configurations
//	@Description	List configurations with pagination and optional filters
//	@Tags			Configurations
//	@ID				list_configs
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int												false	"Page number"	default(1)
//	@Param			page_size	query		int												false	"Page size"		default(20)
//	@Param			category	query		string											false	"Filter by configuration category"
//	@Param			value_type	query		consts.ConfigValueType							false	"Filter by configuration value type"
//	@Param			is_secret	query		bool											false	"Filter by secret status"
//	@Param			updated_by	query		int												false	"Filter by ID of the user who last updated the config"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[ConfigResp]]	"Configurations retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]						"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/system/configs [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) ListConfigs(c *gin.Context) {
	var req ListConfigReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.ListConfigs(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// RollbackConfigValue rolls back a configuration value to previous value from history
//
//	@Summary		Rollback configuration value
//	@Description	Rollback a configuration value to a previous value from history
//	@Tags			Configurations
//	@ID				rollback_config_value
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			config_id	path		int							true	"Configuration ID"
//	@Param			rollback	body		RollbackConfigReq			true	"Rollback request with history_id and reason"
//	@Success		202			{object}	dto.GenericResponse[any]	"Configuration value rolled back successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid config ID/request format/history is not a value change"
//	@Failure		401			{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		404			{object}	dto.GenericResponse[any]	"Configuration or history not found"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/system/configs/{config_id}/value/rollback [post]
//	@x-api-type		{"admin":"true"}
func (h *Handler) RollbackConfigValue(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	configID, ok := parseID(c, consts.URLPathConfigID, "Invalid config ID")
	if !ok {
		return
	}

	var req RollbackConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	err := h.service.RollbackConfigValue(c.Request.Context(), &req, configID, userID, c.ClientIP(), c.Request.UserAgent())
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.JSONResponse[any](c, http.StatusAccepted, "Configuration value rolled back successfully", nil)
}

// RollbackConfigMetadata rolls back a configuration metadata field to previous value from history
//
//	@Summary		Rollback configuration metadata
//	@Description	Rollback a configuration metadata field (e.g., min_value, max_value, pattern) to a previous value from history
//	@Tags			Configurations
//	@ID				rollback_config_metadata
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			config_id	path		int								true	"Configuration ID"
//	@Param			rollback	body		RollbackConfigReq				true	"Rollback request with history_id and reason"
//	@Success		200			{object}	dto.GenericResponse[ConfigResp]	"Configuration metadata rolled back successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]		"Invalid config ID/request format/history is a value change"
//	@Failure		401			{object}	dto.GenericResponse[any]		"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]		"Permission denied - admin only"
//	@Failure		404			{object}	dto.GenericResponse[any]		"Configuration or history not found"
//	@Failure		500			{object}	dto.GenericResponse[any]		"Internal server error"
//	@Router			/system/configs/{config_id}/metadata/rollback [post]
//	@x-api-type		{"admin":"true"}
func (h *Handler) RollbackConfigMetadata(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	configID, ok := parseID(c, consts.URLPathConfigID, "Invalid config ID")
	if !ok {
		return
	}

	var req RollbackConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	resp, err := h.service.RollbackConfigMetadata(c.Request.Context(), &req, configID, userID, c.ClientIP(), c.Request.UserAgent())
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.JSONResponse(c, http.StatusOK, "Configuration metadata rolled back successfully", resp)
}

// UpdateConfigValue updates a configuration value (runtime operational change)
//
//	@Summary		Update configuration value
//	@Description	Update a configuration value with validation and history tracking. This is for frequent operational adjustments.
//	@Tags			Configurations
//	@ID				update_config_value
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			config_id	path		int							true	"Configuration ID"
//	@Param			request		body		UpdateConfigValueReq		true	"Configuration value update request"
//	@Success		202			{object}	dto.GenericResponse[any]	"Configuration value updated successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid config ID/request"
//	@Failure		401			{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		404			{object}	dto.GenericResponse[any]	"Configuration not found"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/system/configs/{config_id} [patch]
//	@x-api-type		{"admin":"true"}
func (h *Handler) UpdateConfigValue(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	configID, ok := parseID(c, consts.URLPathConfigID, "Invalid config ID")
	if !ok {
		return
	}

	var req UpdateConfigValueReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	err := h.service.UpdateConfigValue(c.Request.Context(), &req, configID, userID, c.ClientIP(), c.Request.UserAgent())
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.JSONResponse[any](c, http.StatusAccepted, "Configuration value updated successfully", nil)
}

// UpdateConfigMetadata updates configuration metadata (rare admin operation)
//
//	@Summary		Update configuration metadata
//	@Description	Update configuration metadata such as min/max values, validation rules, etc. This is a high-privilege operation.
//	@Tags			Configurations
//	@ID				update_config_metadata
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			config_id	path		int								true	"Configuration ID"
//	@Param			request		body		UpdateConfigMetadataReq			true	"Configuration metadata update request"
//	@Success		200			{object}	dto.GenericResponse[ConfigResp]	"Configuration metadata updated successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]		"Invalid config ID/request"
//	@Failure		401			{object}	dto.GenericResponse[any]		"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]		"Permission denied - admin only"
//	@Failure		404			{object}	dto.GenericResponse[any]		"Configuration not found"
//	@Failure		500			{object}	dto.GenericResponse[any]		"Internal server error"
//	@Router			/system/configs/{config_id}/metadata [put]
//	@x-api-type		{"admin":"true"}
func (h *Handler) UpdateConfigMetadata(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	configID, ok := parseID(c, consts.URLPathConfigID, "Invalid config ID")
	if !ok {
		return
	}

	var req UpdateConfigMetadataReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.UpdateConfigMetadata(c.Request.Context(), &req, configID, userID, c.ClientIP(), c.Request.UserAgent())
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.JSONResponse(c, http.StatusOK, "Configuration metadata updated successfully", resp)
}

// ListConfigHistories handles listing config histories with pagination and filtering
//
//	@Summary		List configuration histories
//	@Description	Get paginated list of config histories for a specific config
//	@Tags			Configurations
//	@ID				list_config_histories
//	@Produce		json
//	@Security		BearerAuth
//	@Param			config_id	path		int														true	"Configuration ID"
//	@Param			page		query		int														false	"Page number"	default(1)
//	@Param			size		query		int														false	"Page size"		default(20)
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[ConfigHistoryResp]]	"Config histories retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]								"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/system/configs/{config_id}/histories [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) ListConfigHistories(c *gin.Context) {
	configID, ok := parseID(c, consts.URLPathConfigID, "Invalid config ID")
	if !ok {
		return
	}

	var req ListConfigHistoryReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.ListConfigHistories(c.Request.Context(), &req, configID)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.JSONResponse(c, http.StatusOK, "Config historys retrieved successfully", resp)
}

func parseID(c *gin.Context, param, message string) (int, bool) {
	value := c.Param(param)
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, message)
		return 0, false
	}
	return id, true
}
