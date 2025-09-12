package repository

import (
	"fmt"
	"time"

	"github.com/LGU-SE-Internal/rcabench/database"
)

// RelationInfo represents a relationship between entities
type RelationInfo struct {
	Type       string    `json:"type"`        // Type of relation (user-role, role-permission, etc.)
	SourceType string    `json:"source_type"` // Source entity type
	SourceID   int       `json:"source_id"`   // Source entity ID
	SourceName string    `json:"source_name"` // Source entity name
	TargetType string    `json:"target_type"` // Target entity type
	TargetID   int       `json:"target_id"`   // Target entity ID
	TargetName string    `json:"target_name"` // Target entity name
	CreatedAt  time.Time `json:"created_at"`  // When the relation was created
	CreatedBy  *int      `json:"created_by"`  // Who created the relation
	ProjectID  *int      `json:"project_id"`  // Project context (if applicable)
}

// GetUserRoleRelations returns user-role relationships
func GetUserRoleRelations(page, pageSize int, userID, roleID *int) ([]RelationInfo, int64, error) {
	query := database.DB.Table("user_roles ur").
		Select(`
			'user-role' as type,
			'user' as source_type,
			ur.user_id as source_id,
			u.username as source_name,
			'role' as target_type,
			ur.role_id as target_id,
			r.name as target_name,
			ur.created_at,
			ur.created_by,
			ur.project_id
		`).
		Joins("LEFT JOIN users u ON ur.user_id = u.id").
		Joins("LEFT JOIN roles r ON ur.role_id = r.id").
		Where("u.status != -1 AND r.status != -1")

	if userID != nil {
		query = query.Where("ur.user_id = ?", *userID)
	}
	if roleID != nil {
		query = query.Where("ur.role_id = ?", *roleID)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user-role relations: %v", err)
	}

	// Get paginated results
	var relations []RelationInfo
	offset := (page - 1) * pageSize
	if err := query.Order("ur.created_at DESC").Offset(offset).Limit(pageSize).Find(&relations).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get user-role relations: %v", err)
	}

	return relations, total, nil
}

// GetRolePermissionRelations returns role-permission relationships
func GetRolePermissionRelations(page, pageSize int, roleID, permissionID *int) ([]RelationInfo, int64, error) {
	query := database.DB.Table("role_permissions rp").
		Select(`
			'role-permission' as type,
			'role' as source_type,
			rp.role_id as source_id,
			r.name as source_name,
			'permission' as target_type,
			rp.permission_id as target_id,
			p.name as target_name,
			rp.created_at,
			rp.created_by,
			NULL as project_id
		`).
		Joins("LEFT JOIN roles r ON rp.role_id = r.id").
		Joins("LEFT JOIN permissions p ON rp.permission_id = p.id").
		Where("r.status != -1 AND p.status != -1")

	if roleID != nil {
		query = query.Where("rp.role_id = ?", *roleID)
	}
	if permissionID != nil {
		query = query.Where("rp.permission_id = ?", *permissionID)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count role-permission relations: %v", err)
	}

	// Get paginated results
	var relations []RelationInfo
	offset := (page - 1) * pageSize
	if err := query.Order("rp.created_at DESC").Offset(offset).Limit(pageSize).Find(&relations).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get role-permission relations: %v", err)
	}

	return relations, total, nil
}

// GetUserPermissionRelations returns user-permission relationships (direct assignments)
func GetUserPermissionRelations(page, pageSize int, userID, permissionID *int) ([]RelationInfo, int64, error) {
	query := database.DB.Table("user_permissions up").
		Select(`
			'user-permission' as type,
			'user' as source_type,
			up.user_id as source_id,
			u.username as source_name,
			'permission' as target_type,
			up.permission_id as target_id,
			p.name as target_name,
			up.created_at,
			up.created_by,
			up.project_id
		`).
		Joins("LEFT JOIN users u ON up.user_id = u.id").
		Joins("LEFT JOIN permissions p ON up.permission_id = p.id").
		Where("u.status != -1 AND p.status != -1")

	if userID != nil {
		query = query.Where("up.user_id = ?", *userID)
	}
	if permissionID != nil {
		query = query.Where("up.permission_id = ?", *permissionID)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user-permission relations: %v", err)
	}

	// Get paginated results
	var relations []RelationInfo
	offset := (page - 1) * pageSize
	if err := query.Order("up.created_at DESC").Offset(offset).Limit(pageSize).Find(&relations).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get user-permission relations: %v", err)
	}

	return relations, total, nil
}

