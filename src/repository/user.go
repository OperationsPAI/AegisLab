package repository

import (
	"errors"
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
	"gorm.io/gorm"
)

// CreateUser 创建用户
func CreateUser(user *database.User) error {
	if err := database.DB.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}
	return nil
}

// GetUserByID 根据ID获取用户
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

// GetUserByUsername 根据用户名获取用户
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

// GetUserByEmail 根据邮箱获取用户
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

// UpdateUser 更新用户信息
func UpdateUser(user *database.User) error {
	if err := database.DB.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}
	return nil
}

// DeleteUser 软删除用户（设置状态为-1）
func DeleteUser(id int) error {
	if err := database.DB.Model(&database.User{}).Where("id = ?", id).Update("status", -1).Error; err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	return nil
}

// ListUsers 获取用户列表
func ListUsers(page, pageSize int, status *int) ([]database.User, int64, error) {
	var users []database.User
	var total int64

	query := database.DB.Model(&database.User{})

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %v", err)
	}

	return users, total, nil
}

// UpdateUserLoginTime 更新用户最后登录时间
func UpdateUserLoginTime(userID int) error {
	now := database.DB.NowFunc()
	if err := database.DB.Model(&database.User{}).Where("id = ?", userID).Update("last_login_at", now).Error; err != nil {
		return fmt.Errorf("failed to update user login time: %v", err)
	}
	return nil
}

// GetUserRoles 获取用户的全局角色
func GetUserRoles(userID int) ([]database.Role, error) {
	var roles []database.Role
	if err := database.DB.Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.status = 1", userID).
		Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to get user roles: %v", err)
	}
	return roles, nil
}

// GetUserProjectRoles 获取用户在特定项目中的角色
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

// GetUserProjects 获取用户参与的项目
func GetUserProjects(userID int) ([]database.UserProject, error) {
	var userProjects []database.UserProject
	if err := database.DB.Preload("Project").Preload("Role").
		Where("user_id = ? AND status = 1", userID).
		Find(&userProjects).Error; err != nil {
		return nil, fmt.Errorf("failed to get user projects: %v", err)
	}
	return userProjects, nil
}

// AddUserToProject 将用户添加到项目
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

// RemoveUserFromProject 将用户从项目中移除
func RemoveUserFromProject(userID, projectID int) error {
	if err := database.DB.Model(&database.UserProject{}).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Update("status", -1).Error; err != nil {
		return fmt.Errorf("failed to remove user from project: %v", err)
	}
	return nil
}

// AssignRoleToUser 给用户分配全局角色
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

// RemoveRoleFromUser 移除用户的全局角色
func RemoveRoleFromUser(userID, roleID int) error {
	if err := database.DB.Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&database.UserRole{}).Error; err != nil {
		return fmt.Errorf("failed to remove role from user: %v", err)
	}
	return nil
}
