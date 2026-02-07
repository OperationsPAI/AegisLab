package producer

import (
	"errors"
	"fmt"

	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"

	"gorm.io/gorm"
)

// CreateTeam creates a new team
func CreateTeam(req *dto.CreateTeamReq, userID int) (*dto.TeamResp, error) {
	team := req.ConvertToTeam()

	// Get super_admin role
	superAdminRole, err := repository.GetRoleByName(database.DB, consts.RoleSuperAdmin.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get super_admin role: %w", err)
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := repository.CreateTeam(tx, team); err != nil {
			if errors.Is(err, consts.ErrAlreadyExists) {
				return consts.ErrAlreadyExists
			}
			return fmt.Errorf("failed to create team: %w", err)
		}

		// Add creator as team admin
		userTeam := &database.UserTeam{
			UserID: userID,
			TeamID: team.ID,
			RoleID: superAdminRole.ID,
			Status: consts.CommonEnabled,
		}
		if err := repository.CreateUserTeam(tx, userTeam); err != nil {
			return fmt.Errorf("failed to add creator to team: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return dto.NewTeamResp(team), nil
}

// DeleteTeam soft deletes a team
func DeleteTeam(teamID int) error {
	rowsAffected, err := repository.DeleteTeam(database.DB, teamID)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return consts.ErrNotFound
	}
	return nil
}

// GetTeamDetail retrieves detailed team information
func GetTeamDetail(teamID int) (*dto.TeamDetailResp, error) {
	team, err := repository.GetTeamByID(database.DB, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, consts.ErrNotFound
		}
		return nil, err
	}

	resp := dto.NewTeamDetailResp(team)

	// Get user count
	userCount, err := repository.GetTeamUserCount(database.DB, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team user count: %w", err)
	}
	resp.UserCount = userCount

	// Get project count
	projectCount, err := repository.GetTeamProjectCount(database.DB, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team project count: %w", err)
	}
	resp.ProjectCount = projectCount

	return resp, nil
}

// ListTeams lists teams with pagination and filtering
func ListTeams(req *dto.ListTeamReq, userID int, isAdmin bool) (*dto.ListResp[dto.TeamResp], error) {
	var teamIDs []int
	if !isAdmin {
		userTeams, err := repository.ListUserTeamsByUserID(database.DB, userID, consts.CommonEnabled)
		if err != nil {
			return nil, fmt.Errorf("failed to get user teams: %w", err)
		}
		for _, ut := range userTeams {
			teamIDs = append(teamIDs, ut.TeamID)
		}
	}

	limit, offset := req.ToGormParams()
	teams, total, err := repository.ListTeams(database.DB, limit, offset, req.IsPublic, req.Status, teamIDs)
	if err != nil {
		return nil, err
	}

	items := make([]dto.TeamResp, len(teams))
	for i, team := range teams {
		items[i] = *dto.NewTeamResp(&team)
	}

	resp := dto.ListResp[dto.TeamResp]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// UpdateTeam updates team information
func UpdateTeam(req *dto.UpdateTeamReq, teamID int) (*dto.TeamResp, error) {
	team, err := repository.GetTeamByID(database.DB, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, consts.ErrNotFound
		}
		return nil, err
	}

	req.PatchTeamModel(team)

	if err := repository.UpdateTeam(database.DB, team); err != nil {
		return nil, err
	}

	return dto.NewTeamResp(team), nil
}

// ListTeamProjects lists all projects belonging to a team
func ListTeamProjects(req *dto.ListProjectReq, teamID int) (*dto.ListResp[dto.ProjectResp], error) {
	// Get paginated projects
	limit, offset := req.ToGormParams()
	projects, total, err := repository.ListProjectsByTeamID(database.DB, teamID, limit, offset, req.IsPublic, req.Status)
	if err != nil {
		return nil, err
	}

	projectIDs := make([]int, 0, len(projects))
	for _, p := range projects {
		projectIDs = append(projectIDs, p.ID)
	}

	statsMap, err := repository.BatchGetProjectStatistics(database.DB, projectIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get project statistics: %w", err)
	}

	projectResps := make([]dto.ProjectResp, 0, len(projects))
	for i := range projects {
		stats := statsMap[projects[i].ID]
		projectResps = append(projectResps, *dto.NewProjectResp(&projects[i], stats))
	}

	resp := dto.ListResp[dto.ProjectResp]{
		Items:      projectResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// AddTeamMember adds a user to team
func AddTeamMember(req *dto.AddTeamMemberReq, teamID int) error {
	// Verify team exists
	_, err := repository.GetTeamByID(database.DB, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return consts.ErrNotFound
		}
		return err
	}

	// Get user by username
	user, err := repository.GetUserByUsername(database.DB, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found: %s", req.Username)
		}
		return err
	}

	// Verify role exists
	_, err = repository.GetRoleByID(database.DB, req.RoleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("role not found")
		}
		return err
	}

	userTeam := &database.UserTeam{
		UserID: user.ID,
		TeamID: teamID,
		RoleID: req.RoleID,
		Status: consts.CommonEnabled,
	}

	if err := repository.CreateUserTeam(database.DB, userTeam); err != nil {
		if errors.Is(err, consts.ErrAlreadyExists) {
			return consts.ErrAlreadyExists
		}
		return err
	}

	return nil
}

