package v2

import (
	"aegis/consts"
	"net/http"
	"strconv"

	"aegis/dto"
	"aegis/handlers"
	"aegis/middleware"
	producer "aegis/service/prodcuer"

	"github.com/gin-gonic/gin"
)

// CreateProject handles project creation
//
//	@Summary		Create a new project
//	@Description	Create a new project with specified details
//	@Tags			Projects
//	@ID				create_project
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateProjectReq					true	"Project creation request"
//	@Success		201		{object}	dto.GenericResponse[dto.ProjectResp]	"Project created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		409		{object}	dto.GenericResponse[any]				"Project already exists"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/projects [post]
func CreateProject(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req dto.CreateProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.CreateProject(&req, userID)
	if handlers.HandleServiceError(c, err) {
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
func DeleteProject(c *gin.Context) {
	projectIdStr := c.Param(consts.URLPathProjectID)
	projectID, err := strconv.Atoi(projectIdStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	err = producer.DeleteProject(projectID)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			project_id	path		int											true	"Project ID"
//	@Success		200			{object}	dto.GenericResponse[dto.ProjectDetailResp]	"Project retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]					"Invalid project ID"
//	@Failure		401			{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]					"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]					"Project not found"
//	@Failure		500			{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/projects/{project_id} [get]/
func GetProjectDetail(c *gin.Context) {
	projectIdStr := c.Param(consts.URLPathProjectID)
	projectID, err := strconv.Atoi(projectIdStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	resp, err := producer.GetProjectDetail(projectID)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			page		query		int													false	"Page number"	default(1)
//	@Param			size		query		int													false	"Page size"		default(20)
//	@Param			is_public	query		bool												false	"Filter by public status"
//	@Param			status		query		int													false	"Filter by status"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[dto.ProjectResp]]	"Projects retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]							"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/projects [get]
func ListProjects(c *gin.Context) {
	var req dto.ListProjectReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListProjects(&req)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			project_id	path		int										true	"Project ID"
//	@Param			request		body		dto.UpdateProjectReq					true	"Project update request"
//	@Success		202			{object}	dto.GenericResponse[dto.ProjectResp]	"Project updated successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]				"Invalid project ID or invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]				"Project not found"
//	@Failure		500			{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/projects/{project_id} [patch]
func UpdateProject(c *gin.Context) {
	projectIdStr := c.Param(consts.URLPathProjectID)
	projectID, err := strconv.Atoi(projectIdStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req dto.UpdateProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.UpdateProject(&req, projectID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusAccepted, "Project updated successfully", resp)
}

// ===================== Project-Label API =====================

// ManageProjectCustomLabels manages project custom labels (key-value pairs)
//
//	@Summary		Manage project custom labels
//	@Description	Add or remove custom labels (key-value pairs) for a project
//	@Tags			Projects
//	@ID				update_project_labels
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			project_id	path		int										true	"Project ID"
//	@Param			manage		body		dto.ManageProjectLabelReq				true	"Label management request"
//	@Success		200			{object}	dto.GenericResponse[dto.ProjectResp]	"Labels managed successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]				"Invalid project ID or invalid request format/parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]				"Project not found"
//	@Failure		500			{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/projects/{project_id}/labels [patch]
func ManageProjectCustomLabels(c *gin.Context) {
	projectIDStr := c.Param(consts.URLPathProjectID)
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil || projectID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req dto.ManageProjectLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ManageProjectLabels(&req, projectID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}
