package repository

import (
	"fmt"

	"aegis/consts"
	"aegis/database"

	"gorm.io/gorm"
)

const (
	teamOmitFields = "ActiveName"
)

// =====================================================================
// Team Repository Functions
// =====================================================================

// CreateTeam creates a new team
func CreateTeam(db *gorm.DB, team *database.Team) error {
	if err := db.Omit(teamOmitFields).Create(team).Error; err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}
	return nil
}

// DeleteTeam soft deletes a team by setting its status to deleted
func DeleteTeam(db *gorm.DB, teamID int) (int64, error) {
	result := db.Model(&database.Team{}).
		Where("id = ? AND status != ?", teamID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to soft delete team %d: %w", teamID, result.Error)
	}
	return result.RowsAffected, nil
}

// GetTeamByID retrieves a team by its ID
func GetTeamByID(db *gorm.DB, id int) (*database.Team, error) {
	var team database.Team
	if err := db.Where("id = ?", id).First(&team).Error; err != nil {
		return nil, fmt.Errorf("failed to find team with id %d: %w", id, err)
	}
	return &team, nil
}

// GetTeamByName retrieves a team by its name
func GetTeamByName(db *gorm.DB, name string) (*database.Team, error) {
	var team database.Team
	if err := db.Where("name = ? AND status != ?", name, consts.CommonDeleted).First(&team).Error; err != nil {
		return nil, fmt.Errorf("failed to find team with name %s: %w", name, err)
	}
	return &team, nil
}

// GetTeamUserCount gets the count of users in a team
func GetTeamUserCount(db *gorm.DB, teamID int) (int, error) {
	var count int64
	if err := db.Model(&database.UserTeam{}).
		Where("team_id = ? AND status = ?", teamID, consts.CommonEnabled).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count team users: %w", err)
	}
	return int(count), nil
}

// GetTeamProjectCount gets the count of projects in a team
func GetTeamProjectCount(db *gorm.DB, teamID int) (int, error) {
	var count int64
	if err := db.Model(&database.Project{}).
		Where("team_id = ? AND status != ?", teamID, consts.CommonDeleted).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count team projects: %w", err)
	}
	return int(count), nil
}

// ListTeams lists teams based on filter options
func ListTeams(db *gorm.DB, limit, offset int, isPublic *bool, status *consts.StatusType, ids []int) ([]database.Team, int64, error) {
	var teams []database.Team
	var total int64

	query := db.Model(&database.Team{})
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if len(ids) > 0 {
		query = query.Where("id IN ?", ids)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count teams: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Find(&teams).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list teams: %w", err)
	}

	return teams, total, nil
}

// UpdateTeam updates a team
func UpdateTeam(db *gorm.DB, team *database.Team) error {
	if err := db.Omit(teamOmitFields).Save(team).Error; err != nil {
		return fmt.Errorf("failed to update team: %w", err)
	}
	return nil
}

// ListProjectsByTeamID lists all projects belonging to a team with pagination and filtering
func ListProjectsByTeamID(db *gorm.DB, teamID int, limit, offset int, isPublic *bool, status *consts.StatusType) ([]database.Project, int64, error) {
	var projects []database.Project
	var total int64

	query := db.Model(&database.Project{}).Where("team_id = ? AND status != ?", teamID, consts.CommonDeleted)

	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count projects for team %d: %w", teamID, err)
	}

	if err := query.Limit(limit).Offset(offset).Find(&projects).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list projects for team %d: %w", teamID, err)
	}
	return projects, total, nil
}

// =====================================================================
// Team Member Relationship Functions
// =====================================================================

// CreateUserTeam creates a user-team association
func CreateUserTeam(db *gorm.DB, userTeam *database.UserTeam) error {
	if err := db.Omit(userTeamOmitFields).Create(userTeam).Error; err != nil {
		return fmt.Errorf("failed to create user-team association: %w", err)
	}
	return nil
}

