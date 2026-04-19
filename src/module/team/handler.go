package team

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

// CreateTeam handles team creation
//
//	@Summary		Create a new team
//	@Description	Create a new team with specified details
//	@Tags			Teams
//	@ID				create_team
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateTeamReq					true	"Team creation request"
//	@Success		201		{object}	dto.GenericResponse[TeamResp]	"Team created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]		"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]		"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]		"Permission denied"
//	@Failure		409		{object}	dto.GenericResponse[any]		"Team already exists"
//	@Failure		500		{object}	dto.GenericResponse[any]		"Internal server error"
//	@Router			/api/v2/teams [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) CreateTeam(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}
	var req CreateTeamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.CreateTeam(c.Request.Context(), &req, userID)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.JSONResponse(c, http.StatusCreated, "Team created successfully", resp)
}

// DeleteTeam handles team deletion
//
//	@Summary		Delete team
//	@Description	Delete a team
//	@Tags			Teams
//	@ID				delete_team
//	@Produce		json
//	@Security		BearerAuth
//	@Param			team_id	path		int							true	"Team ID"
//	@Success		204		{object}	dto.GenericResponse[any]	"Team deleted successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid team ID"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]	"Team not found"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/teams/{team_id} [delete]
//	@x-api-type		{"portal":"true"}
func (h *Handler) DeleteTeam(c *gin.Context) {
	teamID, ok := parseTeamID(c)
	if !ok {
		return
	}
	if httpx.HandleServiceError(c, h.service.DeleteTeam(c.Request.Context(), teamID)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusNoContent, "Team deleted successfully", nil)
}

// GetTeamDetail handles getting a single team by ID
//
//	@Summary		Get team by ID
//	@Description	Get detailed information about a specific team
//	@Tags			Teams
//	@ID				get_team_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			team_id	path		int									true	"Team ID"
//	@Success		200		{object}	dto.GenericResponse[TeamDetailResp]	"Team retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid team ID"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]			"Team not found"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/teams/{team_id} [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) GetTeamDetail(c *gin.Context) {
	teamID, ok := parseTeamID(c)
	if !ok {
		return
	}
	resp, err := h.service.GetTeamDetail(c.Request.Context(), teamID)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// ListTeams handles listing teams with pagination and filtering
//
//	@Summary		List teams
//	@Description	Get paginated list of teams with filtering
//	@Tags			Teams
//	@ID				list_teams
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int											false	"Page number"	default(1)
//	@Param			size		query		int											false	"Page size"		default(20)
//	@Param			is_public	query		bool										false	"Filter by public status"
//	@Param			status		query		consts.StatusType							false	"Filter by status"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[TeamResp]]	"Teams retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]					"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]					"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/teams [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) ListTeams(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}
	var req ListTeamReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.ListTeams(c.Request.Context(), &req, userID, middleware.IsCurrentUserAdmin(c))
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// UpdateTeam handles team updates
//
//	@Summary		Update team
//	@Description	Update an existing team's information
//	@Tags			Teams
//	@ID				update_team
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			team_id	path		int								true	"Team ID"
//	@Param			request	body		UpdateTeamReq					true	"Team update request"
//	@Success		202		{object}	dto.GenericResponse[TeamResp]	"Team updated successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]		"Invalid team ID or invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]		"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]		"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]		"Team not found"
//	@Failure		500		{object}	dto.GenericResponse[any]		"Internal server error"
//	@Router			/api/v2/teams/{team_id} [patch]
//	@x-api-type		{"portal":"true"}
func (h *Handler) UpdateTeam(c *gin.Context) {
	teamID, ok := parseTeamID(c)
	if !ok {
		return
	}
	var req UpdateTeamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.UpdateTeam(c.Request.Context(), &req, teamID)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.JSONResponse(c, http.StatusAccepted, "Team updated successfully", resp)
}

// ListTeamProjects lists all projects belonging to a team
//
//	@Summary		List team projects
//	@Description	Get paginated list of projects belonging to a specific team with filtering
//	@Tags			Teams
//	@ID				list_team_projects
//	@Produce		json
//	@Security		BearerAuth
//	@Param			team_id		path		int													true	"Team ID"
//	@Param			page		query		int													false	"Page number"	default(1)
//	@Param			size		query		int													false	"Page size"		default(20)
//	@Param			is_public	query		bool												false	"Filter by public status"
//	@Param			status		query		consts.StatusType									false	"Filter by status"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[TeamProjectItem]]	"Projects retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]							"Invalid team ID or request parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]							"Team not found"
//	@Failure		500			{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/teams/{team_id}/projects [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) ListTeamProjects(c *gin.Context) {
	teamID, ok := parseTeamID(c)
	if !ok {
		return
	}
	var req TeamProjectListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.ListTeamProjects(c.Request.Context(), &req, teamID)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// AddTeamMember adds a user to team
