package team

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	project "aegis/module/project"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) createTeamWithCreator(team *model.Team, userID int) error {
	var superAdminRole model.Role
	if err := r.db.Where("name = ? AND status != ?", consts.RoleSuperAdmin.String(), consts.CommonDeleted).
		First(&superAdminRole).Error; err != nil {
		return fmt.Errorf("failed to get super_admin role: %w", err)
	}

	if err := r.db.Omit("ActiveName").Create(team).Error; err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	if err := r.db.Omit("active_user_team").Create(&model.UserTeam{
		UserID: userID,
		TeamID: team.ID,
		RoleID: superAdminRole.ID,
		Status: consts.CommonEnabled,
	}).Error; err != nil {
		return fmt.Errorf("failed to create user-team association: %w", err)
	}
	return nil
}

func (r *Repository) loadTeamDetailBase(teamID int) (*model.Team, int, error) {
	team, err := r.loadTeam(teamID)
	if err != nil {
		return nil, 0, err
	}

	var userCount int64
	if err := r.db.Model(&model.UserTeam{}).
		Where("team_id = ? AND status = ?", teamID, consts.CommonEnabled).
		Count(&userCount).Error; err != nil {
		return nil, 0, err
	}

	return team, int(userCount), nil
}

func (r *Repository) listVisibleTeams(limit, offset int, req *ListTeamReq, userID int, isAdmin bool) ([]model.Team, int64, error) {
	var teamIDs []int
	if !isAdmin {
		if err := r.db.Model(&model.UserTeam{}).
			Where("user_id = ? AND status = ?", userID, consts.CommonEnabled).
			Pluck("team_id", &teamIDs).Error; err != nil {
			return nil, 0, err
		}
		if len(teamIDs) == 0 {
			return []model.Team{}, 0, nil
		}
	}

	var teams []model.Team
	var total int64

	query := r.db.Model(&model.Team{})
	if req.IsPublic != nil {
		query = query.Where("is_public = ?", *req.IsPublic)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if len(teamIDs) > 0 {
		query = query.Where("id IN ?", teamIDs)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count teams: %w", err)
	}
	if err := query.Limit(limit).Offset(offset).Find(&teams).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list teams: %w", err)
	}
	return teams, total, nil
}

func (r *Repository) updateMutableTeam(teamID int, patch func(*model.Team)) (*model.Team, error) {
	team, err := r.loadTeam(teamID)
	if err != nil {
		return nil, err
	}
	patch(team)
	if err := r.db.Omit("ActiveName").Save(team).Error; err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}
	return team, nil
}

func (r *Repository) listTeamProjectViews(teamID, limit, offset int, isPublic *bool, status *consts.StatusType) ([]model.Project, map[int]*dto.ProjectStatistics, int64, error) {
	var (
		projects []model.Project
		total    int64
	)

	query := r.db.Model(&model.Project{}).Where("team_id = ? AND status != ?", teamID, consts.CommonDeleted)
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, nil, 0, fmt.Errorf("failed to count projects for team %d: %w", teamID, err)
	}
	if err := query.Limit(limit).Offset(offset).Find(&projects).Error; err != nil {
		return nil, nil, 0, fmt.Errorf("failed to list projects for team %d: %w", teamID, err)
	}

	projectIDs := make([]int, 0, len(projects))
	for _, project := range projects {
		projectIDs = append(projectIDs, project.ID)
	}

	statsMap, err := project.NewRepository(r.db).ListProjectStatistics(projectIDs)
	if err != nil {
		return nil, nil, 0, err
	}
	return projects, statsMap, total, nil
}

func (r *Repository) addMember(teamID int, username string, roleID int) error {
	if _, err := r.loadTeam(teamID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return consts.ErrNotFound
		}
		return err
	}

	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return fmt.Errorf("failed to find user with username %s: %w", username, err)
	}
	var role model.Role
	if err := r.db.Where("id = ? AND status != ?", roleID, consts.CommonDeleted).First(&role).Error; err != nil {
		return fmt.Errorf("failed to find role with id %d: %w", roleID, err)
	}

	if err := r.db.Omit("active_user_team").Create(&model.UserTeam{
		UserID: user.ID,
		TeamID: teamID,
		RoleID: roleID,
		Status: consts.CommonEnabled,
	}).Error; err != nil {
		return fmt.Errorf("failed to create user-team association: %w", err)
	}
	return nil
}

func (r *Repository) removeMember(teamID, userID int) (int64, error) {
	if _, err := r.loadTeam(teamID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, consts.ErrNotFound
		}
		return 0, err
	}

	result := r.db.Model(&model.UserTeam{}).
		Where("user_id = ? AND team_id = ? AND status != ?", userID, teamID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete user-team association: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func (r *Repository) updateMemberRole(teamID, targetUserID, roleID int) error {
	if _, err := r.loadTeam(teamID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return consts.ErrNotFound
		}
		return err
	}
	var role model.Role
	if err := r.db.Where("id = ? AND status != ?", roleID, consts.CommonDeleted).First(&role).Error; err != nil {
		return fmt.Errorf("failed to find role with id %d: %w", roleID, err)
	}

	var userTeam model.UserTeam
	if err := r.db.Preload("Role").
		Where("user_id = ? AND team_id = ? AND status = ?", targetUserID, teamID, consts.CommonEnabled).
		First(&userTeam).Error; err != nil {
		return err
	}
	userTeam.RoleID = roleID
	return r.db.Save(&userTeam).Error
}

func (r *Repository) listTeamMembers(teamID, limit, offset int) ([]TeamMemberResp, int64, error) {
	if _, err := r.loadTeam(teamID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, consts.ErrNotFound
		}
		return nil, 0, err
	}

	var members []TeamMemberResp
	var total int64

	query := r.db.Table("users").
		Joins("JOIN user_teams ON users.id = user_teams.user_id").
		Joins("LEFT JOIN roles ON roles.id = user_teams.role_id").
		Where("user_teams.team_id = ? AND user_teams.status = ? AND users.status != ?", teamID, consts.CommonEnabled, consts.CommonDeleted)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count team members for team %d: %w", teamID, err)
	}
	if err := query.Select(
		"users.id AS user_id",
		"users.username",
		"users.full_name",
		"users.email",
		"user_teams.role_id",
		"roles.display_name AS role_name",
		"user_teams.created_at AS joined_at",
	).Limit(limit).Offset(offset).Scan(&members).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list team members for team %d: %w", teamID, err)
	}
	return members, total, nil
}

func (r *Repository) loadUserTeamMembership(userID, teamID int) (*model.UserTeam, error) {
	var userTeam model.UserTeam
	if err := r.db.
		Preload("Role").
		Where("user_id = ? AND team_id = ? AND status = ?", userID, teamID, consts.CommonEnabled).
		First(&userTeam).Error; err != nil {
		return nil, err
	}
	return &userTeam, nil
}

func (r *Repository) deleteTeam(teamID int) (int64, error) {
	result := r.db.Model(&model.Team{}).
		Where("id = ? AND status != ?", teamID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to soft delete team %d: %w", teamID, result.Error)
	}
	return result.RowsAffected, nil
}

func (r *Repository) isTeamPublic(teamID int) (bool, error) {
	team, err := r.loadTeam(teamID)
	if err != nil {
		return false, err
	}
	return team.IsPublic, nil
}

func (r *Repository) loadTeam(teamID int) (*model.Team, error) {
	var team model.Team
	if err := r.db.Where("id = ?", teamID).First(&team).Error; err != nil {
		return nil, fmt.Errorf("failed to find team with id %d: %w", teamID, err)
	}
	return &team, nil
}
