package chaossystemmodule

import (
	"aegis/httpx"
	"net/http"

	"aegis/consts"
	"aegis/dto"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service HandlerService
}

func NewHandler(service HandlerService) *Handler {
	return &Handler{service: service}
}

// ListChaosSystemsHandler handles listing chaos systems with pagination
//
//	@Summary		List chaos systems
//	@Description	Get a paginated list of registered chaos systems
//	@Tags			Systems
//	@ID				list_chaos_systems
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int													false	"Page number"	default(1)
//	@Param			size	query		int													false	"Page size"		default(20)
//	@Success		200		{object}	dto.GenericResponse[dto.ListResp[ChaosSystemResp]]	"Systems retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]								"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		500		{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/systems [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) ListSystems(c *gin.Context) {
	var req ListChaosSystemReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.ListSystems(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// GetChaosSystemHandler handles getting a single chaos system by ID
//
//	@Summary		Get chaos system by ID
//	@Description	Get detailed information about a specific chaos system
//	@Tags			Systems
//	@ID				get_chaos_system
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int										true	"System ID"
//	@Success		200	{object}	dto.GenericResponse[ChaosSystemResp]	"System retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]				"Invalid system ID"
//	@Failure		404	{object}	dto.GenericResponse[any]				"System not found"
//	@Failure		500	{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/systems/{id} [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) GetSystem(c *gin.Context) {
	id, ok := httpx.ParsePositiveID(c, c.Param(consts.URLPathID), "system ID")
	if !ok {
		return
	}
	resp, err := h.service.GetSystem(c.Request.Context(), id)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// CreateChaosSystemHandler handles creating a new chaos system
//
//	@Summary		Create chaos system
//	@Description	Register a new chaos system
//	@Tags			Systems
//	@ID				create_chaos_system
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateChaosSystemReq					true	"System creation request"
//	@Success		201		{object}	dto.GenericResponse[ChaosSystemResp]	"System created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		409		{object}	dto.GenericResponse[any]				"System already exists"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/systems [post]
//	@x-api-type		{"admin":"true"}
func (h *Handler) CreateSystem(c *gin.Context) {
	var req CreateChaosSystemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	resp, err := h.service.CreateSystem(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.JSONResponse(c, http.StatusCreated, "System created successfully", resp)
}

// UpdateChaosSystemHandler handles updating a chaos system
//
//	@Summary		Update chaos system
//	@Description	Update an existing chaos system
//	@Tags			Systems
//	@ID				update_chaos_system
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int										true	"System ID"
//	@Param			request	body		UpdateChaosSystemReq					true	"System update request"
//	@Success		200		{object}	dto.GenericResponse[ChaosSystemResp]	"System updated successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid request format or parameters"
//	@Failure		404		{object}	dto.GenericResponse[any]				"System not found"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/systems/{id} [put]
//	@x-api-type		{"admin":"true"}
func (h *Handler) UpdateSystem(c *gin.Context) {
	id, ok := httpx.ParsePositiveID(c, c.Param(consts.URLPathID), "system ID")
	if !ok {
		return
	}
	var req UpdateChaosSystemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	resp, err := h.service.UpdateSystem(c.Request.Context(), id, &req)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// DeleteChaosSystemHandler handles deleting a chaos system
//
//	@Summary		Delete chaos system
//	@Description	Soft-delete a chaos system (builtin systems cannot be deleted)
//	@Tags			Systems
//	@ID				delete_chaos_system
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int							true	"System ID"
//	@Success		200	{object}	dto.GenericResponse[any]	"System deleted successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]	"Invalid system ID or builtin system"
//	@Failure		404	{object}	dto.GenericResponse[any]	"System not found"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/systems/{id} [delete]
//	@x-api-type		{"admin":"true"}
func (h *Handler) DeleteSystem(c *gin.Context) {
	id, ok := httpx.ParsePositiveID(c, c.Param(consts.URLPathID), "system ID")
	if !ok {
		return
	}
	if httpx.HandleServiceError(c, h.service.DeleteSystem(c.Request.Context(), id)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusOK, "System deleted successfully", nil)
}

// UpsertChaosSystemMetadataHandler handles bulk upserting metadata for a chaos system
//
//	@Summary		Upsert chaos system metadata
//	@Description	Bulk upsert metadata entries for a chaos system
//	@Tags			Systems
//	@ID				upsert_chaos_system_metadata
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int									true	"System ID"
//	@Param			request	body		BulkUpsertSystemMetadataReq		true	"Metadata upsert request"
//	@Success		200		{object}	dto.GenericResponse[any]			"Metadata upserted successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid request format or parameters"
//	@Failure		404		{object}	dto.GenericResponse[any]			"System not found"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/systems/{id}/metadata [post]
//	@x-api-type		{"admin":"true"}
func (h *Handler) UpsertMetadata(c *gin.Context) {
	id, ok := httpx.ParsePositiveID(c, c.Param(consts.URLPathID), "system ID")
	if !ok {
		return
	}
	var req BulkUpsertSystemMetadataReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if httpx.HandleServiceError(c, h.service.UpsertMetadata(c.Request.Context(), id, &req)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusOK, "Metadata upserted successfully", nil)
}

// ListChaosSystemMetadataHandler handles listing metadata for a chaos system
//
//	@Summary		List chaos system metadata
//	@Description	List metadata entries for a chaos system, optionally filtered by type
//	@Tags			Systems
//	@ID				list_chaos_system_metadata
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int												true	"System ID"
//	@Param			type	query		string											false	"Metadata type filter"
//	@Success		200		{object}	dto.GenericResponse[[]SystemMetadataResp]	"Metadata retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]						"Invalid system ID"
//	@Failure		404		{object}	dto.GenericResponse[any]						"System not found"
//	@Failure		500		{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/systems/{id}/metadata [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) ListMetadata(c *gin.Context) {
	id, ok := httpx.ParsePositiveID(c, c.Param(consts.URLPathID), "system ID")
	if !ok {
		return
	}
	resp, err := h.service.ListMetadata(c.Request.Context(), id, c.Query("type"))
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}
