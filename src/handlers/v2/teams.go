package v2

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

// CreateTeam handles team creation
//
//	@Summary		Create a new team
//	@Description	Create a new team with specified details
//	@Tags			Teams
//	@ID				create_team
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateTeamReq					true	"Team creation request"
//	@Success		201		{object}	dto.GenericResponse[dto.TeamResp]	"Team created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		409		{object}	dto.GenericResponse[any]			"Team already exists"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/teams [post]
//	@x-api-type		{"sdk":"true"}
func CreateTeam(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req dto.CreateTeamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.CreateTeam(&req, userID)
	if handlers.HandleServiceError(c, err) {
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
func DeleteTeam(c *gin.Context) {
	teamIDStr := c.Param(consts.URLPathTeamID)
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	err = producer.DeleteTeam(teamID)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			team_id	path		int										true	"Team ID"
//	@Success		200		{object}	dto.GenericResponse[dto.TeamDetailResp]	"Team retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid team ID"
//	@Failure		401		{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]				"Team not found"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/teams/{team_id} [get]
//	@x-api-type		{"sdk":"true"}
func GetTeamDetail(c *gin.Context) {
	teamIDStr := c.Param(consts.URLPathTeamID)
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	resp, err := producer.GetTeamDetail(teamID)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			page		query		int												false	"Page number"	default(1)
//	@Param			size		query		int												false	"Page size"		default(20)
//	@Param			is_public	query		bool											false	"Filter by public status"
//	@Param			status		query		consts.StatusType								false	"Filter by status"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[dto.TeamResp]]	"Teams retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]						"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/teams [get]
//	@x-api-type		{"sdk":"true"}
func ListTeams(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	isAdmin := middleware.IsCurrentUserAdmin(c)

	var req dto.ListTeamReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListTeams(&req, userID, isAdmin)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			team_id	path		int									true	"Team ID"
//	@Param			request	body		dto.UpdateTeamReq					true	"Team update request"
//	@Success		202		{object}	dto.GenericResponse[dto.TeamResp]	"Team updated successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid team ID or invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]			"Team not found"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/teams/{team_id} [patch]
func UpdateTeam(c *gin.Context) {
	teamIDStr := c.Param(consts.URLPathTeamID)
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	var req dto.UpdateTeamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.UpdateTeam(&req, teamID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusAccepted, "Team updated successfully", resp)
}

// ===================== Team-Project API =====================

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
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[dto.ProjectResp]]	"Projects retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]							"Invalid team ID or request parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]							"Team not found"
//	@Failure		500			{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/teams/{team_id}/projects [get]
//	@x-api-type		{"sdk":"true"}
func ListTeamProjects(c *gin.Context) {
	teamIDStr := c.Param(consts.URLPathTeamID)
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	var req dto.ListProjectReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListTeamProjects(&req, teamID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// ===================== Team-User API =====================

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
//	@Param			request	body		dto.AddTeamMemberReq		true	"Add member request"
//	@Success		201		{object}	dto.GenericResponse[any]	"Member added successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid team ID or request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]	"Team or user not found"
//	@Failure		409		{object}	dto.GenericResponse[any]	"User already in team"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/teams/{team_id}/members [post]
func AddTeamMember(c *gin.Context) {
	teamIDStr := c.Param(consts.URLPathTeamID)
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	var req dto.AddTeamMemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	err = producer.AddTeamMember(&req, teamID)
	if handlers.HandleServiceError(c, err) {
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
func RemoveTeamMember(c *gin.Context) {
	currentUserID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	teamIDStr := c.Param(consts.URLPathTeamID)
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if currentUserID == userID {
		dto.ErrorResponse(c, http.StatusBadRequest, "Cannot remove yourself from the team")
		return
	}

	err = producer.RemoveTeamMember(teamID, currentUserID, userID)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			request	body		dto.UpdateTeamMemberRoleReq	true	"Update role request"
//	@Success		200		{object}	dto.GenericResponse[any]	"Role updated successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid team ID, user ID, or request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied (only admin can update roles)"
//	@Failure		404		{object}	dto.GenericResponse[any]	"Team, user, or role not found"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/teams/{team_id}/members/{user_id}/role [patch]
func UpdateTeamMemberRole(c *gin.Context) {
	currentUserID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	teamIDStr := c.Param(consts.URLPathTeamID)
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.UpdateTeamMemberRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	err = producer.UpdateTeamMemberRole(&req, teamID, userID, currentUserID)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			team_id	path		int														true	"Team ID"
//	@Param			page	query		int														false	"Page number"	default(1)
//	@Param			size	query		int														false	"Page size"		default(20)
//	@Success		200		{object}	dto.GenericResponse[dto.ListResp[dto.TeamMemberResp]]	"Members retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]								"Invalid team ID or request parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]								"Team not found"
//	@Failure		500		{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/teams/{team_id}/members [get]
//	@x-api-type		{"sdk":"true"}
func ListTeamMembers(c *gin.Context) {
	teamIDStr := c.Param(consts.URLPathTeamID)
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	var req dto.ListTeamMemberReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListTeamMembers(&req, teamID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}
