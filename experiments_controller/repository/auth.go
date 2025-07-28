package repository

import (
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
	"gorm.io/gorm"
)

// PermissionChecker 权限检查器结构体
type PermissionChecker struct {
	UserID    int
	ProjectID *int
}

// NewPermissionChecker 创建权限检查器
func NewPermissionChecker(userID int, projectID *int) *PermissionChecker {
	return &PermissionChecker{
		UserID:    userID,
		ProjectID: projectID,
	}
}

// HasPermission 检查是否有特定权限
func (pc *PermissionChecker) HasPermission(action, resourceName string) (bool, error) {
	return CheckUserPermission(pc.UserID, action, resourceName, pc.ProjectID)
}

// CanRead 检查是否有读取权限
func (pc *PermissionChecker) CanRead(resourceName string) (bool, error) {
	return pc.HasPermission("read", resourceName)
}

// CanWrite 检查是否有写入权限
func (pc *PermissionChecker) CanWrite(resourceName string) (bool, error) {
	return pc.HasPermission("write", resourceName)
}

// CanDelete 检查是否有删除权限
func (pc *PermissionChecker) CanDelete(resourceName string) (bool, error) {
	return pc.HasPermission("delete", resourceName)
}

// CanExecute 检查是否有执行权限
func (pc *PermissionChecker) CanExecute(resourceName string) (bool, error) {
	return pc.HasPermission("execute", resourceName)
}

// CanManage 检查是否有管理权限
func (pc *PermissionChecker) CanManage(resourceName string) (bool, error) {
	return pc.HasPermission("manage", resourceName)
}

// IsAdmin 检查是否是管理员
func (pc *PermissionChecker) IsAdmin() (bool, error) {
	roles, err := GetUserRoles(pc.UserID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.Name == "admin" || role.Name == "super_admin" {
			return true, nil
		}
	}

	return false, nil
}

// IsProjectAdmin 检查是否是项目管理员
func (pc *PermissionChecker) IsProjectAdmin() (bool, error) {
	if pc.ProjectID == nil {
		return false, nil
	}

	roles, err := GetUserProjectRoles(pc.UserID, *pc.ProjectID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.Name == "project_admin" || role.Name == "admin" {
			return true, nil
		}
	}

	return false, nil
}

// AuthResult 权限检查结果
type AuthResult struct {
	Allowed bool
	Reason  string
}

// CheckMultiplePermissions 批量检查权限
func (pc *PermissionChecker) CheckMultiplePermissions(permissions map[string]string) (map[string]AuthResult, error) {
	results := make(map[string]AuthResult)

	for key, permission := range permissions {
		// permission 格式: "action:resource"
		// 例如: "read:dataset", "write:project"
		parts := splitPermission(permission)
		if len(parts) != 2 {
			results[key] = AuthResult{Allowed: false, Reason: "invalid permission format"}
			continue
		}

		action, resource := parts[0], parts[1]
		allowed, err := pc.HasPermission(action, resource)
		if err != nil {
			results[key] = AuthResult{Allowed: false, Reason: fmt.Sprintf("error: %v", err)}
		} else {
			reason := "allowed"
			if !allowed {
				reason = "permission denied"
			}
			results[key] = AuthResult{Allowed: allowed, Reason: reason}
		}
	}

	return results, nil
}