// DeleteUserTeam deletes a user-team association
func DeleteUserTeam(db *gorm.DB, userID, teamID int) (int64, error) {
	result := db.Model(&database.UserTeam{}).
		Where("user_id = ? AND team_id = ? AND status != ?", userID, teamID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete user-team association: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// GetUserTeamRole retrieves a user's role in a specific team
func GetUserTeamRole(db *gorm.DB, userID int, teamID int) (*database.UserTeam, error) {
	var userTeam database.UserTeam
	if err := db.
		Preload("Role").
		Where("user_id = ? AND team_id = ? AND status = ?", userID, teamID, consts.CommonEnabled).
		First(&userTeam).Error; err != nil {
		return nil, err
	}
	return &userTeam, nil
}

// ListTeamsByUserID gets teams the user participates in
func ListTeamsByUserID(db *gorm.DB, userID int) ([]database.Team, error) {
	var teams []database.Team
	if err := db.Table("teams").
		Joins("JOIN user_teams ON teams.id = user_teams.team_id").
		Where("user_teams.user_id = ? AND user_teams.status = ? AND teams.status != ?", userID, consts.CommonEnabled, consts.CommonDeleted).
		Find(&teams).Error; err != nil {
		return nil, fmt.Errorf("failed to list teams for user %d: %w", userID, err)
	}
	return teams, nil
}

// ListUserTeamsByUserID gets user-team associations for a specific user
func ListUserTeamsByUserID(db *gorm.DB, userID int, status ...consts.StatusType) ([]database.UserTeam, error) {
	query := db.Preload("Team").Preload("Role")
	if len(status) == 0 {
		query = query.Where("user_id = ? AND status != ?", userID, consts.CommonDeleted)
	} else if len(status) == 1 {
		query = query.Where("user_id = ? AND status = ?", userID, status[0])
	} else {
		query = query.Where("user_id = ? AND status IN (?)", userID, status)
	}

	var userTeams []database.UserTeam
	if err := query.
		Where("user_id = ? AND status != ?", userID, consts.CommonDeleted).
		Find(&userTeams).Error; err != nil {
		return nil, fmt.Errorf("failed to list user-team associations for user %d: %w", userID, err)
	}

	return userTeams, nil
}

// ListUsersByTeamID gets users who are members of a specific team with pagination
func ListUsersByTeamID(db *gorm.DB, teamID int, limit, offset int) ([]database.User, int64, error) {
	var users []database.User
	var total int64

	query := db.Model(&database.User{}).
		Joins("JOIN user_teams ON users.id = user_teams.user_id").
		Where("user_teams.team_id = ? AND user_teams.status = ? AND users.status != ?", teamID, consts.CommonEnabled, consts.CommonDeleted)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users for team %d: %w", teamID, err)
	}

	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users for team %d: %w", teamID, err)
	}

	return users, total, nil
}

// RemoveUsersFromTeam deletes all user-team associations for a given team
func RemoveUsersFromTeam(db *gorm.DB, teamID int) (int64, error) {
	result := db.Model(&database.UserTeam{}).
		Where("team_id = ? AND status != ?", teamID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete all users from team %d: %w", teamID, result.Error)
	}
	return result.RowsAffected, nil
}

// RemoveTeamsFromRole deletes all user-team associations for a given role
func RemoveTeamsFromRole(db *gorm.DB, roleID int) (int64, error) {
	result := db.Model(&database.UserTeam{}).
		Where("role_id = ? AND status != ?", roleID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete all teams with role %d: %w", roleID, result.Error)
	}
	return result.RowsAffected, nil
}

// RemoveTeamsFromUser deletes all user-team associations for a given user
func RemoveTeamsFromUser(db *gorm.DB, userID int) (int64, error) {
	result := db.Model(&database.UserTeam{}).
		Where("user_id = ? AND status != ?", userID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete all teams from user %d: %w", userID, result.Error)
	}
	return result.RowsAffected, nil
}