//
//	@Summary		Add member to team
//	@Description	Add a user to team by username
//	@Tags			Teams
//	@ID				add_team_member
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			team_id	path		int							true	"Team ID"
//	@Param			request	body		AddTeamMemberReq			true	"Add member request"
//	@Success		201		{object}	dto.GenericResponse[any]	"Member added successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid team ID or request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]	"Team or user not found"
//	@Failure		409		{object}	dto.GenericResponse[any]	"User already in team"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/teams/{team_id}/members [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) AddTeamMember(c *gin.Context) {
	teamID, ok := parseTeamID(c)
	if !ok {
		return
	}
	var req AddTeamMemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	if httpx.HandleServiceError(c, h.service.AddMember(c.Request.Context(), &req, teamID)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusCreated, "Member added successfully", nil)
}

// RemoveTeamMember removes a user from team
//
//	@Summary		Remove member from team
//	@Description	Remove a user from team (only admin can remove others, cannot remove self)
//	@Tags			Teams
//	@ID				remove_team_member
//	@Produce		json
//	@Security		BearerAuth
//	@Param			team_id	path		int							true	"Team ID"
//	@Param			user_id	path		int							true	"User ID to remove"
//	@Success		204		{object}	dto.GenericResponse[any]	"Member removed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid team ID or user ID, or cannot remove self"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied (only admin can remove members)"
//	@Failure		404		{object}	dto.GenericResponse[any]	"Team or user not found"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/teams/{team_id}/members/{user_id} [delete]
//	@x-api-type		{"portal":"true"}
func (h *Handler) RemoveTeamMember(c *gin.Context) {
	currentUserID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}
	teamID, ok := parseTeamID(c)
	if !ok {
		return
	}
	userID, ok := parseIntParam(c, "user_id", "Invalid user ID")
	if !ok {
		return
	}
	if currentUserID == userID {
		dto.ErrorResponse(c, http.StatusBadRequest, "Cannot remove yourself from the team")
		return
	}
	if httpx.HandleServiceError(c, h.service.RemoveMember(c.Request.Context(), teamID, currentUserID, userID)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusNoContent, "Member removed successfully", nil)
}

// UpdateTeamMemberRole updates a team member's role
//
//	@Summary		Update team member role
//	@Description	Update a team member's role (only admin can do this)
//	@Tags			Teams
//	@ID				update_team_member_role
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			team_id	path		int							true	"Team ID"
//	@Param			user_id	path		int							true	"User ID"
//	@Param			request	body		UpdateTeamMemberRoleReq		true	"Update role request"
//	@Success		200		{object}	dto.GenericResponse[any]	"Role updated successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid team ID, user ID, or request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied (only admin can update roles)"
//	@Failure		404		{object}	dto.GenericResponse[any]	"Team, user, or role not found"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/teams/{team_id}/members/{user_id}/role [patch]
//	@x-api-type		{"portal":"true"}
func (h *Handler) UpdateTeamMemberRole(c *gin.Context) {
	currentUserID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}
	teamID, ok := parseTeamID(c)
	if !ok {
		return
	}
	userID, ok := parseIntParam(c, "user_id", "Invalid user ID")
	if !ok {
		return
	}
	var req UpdateTeamMemberRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	if httpx.HandleServiceError(c, h.service.UpdateMemberRole(c.Request.Context(), &req, teamID, userID, currentUserID)) {
		return
	}
	dto.JSONResponse[any](c, http.StatusOK, "Role updated successfully", nil)
}

// ListTeamMembers lists all members of a team
//
//	@Summary		List team members
//	@Description	Get paginated list of members of a specific team
//	@Tags			Teams
//	@ID				list_team_members
//	@Produce		json
//	@Security		BearerAuth
//	@Param			team_id	path		int													true	"Team ID"
//	@Param			page	query		int													false	"Page number"	default(1)
//	@Param			size	query		int													false	"Page size"		default(20)
//	@Success		200		{object}	dto.GenericResponse[dto.ListResp[TeamMemberResp]]	"Members retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]							"Invalid team ID or request parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]							"Team not found"
//	@Failure		500		{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/teams/{team_id}/members [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) ListTeamMembers(c *gin.Context) {
	teamID, ok := parseTeamID(c)
	if !ok {
		return
	}
	var req ListTeamMemberReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.ListMembers(c.Request.Context(), &req, teamID)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

func parseTeamID(c *gin.Context) (int, bool) {
	return parseIntParam(c, consts.URLPathTeamID, "Invalid team ID")
}

func parseIntParam(c *gin.Context, key, msg string) (int, bool) {
	v := c.Param(key)
	id, err := strconv.Atoi(v)
	if err != nil || id <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, msg)
		return 0, false
	}
	return id, true
}