// InitializeSystemData 初始化系统数据（角色、权限、资源）
func InitializeSystemData() error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 创建系统资源
		systemResources := []database.Resource{
			{Name: "project", DisplayName: "项目", Type: "table", Category: "core", IsSystem: true, Status: 1},
			{Name: "dataset", DisplayName: "数据集", Type: "table", Category: "core", IsSystem: true, Status: 1},
			{Name: "fault_injection", DisplayName: "故障注入", Type: "table", Category: "core", IsSystem: true, Status: 1},
			{Name: "container", DisplayName: "容器", Type: "table", Category: "core", IsSystem: true, Status: 1},
			{Name: "task", DisplayName: "任务", Type: "table", Category: "core", IsSystem: true, Status: 1},
			{Name: "user", DisplayName: "用户", Type: "table", Category: "admin", IsSystem: true, Status: 1},
			{Name: "role", DisplayName: "角色", Type: "table", Category: "admin", IsSystem: true, Status: 1},
			{Name: "permission", DisplayName: "权限", Type: "table", Category: "admin", IsSystem: true, Status: 1},
		}

		for _, resource := range systemResources {
			var existingResource database.Resource
			if err := tx.Where("name = ?", resource.Name).FirstOrCreate(&existingResource, resource).Error; err != nil {
				return fmt.Errorf("failed to create system resource %s: %v", resource.Name, err)
			}
		}

		// 创建系统权限
		actions := []string{"read", "write", "delete", "execute", "manage"}
		for _, resource := range systemResources {
			var resourceRecord database.Resource
			if err := tx.Where("name = ?", resource.Name).First(&resourceRecord).Error; err != nil {
				continue
			}

			for _, action := range actions {
				permission := database.Permission{
					Name:        fmt.Sprintf("%s_%s", action, resource.Name),
					DisplayName: fmt.Sprintf("%s %s", actionDisplayName(action), resource.DisplayName),
					Action:      action,
					ResourceID:  resourceRecord.ID,
					IsSystem:    true,
					Status:      1,
				}

				var existingPermission database.Permission
				if err := tx.Where("name = ?", permission.Name).FirstOrCreate(&existingPermission, permission).Error; err != nil {
					return fmt.Errorf("failed to create system permission %s: %v", permission.Name, err)
				}
			}
		}

		// 创建系统角色
		systemRoles := []database.Role{
			{Name: "super_admin", DisplayName: "超级管理员", Type: "system", IsSystem: true, Status: 1},
			{Name: "admin", DisplayName: "管理员", Type: "system", IsSystem: true, Status: 1},
			{Name: "project_admin", DisplayName: "项目管理员", Type: "system", IsSystem: true, Status: 1},
			{Name: "developer", DisplayName: "开发者", Type: "system", IsSystem: true, Status: 1},
			{Name: "viewer", DisplayName: "查看者", Type: "system", IsSystem: true, Status: 1},
		}

		for _, role := range systemRoles {
			var existingRole database.Role
			if err := tx.Where("name = ?", role.Name).FirstOrCreate(&existingRole, role).Error; err != nil {
				return fmt.Errorf("failed to create system role %s: %v", role.Name, err)
			}
		}

		// 为系统角色分配权限
		if err := assignSystemRolePermissions(tx); err != nil {
			return fmt.Errorf("failed to assign system role permissions: %v", err)
		}

		// 创建超级管理员用户和默认项目
		if err := initializeAdminUserAndProjects(tx); err != nil {
			return fmt.Errorf("failed to initialize admin user and projects: %v", err)
		}

		return nil
	})
}

// assignSystemRolePermissions 为系统角色分配权限
func assignSystemRolePermissions(tx *gorm.DB) error {
	// super_admin: 所有权限
	if err := assignAllPermissionsToRole(tx, "super_admin"); err != nil {
		return err
	}

	// admin: 除了用户管理的其他权限
	adminPermissions := []string{
		"read_project", "write_project", "delete_project", "manage_project",
		"read_dataset", "write_dataset", "delete_dataset", "manage_dataset",
		"read_fault_injection", "write_fault_injection", "delete_fault_injection", "execute_fault_injection",
		"read_container", "write_container", "delete_container", "manage_container",
		"read_task", "write_task", "delete_task", "execute_task",
		"read_role", "read_permission",
	}
	if err := assignPermissionsToRole(tx, "admin", adminPermissions); err != nil {
		return err
	}

	// project_admin: 项目相关权限
	projectAdminPermissions := []string{
		"read_project", "write_project", "manage_project",
		"read_dataset", "write_dataset", "delete_dataset",
		"read_fault_injection", "write_fault_injection", "delete_fault_injection", "execute_fault_injection",
		"read_container", "write_container",
		"read_task", "write_task", "execute_task",
	}
	if err := assignPermissionsToRole(tx, "project_admin", projectAdminPermissions); err != nil {
		return err
	}

	// developer: 开发者权限
	developerPermissions := []string{
		"read_project", "read_dataset", "write_dataset",
		"read_fault_injection", "write_fault_injection", "execute_fault_injection",
		"read_container", "read_task", "write_task", "execute_task",
	}
	if err := assignPermissionsToRole(tx, "developer", developerPermissions); err != nil {
		return err
	}

	// viewer: 只读权限
	viewerPermissions := []string{
		"read_project", "read_dataset", "read_fault_injection", "read_container", "read_task",
	}
	if err := assignPermissionsToRole(tx, "viewer", viewerPermissions); err != nil {
		return err
	}

	return nil
}

