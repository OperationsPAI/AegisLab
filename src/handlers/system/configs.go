package system

import (
	"net/http"
	"strconv"

	"aegis/consts"
	"aegis/dto"
	"aegis/handlers"
	"aegis/middleware"
	producer "aegis/service/producer"

	"github.com/gin-gonic/gin"
)

// GetConfig retrieves a configuration by ID
//
//	@Summary		Get configuration
//	@Description	Get detailed information about a specific configuration
//	@Tags			Configurations
//	@ID				get_config_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			config_id	path		int									true	"Configuration ID"
//	@Success		200			{object}	dto.GenericResponse[dto.ConfigResp]	"Configuration retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]			"Invalid config ID"
//	@Failure		401			{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]			"Config not found"
//	@Failure		500			{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/system/configs/{config_id} [get]
func GetConfig(c *gin.Context) {
	configIDStr := c.Param(consts.URLPathConfigID)
	configID, err := strconv.Atoi(configIDStr)
	if err != nil || configID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid config ID")
		return
	}

	resp, err := producer.GetConfigDetail(configID)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			page		query		int													false	"Page number"	default(1)
//	@Param			page_size	query		int													false	"Page size"		default(20)
//	@Param			category	query		string												false	"Filter by configuration category"
//	@Param			value_type	query		consts.ConfigValueType								false	"Filter by configuration value type"
//	@Param			is_secret	query		bool												false	"Filter by secret status"
//	@Param			updated_by	query		int													false	"Filter by ID of the user who last updated the config"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[dto.ConfigResp]]	"Configurations retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]							"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/system/configs [get]
func ListConfigs(c *gin.Context) {
	var req dto.ListConfigReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListConfigs(&req)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			rollback	body		dto.RollbackConfigReq		true	"Rollback request with history_id and reason"
//	@Success		202			{object}	dto.GenericResponse[any]	"Configuration value rolled back successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid config ID/request format/history is not a value change"
//	@Failure		401			{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		404			{object}	dto.GenericResponse[any]	"Configuration or history not found"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/system/configs/{config_id}/value/rollback [post]
func RollbackConfigValue(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	configIDStr := c.Param(consts.URLPathConfigID)
	configID, err := strconv.Atoi(configIDStr)
	if err != nil || configID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid config ID")
		return
	}

	ctx := c.Request.Context()

	var req dto.RollbackConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	err = producer.RollbackConfigValue(ctx, &req, configID, userID, ipAddress, userAgent)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			config_id	path		int									true	"Configuration ID"
//	@Param			rollback	body		dto.RollbackConfigReq				true	"Rollback request with history_id and reason"
//	@Success		200			{object}	dto.GenericResponse[dto.ConfigResp]	"Configuration metadata rolled back successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]			"Invalid config ID/request format/history is a value change"
//	@Failure		401			{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]			"Permission denied - admin only"
//	@Failure		404			{object}	dto.GenericResponse[any]			"Configuration or history not found"
//	@Failure		500			{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/system/configs/{config_id}/metadata/rollback [post]
func RollbackConfigMetadata(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	configIDStr := c.Param(consts.URLPathConfigID)
	configID, err := strconv.Atoi(configIDStr)
	if err != nil || configID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid config ID")
		return
	}

	var req dto.RollbackConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	resp, err := producer.RollbackConfigMetadata(&req, configID, userID, ipAddress, userAgent)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			request		body		dto.UpdateConfigValueReq	true	"Configuration value update request"
//	@Success		202			{object}	dto.GenericResponse[any]	"Configuration value updated successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid config ID/request"
//	@Failure		401			{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		404			{object}	dto.GenericResponse[any]	"Configuration not found"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/system/configs/{config_id} [patch]
func UpdateConfigValue(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	configIDStr := c.Param(consts.URLPathConfigID)
	configID, err := strconv.Atoi(configIDStr)
	if err != nil || configID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid config ID")
		return
	}

	ctx := c.Request.Context()

	var req dto.UpdateConfigValueReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	err = producer.UpdateConfigValue(ctx, &req, configID, userID, ipAddress, userAgent)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			config_id	path		int									true	"Configuration ID"
//	@Param			request		body		dto.UpdateConfigMetadataReq			true	"Configuration metadata update request"
//	@Success		200			{object}	dto.GenericResponse[dto.ConfigResp]	"Configuration metadata updated successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]			"Invalid config ID/request"
//	@Failure		401			{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]			"Permission denied - admin only"
//	@Failure		404			{object}	dto.GenericResponse[any]			"Configuration not found"
//	@Failure		500			{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/system/configs/{config_id}/metadata [put]
func UpdateConfigMetadata(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	configIDStr := c.Param(consts.URLPathConfigID)
	configID, err := strconv.Atoi(configIDStr)
	if err != nil || configID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid config ID")
		return
	}

	var req dto.UpdateConfigMetadataReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	resp, err := producer.UpdateConfigMetadata(&req, configID, userID, ipAddress, userAgent)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Configuration metadata updated successfully", resp)
}

// ===================== Config History =====================

// ListConfigHistories handles listing config histories with pagination and filtering
//
//	@Summary		List configuration histories
//	@Description	Get paginated list of config histories for a specific config
//	@Tags			Configurations
//	@ID				list_config_histories
//	@Produce		json
//	@Security		BearerAuth
//	@Param			config_id	path		int															true	"Configuration ID"
//	@Param			page		query		int															false	"Page number"	default(1)
//	@Param			size		query		int															false	"Page size"		default(20)
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[dto.ConfigHistoryResp]]	"Config histories retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]									"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]									"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]									"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]									"Internal server error"
//	@Router			/system/configs/{config_id}/histories [get]
func ListConfigHistories(c *gin.Context) {
	configIDStr := c.Param(consts.URLPathConfigID)
	configID, err := strconv.Atoi(configIDStr)
	if err != nil || configID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid config ID")
		return
	}

	var req dto.ListConfigHistoryReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListConfigHistories(&req, configID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Config historys retrieved successfully", resp)
}
