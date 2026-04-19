package project

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

// CreateProject handles project creation
//
//	@Summary		Create a new project
//	@Description	Create a new project with specified details
//	@Tags			Projects
//	@ID				create_project
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateProjectReq					true	"Project creation request"
//	@Success		201		{object}	dto.GenericResponse[ProjectResp]	"Project created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		409		{object}	dto.GenericResponse[any]			"Project already exists"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/projects [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) CreateProject(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req CreateProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.service.CreateProject(c.Request.Context(), &req, userID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusCreated, "Project created successfully", resp)
}

// DeleteProject handles project deletion
//
//	@Summary		Delete project
//	@Description	Delete a project
//	@Tags			Projects
//	@ID				delete_project
//	@Produce		json
//	@Security		BearerAuth
//	@Param			project_id	path		int							true	"Project ID"
//	@Success		204			{object}	dto.GenericResponse[any]	"Project deleted successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid project ID"
//	@Failure		401			{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]	"Project not found"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/projects/{project_id} [delete]
//	@x-api-type		{"portal":"true"}
func (h *Handler) DeleteProject(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	err := h.service.DeleteProject(c.Request.Context(), projectID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Project deleted successfully", nil)
}

// GetProjectDetail handles getting a single project by ID
//
//	@Summary		Get project by ID
//	@Description	Get detailed information about a specific project
//	@Tags			Projects
//	@ID				get_project_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			project_id	path		int										true	"Project ID"
//	@Success		200			{object}	dto.GenericResponse[ProjectDetailResp]	"Project retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]				"Invalid project ID"
//	@Failure		401			{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]				"Project not found"
//	@Failure		500			{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/projects/{project_id} [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) GetProjectDetail(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	resp, err := h.service.GetProjectDetail(c.Request.Context(), projectID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// ListProjects handles listing projects with pagination and filtering
//
//	@Summary		List projects
//	@Description	Get paginated list of projects with filtering
//	@Tags			Projects
//	@ID				list_projects
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int												false	"Page number"	default(1)
//	@Param			size		query		int												false	"Page size"		default(20)
//	@Param			is_public	query		bool											false	"Filter by public status"
//	@Param			status		query		consts.StatusType								false	"Filter by status"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[ProjectResp]]	"Projects retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]						"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/projects [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) ListProjects(c *gin.Context) {
	var req ListProjectReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.service.ListProjects(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// UpdateProject handles project updates
//
//	@Summary		Update project
//	@Description	Update an existing project's information
//	@Tags			Projects
//	@ID				update_project
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			project_id	path		int									true	"Project ID"
//	@Param			request		body		UpdateProjectReq					true	"Project update request"
//	@Success		202			{object}	dto.GenericResponse[ProjectResp]	"Project updated successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]			"Invalid project ID or invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]			"Project not found"
//	@Failure		500			{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/projects/{project_id} [patch]
//	@x-api-type		{"portal":"true"}
func (h *Handler) UpdateProject(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req UpdateProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.service.UpdateProject(c.Request.Context(), &req, projectID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusAccepted, "Project updated successfully", resp)
}

// ManageProjectCustomLabels manages project custom labels (key-value pairs)
//
//	@Summary		Manage project custom labels
//	@Description	Add or remove custom labels (key-value pairs) for a project
//	@Tags			Projects
//	@ID				update_project_labels
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			project_id	path		int									true	"Project ID"
//	@Param			manage		body		ManageProjectLabelReq				true	"Label management request"
//	@Success		200			{object}	dto.GenericResponse[ProjectResp]	"Labels managed successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]			"Invalid project ID or invalid request format/parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]			"Project not found"
//	@Failure		500			{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/projects/{project_id}/labels [patch]
//	@x-api-type		{"portal":"true"}
func (h *Handler) ManageProjectCustomLabels(c *gin.Context) {
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req ManageProjectLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.service.ManageProjectLabels(c.Request.Context(), &req, projectID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
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