// 辅助函数
func splitPermission(permission string) []string {
	for i, char := range permission {
		if char == ':' {
			return []string{permission[:i], permission[i+1:]}
		}
	}
	return []string{permission}
}

func actionDisplayName(action string) string {
	switch action {
	case "read":
		return "查看"
	case "write":
		return "编辑"
	case "delete":
		return "删除"
	case "execute":
		return "执行"
	case "manage":
		return "管理"
	default:
		return action
	}
}

func assignAllPermissionsToRole(tx *gorm.DB, roleName string) error {
	var role database.Role
	if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
		return err
	}

	var permissions []database.Permission
	if err := tx.Where("is_system = true AND status = 1").Find(&permissions).Error; err != nil {
		return err
	}

	for _, permission := range permissions {
		rolePermission := database.RolePermission{
			RoleID:       role.ID,
			PermissionID: permission.ID,
		}
		if err := tx.FirstOrCreate(&rolePermission, rolePermission).Error; err != nil {
			return err
		}
	}

	return nil
}

func assignPermissionsToRole(tx *gorm.DB, roleName string, permissionNames []string) error {
	var role database.Role
	if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
		return err
	}

	for _, permName := range permissionNames {
		var permission database.Permission
		if err := tx.Where("name = ?", permName).First(&permission).Error; err != nil {
			continue // 忽略不存在的权限
		}

		rolePermission := database.RolePermission{
			RoleID:       role.ID,
			PermissionID: permission.ID,
		}
		if err := tx.FirstOrCreate(&rolePermission, rolePermission).Error; err != nil {
			return err
		}
	}

	return nil
}

// initializeAdminUserAndProjects 初始化超级管理员用户和默认项目
func initializeAdminUserAndProjects(tx *gorm.DB) error {
	// 1. 创建超级管理员用户
	adminUser := database.User{
		Username: "admin",
		Email:    "admin@rcabench.local",
		// 密码: admin123，使用项目标准的SHA256+盐值加密
		Password: "60c873a916c7659b9798e17015e9130c0cb9c9f4f7f7c022222c0b869243fd6b:98a126542e7a0e2bf0322965b28885e8eb628c605f3cd0228b74e3d36e5edeee",
		FullName: "系统管理员",
		Status:   1,
		IsActive: true,
	}

	var existingUser database.User
	if err := tx.Where("username = ?", adminUser.Username).FirstOrCreate(&existingUser, adminUser).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %v", err)
	}

	// 2. 为超级管理员分配超级管理员角色
	var superAdminRole database.Role
	if err := tx.Where("name = ?", "super_admin").First(&superAdminRole).Error; err != nil {
		return fmt.Errorf("failed to find super_admin role: %v", err)
	}

	userRole := database.UserRole{
		UserID: existingUser.ID,
		RoleID: superAdminRole.ID,
	}
	if err := tx.Where("user_id = ? AND role_id = ?", existingUser.ID, superAdminRole.ID).FirstOrCreate(&userRole).Error; err != nil {
		return fmt.Errorf("failed to assign super_admin role to admin user: %v", err)
	}

	// 3. 创建默认项目
	defaultProjects := []database.Project{
		{
			Name:        "Default Project",
			Description: "系统默认项目，用于初始化和测试",
			Status:      1,
		},
		{
			Name:        "Demo Project",
			Description: "演示项目，用于展示系统功能",
			Status:      1,
		},
	}

	for _, project := range defaultProjects {
		var existingProject database.Project
		if err := tx.Where("name = ?", project.Name).FirstOrCreate(&existingProject, project).Error; err != nil {
			return fmt.Errorf("failed to create project %s: %v", project.Name, err)
		}

		// 将超级管理员加入项目
		userProject := database.UserProject{
			UserID:    existingUser.ID,
			ProjectID: existingProject.ID,
			RoleID:    superAdminRole.ID,
			Status:    1,
		}
		if err := tx.Where("user_id = ? AND project_id = ?", existingUser.ID, existingProject.ID).FirstOrCreate(&userProject).Error; err != nil {
			return fmt.Errorf("failed to add admin user to project %s: %v", project.Name, err)
		}
	}

	return nil
}
