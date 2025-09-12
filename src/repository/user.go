package repository

import (
	"errors"
	"fmt"

	"rcabench/database"
	"gorm.io/gorm"
)

// CreateUser creates a user
func CreateUser(user *database.User) error {
	if err := database.DB.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}
	return nil
}

// GetUserByID gets a user by ID
func GetUserByID(id int) (*database.User, error) {
	var user database.User
	if err := database.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return &user, nil
}

// GetUserByUsername gets a user by username
func GetUserByUsername(username string) (*database.User, error) {
	var user database.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user '%s' not found", username)
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return &user, nil
}

// GetUserByEmail gets a user by email
func GetUserByEmail(email string) (*database.User, error) {
	var user database.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with email '%s' not found", email)
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return &user, nil
}

// GetUserProjectsMap gets all projects for multiple users in batch (optimized)
func GetUserProjectsMap(userIDs []int) (map[int][]database.UserProject, error) {
	if len(userIDs) == 0 {
		return make(map[int][]database.UserProject), nil
	}

	var relations []database.UserProject
	if err := database.DB.Preload("Project").Preload("Role").
		Where("user_id IN ?", userIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get user project relations: %v", err)
	}

	projectsMap := make(map[int][]database.UserProject)
	for _, relation := range relations {
		projectsMap[relation.UserID] = append(projectsMap[relation.UserID], relation)
	}

	for _, id := range userIDs {
		if _, exists := projectsMap[id]; !exists {
			projectsMap[id] = []database.UserProject{}
		}
	}

	return projectsMap, nil
}

// GetUserRolesMap gets all roles for multiple users in batch (optimized)
func GetUserRolesMap(userIDs []int) (map[int][]database.Role, error) {
	if len(userIDs) == 0 {
		return make(map[int][]database.Role), nil
	}

	var relations []database.UserRole
	if err := database.DB.Preload("Role").
		Where("user_id IN ?", userIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get user role relations: %v", err)
	}

	rolesMap := make(map[int][]database.Role)
	for _, relation := range relations {
		if relation.Role != nil {
			rolesMap[relation.UserID] = append(rolesMap[relation.UserID], *relation.Role)
		}
	}

	for _, id := range userIDs {
		if _, exists := rolesMap[id]; !exists {
			rolesMap[id] = []database.Role{}
		}
	}

	return rolesMap, nil
}

// UpdateUser updates user information
func UpdateUser(user *database.User) error {
	if err := database.DB.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}
	return nil
}

// DeleteUser soft deletes a user (sets status to -1)
func DeleteUser(id int) error {
	if err := database.DB.Model(&database.User{}).Where("id = ?", id).Update("status", -1).Error; err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	return nil
}

// ListUsers gets a list of users
func ListUsers(page, pageSize int, status *int) ([]database.User, int64, error) {
	var users []database.User
	var total int64

	query := database.DB.Model(&database.User{})

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %v", err)
	}

	// Paginated query
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %v", err)
	}

	return users, total, nil
}

// UpdateUserLoginTime updates user's last login time
func UpdateUserLoginTime(userID int) error {
	now := database.DB.NowFunc()
	if err := database.DB.Model(&database.User{}).Where("id = ?", userID).Update("last_login_at", now).Error; err != nil {
		return fmt.Errorf("failed to update user login time: %v", err)
	}
	return nil
}

// GetUserRoles gets user's global roles (optimized)
func GetUserRoles(userID int) ([]database.Role, error) {
	rolesMap, err := GetUserRolesMap([]int{userID})
	if err != nil {
		return nil, err
	}
	return rolesMap[userID], nil
}

// GetUserProjectRoles gets user's roles in a specific project
func GetUserProjectRoles(userID, projectID int) ([]database.Role, error) {
	var roles []database.Role
	if err := database.DB.Table("roles").
		Joins("JOIN user_projects ON roles.id = user_projects.role_id").
		Where("user_projects.user_id = ? AND user_projects.project_id = ? AND user_projects.status = 1", userID, projectID).
		Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to get user project roles: %v", err)
	}
	return roles, nil
}

// GetUserProjects gets projects the user participates in (optimized)
func GetUserProjects(userID int) ([]database.UserProject, error) {
	projectsMap, err := GetUserProjectsMap([]int{userID})
	if err != nil {
		return nil, err
	}
	// Filter active projects
	var activeProjects []database.UserProject
	for _, project := range projectsMap[userID] {
		if project.Status == 1 {
			activeProjects = append(activeProjects, project)
		}
	}
	return activeProjects, nil
}

// AddUserToProject adds a user to a project
func AddUserToProject(userID, projectID, roleID int) error {
	userProject := &database.UserProject{
		UserID:    userID,
		ProjectID: projectID,
		RoleID:    roleID,
		Status:    1,
	}

	if err := database.DB.Create(userProject).Error; err != nil {
		return fmt.Errorf("failed to add user to project: %v", err)
	}
	return nil
}

// RemoveUserFromProject removes a user from a project
func RemoveUserFromProject(userID, projectID int) error {
	if err := database.DB.Model(&database.UserProject{}).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Update("status", -1).Error; err != nil {
		return fmt.Errorf("failed to remove user from project: %v", err)
	}
	return nil
}

// AssignRoleToUser assigns a global role to a user
func AssignRoleToUser(userID, roleID int) error {
	userRole := &database.UserRole{
		UserID: userID,
		RoleID: roleID,
	}

	if err := database.DB.Create(userRole).Error; err != nil {
		return fmt.Errorf("failed to assign role to user: %v", err)
	}
	return nil
}

// RemoveRoleFromUser removes a global role from a user
func RemoveRoleFromUser(userID, roleID int) error {
	if err := database.DB.Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&database.UserRole{}).Error; err != nil {
		return fmt.Errorf("failed to remove role from user: %v", err)
	}
	return nil
}