// RemoveTeamMember removes a user from team (only admin can remove others, cannot remove self)
func RemoveTeamMember(teamID, currentUserID, targetUserID int) error {
	// Cannot remove self
	if targetUserID == currentUserID {
		return fmt.Errorf("cannot remove yourself from the team")
	}

	// Verify team exists
	_, err := repository.GetTeamByID(database.DB, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return consts.ErrNotFound
		}
		return err
	}

	// Remove user from team
	rowsAffected, err := repository.DeleteUserTeam(database.DB, targetUserID, teamID)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user is not a member of this team")
	}

	return nil
}

// UpdateTeamMemberRole updates a team member's role (only admin can do this)
func UpdateTeamMemberRole(req *dto.UpdateTeamMemberRoleReq, teamID, targetUserID, currentUserID int) error {
	// Verify team exists
	_, err := repository.GetTeamByID(database.DB, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return consts.ErrNotFound
		}
		return err
	}

	// Verify new role exists
	_, err = repository.GetRoleByID(database.DB, req.RoleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("role not found")
		}
		return err
	}

	// Get existing user-team association
	userTeams, err := repository.ListUserTeamsByUserID(database.DB, targetUserID)
	if err != nil {
		return err
	}

	var targetUserTeam *database.UserTeam
	for i := range userTeams {
		if userTeams[i].TeamID == teamID {
			targetUserTeam = &userTeams[i]
			break
		}
	}

	if targetUserTeam == nil {
		return fmt.Errorf("user is not a member of this team")
	}

	// Update role
	targetUserTeam.RoleID = req.RoleID
	if err := database.DB.Save(targetUserTeam).Error; err != nil {
		return fmt.Errorf("failed to update team member role: %w", err)
	}

	return nil
}

// ListTeamMembers lists all members of a team with pagination
func ListTeamMembers(req *dto.ListTeamMemberReq, teamID int) (*dto.ListResp[dto.TeamMemberResp], error) {
	// Get paginated team members
	limit, offset := req.ToGormParams()
	users, total, err := repository.ListUsersByTeamID(database.DB, teamID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Build response
	members := make([]dto.TeamMemberResp, 0, len(users))
	for _, user := range users {
		userTeams, err := repository.ListUserTeamsByUserID(database.DB, user.ID)
		if err != nil {
			return nil, err
		}

		for _, ut := range userTeams {
			if ut.TeamID == teamID && ut.Status == consts.CommonEnabled {
				member := dto.TeamMemberResp{
					UserID:   user.ID,
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
					RoleID:   ut.RoleID,
					JoinedAt: ut.CreatedAt,
				}

				if ut.Role != nil {
					member.RoleName = ut.Role.DisplayName
				}

				members = append(members, member)
				break
			}
		}
	}

	resp := dto.ListResp[dto.TeamMemberResp]{
		Items:      members,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// ============================================================================
// Team Permission Check Helper Functions (exported for middleware)
// ============================================================================

// IsUserInTeam checks if a user is a member of a team
func IsUserInTeam(userID, teamID int) (bool, error) {
	ut, err := repository.GetUserTeamRole(database.DB, userID, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return ut != nil, nil
}

// IsUserTeamAdmin checks if a user has team admin role in a specific team
func IsUserTeamAdmin(userID, teamID int) (bool, error) {
	ut, err := repository.GetUserTeamRole(database.DB, userID, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return ut != nil && ut.Role != nil && ut.Role.Name == consts.RoleTeamAdmin.String(), nil
}

// IsTeamPublic checks if a team is publicly accessible
func IsTeamPublic(teamID int) (bool, error) {
	team, err := repository.GetTeamByID(database.DB, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return team.IsPublic, nil
}