// GetAllRelations returns all types of relationships with pagination
func GetAllRelations(page, pageSize int, relationType string) ([]RelationInfo, int64, error) {
	var allRelations []RelationInfo
	var totalCount int64

	switch relationType {
	case "user-role":
		relations, count, err := GetUserRoleRelations(page, pageSize, nil, nil)
		if err != nil {
			return nil, 0, err
		}
		return relations, count, nil

	case "role-permission":
		relations, count, err := GetRolePermissionRelations(page, pageSize, nil, nil)
		if err != nil {
			return nil, 0, err
		}
		return relations, count, nil

	case "user-permission":
		relations, count, err := GetUserPermissionRelations(page, pageSize, nil, nil)
		if err != nil {
			return nil, 0, err
		}
		return relations, count, nil

	case "all":
		// Combine all relation types
		userRoles, _, err := GetUserRoleRelations(1, 1000, nil, nil) // Get more for combining
		if err != nil {
			return nil, 0, err
		}
		allRelations = append(allRelations, userRoles...)

		rolePerms, _, err := GetRolePermissionRelations(1, 1000, nil, nil)
		if err != nil {
			return nil, 0, err
		}
		allRelations = append(allRelations, rolePerms...)

		userPerms, _, err := GetUserPermissionRelations(1, 1000, nil, nil)
		if err != nil {
			return nil, 0, err
		}
		allRelations = append(allRelations, userPerms...)

		totalCount = int64(len(allRelations))

		// Manual pagination for combined results
		start := (page - 1) * pageSize
		end := start + pageSize
		if start >= len(allRelations) {
			return []RelationInfo{}, totalCount, nil
		}
		if end > len(allRelations) {
			end = len(allRelations)
		}
		return allRelations[start:end], totalCount, nil

	default:
		return nil, 0, fmt.Errorf("unsupported relation type: %s", relationType)
	}
}

// GetRelationStatistics returns statistics about all relationships
func GetRelationStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// User-Role relations
	var userRoleCount int64
	if err := database.DB.Model(&database.UserRole{}).Count(&userRoleCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count user-role relations: %v", err)
	}
	stats["user_roles"] = userRoleCount

	// Role-Permission relations
	var rolePermCount int64
	if err := database.DB.Model(&database.RolePermission{}).Count(&rolePermCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count role-permission relations: %v", err)
	}
	stats["role_permissions"] = rolePermCount

	// User-Permission relations (direct)
	var userPermCount int64
	if err := database.DB.Model(&database.UserPermission{}).Count(&userPermCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count user-permission relations: %v", err)
	}
	stats["user_permissions"] = userPermCount

	// User-Project relations
	var userProjectCount int64
	if err := database.DB.Model(&database.UserProject{}).Where("status != -1").Count(&userProjectCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count user-project relations: %v", err)
	}
	stats["user_projects"] = userProjectCount

	// Dataset-Label relations
	var datasetLabelCount int64
	if err := database.DB.Model(&database.DatasetLabel{}).Count(&datasetLabelCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count dataset-label relations: %v", err)
	}
	stats["dataset_labels"] = datasetLabelCount

	// Container-Label relations
	var containerLabelCount int64
	if err := database.DB.Model(&database.ContainerLabel{}).Count(&containerLabelCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count container-label relations: %v", err)
	}
	stats["container_labels"] = containerLabelCount

	// Project-Label relations
	var projectLabelCount int64
	if err := database.DB.Model(&database.ProjectLabel{}).Count(&projectLabelCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count project-label relations: %v", err)
	}
	stats["project_labels"] = projectLabelCount

	// FaultInjection-Label relations
	var faultInjectionLabelCount int64
	if err := database.DB.Model(&database.FaultInjectionLabel{}).Count(&faultInjectionLabelCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count fault injection-label relations: %v", err)
	}
	stats["fault_injection_labels"] = faultInjectionLabelCount

	// Total relations
	stats["total_relations"] = userRoleCount + rolePermCount + userPermCount + userProjectCount + datasetLabelCount + containerLabelCount + projectLabelCount + faultInjectionLabelCount

	// Most used roles
	type RoleUsage struct {
		RoleName string `json:"role_name"`
		Count    int64  `json:"count"`
	}
	var topRoles []RoleUsage
	err := database.DB.Table("user_roles ur").
		Select("r.name as role_name, COUNT(*) as count").
		Joins("LEFT JOIN roles r ON ur.role_id = r.id").
		Where("r.status != -1").
		Group("r.name").
		Order("count DESC").
		Limit(5).
		Find(&topRoles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get top roles: %v", err)
	}
	stats["top_roles"] = topRoles

	// Most used permissions
	type PermissionUsage struct {
		PermissionName string `json:"permission_name"`
		Count          int64  `json:"count"`
	}
	var topPermissions []PermissionUsage
	err = database.DB.Table("role_permissions rp").
		Select("p.name as permission_name, COUNT(*) as count").
		Joins("LEFT JOIN permissions p ON rp.permission_id = p.id").
		Where("p.status != -1").
		Group("p.name").
		Order("count DESC").
		Limit(5).
		Find(&topPermissions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get top permissions: %v", err)
	}
	stats["top_permissions"] = topPermissions

	return stats, nil
}
