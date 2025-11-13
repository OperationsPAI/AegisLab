package system

import (
	"net/http"
	"strconv"

	"aegis/dto"
	"aegis/handlers"
	producer "aegis/service/prodcuer"

	"github.com/gin-gonic/gin"
)

// GetAuditLog handles single audit log retrieval
//
//	@Summary		Get audit log by ID
//	@Description	Get a specific audit log entry by ID
//	@Tags			System
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int											true	"Audit log ID"
//	@Success		200	{object}	dto.GenericResponse[dto.AuditLogDetailResp]	"Audit log retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]					"Invalid ID"
//	@Failure		401	{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]					"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]					"Audit log not found"
//	@Failure		500	{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/system/audit/{id} [get]
func GetAuditLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid audit log ID")
		return
	}

	resp, err := producer.GetAuditLogDetail(id)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			page		query		int													false	"Page number"	default(1)
//	@Param			size		query		int													false	"Page size"		default(20)
//	@Param			action		query		string												false	"Filter by action"
//	@Param			user_id		query		int													false	"Filter by user ID"
//	@Param			resource_id	query		int													false	"Filter by resource ID"
//	@Param			state		query		int													false	"Filter by state"
//	@Param			status		query		int													false	"Filter by status"
//	@Param			start_date	query		string												false	"Filter from date (YYYY-MM-DD)"
//	@Param			end_date	query		string												false	"Filter to date (YYYY-MM-DD)"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[dto.AuditLogResp]]	"Audit logs retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]							"Invalid request format/parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/system/audit [get]
func ListAuditLogs(c *gin.Context) {
	var req dto.ListAuditLogReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	resp, err := producer.ListAuditLogs(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Audit logs retrieved successfully", resp)
}
