package initialization

import (
	"errors"
	"fmt"
	"path/filepath"

	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/repository"
	producer "aegis/service/producer"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
)

var resourceIDMap map[consts.ResourceName]int

func InitializeProducer() {
	user, err := repository.GetUserByUsername(database.DB, AdminUsername)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Info("Seeding initial system data for producer...")
			if err := initializeProducer(); err != nil {
				logrus.Fatalf("Failed to initialize system data for producer: %v", err)
			}
			logrus.Info("Successfully seeded initial system data for producer")
		} else {
			logrus.Fatalf("Failed to check for %s: %v", AdminUsername, err)
		}
	}

	if user != nil {
		logrus.Info("Initial system data for producer already seeded, skipping initialization")
	}
}

func initializeProducer() error {
	dataPath := config.GetString("initialization.data_path")
	filePath := filepath.Join(dataPath, consts.InitialFilename)
	initialData, err := loadInitialDataFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load initial data from file: %w", err)
	}

	resources := []database.Resource{
		{Name: consts.ResourceSystem, Type: consts.ResourceTypeSystem, Category: consts.ResourceAdmin},
		{Name: consts.ResourceAudit, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceConfigruation, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceContainer, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceContainerVersion, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceDataset, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceDatasetVersion, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceProject, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceLabel, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceUser, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceRole, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourcePermission, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceTask, Type: consts.ResourceTypeTable, Category: consts.ResourceCore},
		{Name: consts.ResourceTrace, Type: consts.ResourceTypeTable, Category: consts.ResourceCore},
		{Name: consts.ResourceInjection, Type: consts.ResourceTypeTable, Category: consts.ResourceCore},
		{Name: consts.ResourceExecution, Type: consts.ResourceTypeTable, Category: consts.ResourceCore},
	}

	for i := range resources {
		resources[i].DisplayName = consts.GetResourceDisplayName(resources[i].Name)
	}

	systemRoles := make([]database.Role, 0)
	for role, displayName := range consts.SystemRoleDisplayNames {
		systemRoles = append(systemRoles, database.Role{
			Name:        string(role),
			DisplayName: displayName,
			IsSystem:    true,
			Status:      consts.CommonEnabled,
		})
	}

	actions := make([]consts.ActionName, 0, len(consts.ValidActions))
	for action := range consts.ValidActions {
		actions = append(actions, action)
	}

	return withOptimizedDBSettings(func() error {
		return database.DB.Transaction(func(tx *gorm.DB) error {
			if err := repository.BatchUpsertResources(tx, resources); err != nil {
				return fmt.Errorf("failed to create system resources: %w", err)
			}

			resourceNames := make([]consts.ResourceName, 0, len(resources))
			for _, res := range resources {
				resourceNames = append(resourceNames, res.Name)
			}

			allResourcesInDB, err := repository.ListResourcesByNames(tx, resourceNames)
			if err != nil {
				return fmt.Errorf("failed to get system resources from database: %w", err)
			}

			if len(allResourcesInDB) != len(resources) {
				return fmt.Errorf("mismatch in number of resources created and fetched")
			}

			resourceMap := make(map[consts.ResourceName]*database.Resource, len(allResourcesInDB))
			resourceIDMap = make(map[consts.ResourceName]int, len(allResourcesInDB))
			for _, res := range allResourcesInDB {
				resourceIDMap[res.Name] = res.ID
				resourceMap[res.Name] = &res
			}

			resourceMap[consts.ResourceContainerVersion].ParentID = utils.IntPtr(resourceIDMap[consts.ResourceContainer])
			resourceMap[consts.ResourceDatasetVersion].ParentID = utils.IntPtr(resourceIDMap[consts.ResourceDataset])

			toUpdatedResources := []database.Resource{
				*resourceMap[consts.ResourceContainerVersion],
				*resourceMap[consts.ResourceDatasetVersion],
			}

			if err := repository.BatchUpsertResources(tx, toUpdatedResources); err != nil {
				return fmt.Errorf("failed to update resource parent IDs: %w", err)
			}

			var permissionsToCreate []database.Permission
			for _, resource := range allResourcesInDB {
				for _, action := range actions {
					permission := database.Permission{
						Name:        producer.GetPermissionName(action, resource.Name),
						DisplayName: producer.GetPermissionDisplayName(action, resource.DisplayName),
						Action:      string(action),
						ResourceID:  resourceIDMap[resource.Name],
						IsSystem:    true,
						Status:      consts.CommonEnabled,
					}
					permissionsToCreate = append(permissionsToCreate, permission)
				}
			}

			if err := repository.BatchUpsertPermissions(tx, permissionsToCreate); err != nil {
				return fmt.Errorf("failed to create system permissions: %w", err)
			}

			if err := repository.BatchUpsertRoles(tx, systemRoles); err != nil {
				return fmt.Errorf("failed to create system roles: %w", err)
			}

			if err := assignSystemRolePermissions(tx); err != nil {
				return fmt.Errorf("failed to assign system role permissions: %w", err)
			}

			adminUser, err := initializeAdminUserAndProjects(tx, initialData)
			if err != nil {
				return fmt.Errorf("failed to initialize admin user and projects: %w", err)
			}

			if err := initializeContainers(tx, initialData, adminUser.ID); err != nil {
				return fmt.Errorf("failed to initialize containers: %w", err)
			}

			if err := initializeDatasets(tx, initialData, adminUser.ID); err != nil {
				return fmt.Errorf("failed to initialize datasets: %w", err)
			}

			if err := initializeExecutionLabels(tx); err != nil {
				return fmt.Errorf("failed to initialize execution labels: %w", err)
			}

			return nil
		})
	})
}

func assignSystemRolePermissions(tx *gorm.DB) error {
	for roleName, permissionNames := range consts.SystemRolePermissions {
		role, err := repository.GetRoleByName(tx, roleName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("role %s not found", roleName)
			}
			return err
		}

		if roleName == consts.RoleSuperAdmin {
			permissions, err := repository.ListSystemPermissions(tx)
			if err != nil {
				return fmt.Errorf("failed to list system permissions: %w", err)
			}

			var rolePermissions []database.RolePermission
			for _, perm := range permissions {
				rolePermissions = append(rolePermissions, database.RolePermission{
					RoleID:       role.ID,
					PermissionID: perm.ID,
				})
			}

			if err := repository.BatchCreateRolePermissions(tx, rolePermissions); err != nil {
				return fmt.Errorf("failed to assign all permissions to super admin role: %w", err)
			}
		} else {
			var permissionStrs []string
			for _, name := range permissionNames {
				permissionStrs = append(permissionStrs, string(name))
			}

			permissions, err := repository.ListPermissionsByNames(tx, permissionStrs)
			if err != nil {
				return fmt.Errorf("failed to list permissions for role %s: %w", roleName, err)
			}

			var rolePermissions []database.RolePermission
			for _, perm := range permissions {
				rolePermissions = append(rolePermissions, database.RolePermission{
					RoleID:       role.ID,
					PermissionID: perm.ID,
				})
			}

			if err := repository.BatchCreateRolePermissions(tx, rolePermissions); err != nil {
				return fmt.Errorf("failed to assign permissions to role %s: %w", roleName, err)
			}
		}
	}

	return nil
}

func initializeAdminUserAndProjects(tx *gorm.DB, data *InitialData) (*database.User, error) {
	adminUser := data.AdminUser.ConvertToDBUser()

	if err := repository.CreateUser(tx, adminUser); err != nil {
		if errors.Is(err, consts.ErrAlreadyExists) {
			return nil, fmt.Errorf("admin user already exists")
		}
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	superAdminRole, err := repository.GetRoleByName(tx, "super_admin")
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("super_admin role not found, ensure system roles are initialized first")
		}
		return nil, fmt.Errorf("failed to get super_admin role: %w", err)
	}

	userRole := database.UserRole{
		UserID: adminUser.ID,
		RoleID: superAdminRole.ID,
	}
	if err := repository.CreateUserRole(tx, &userRole); err != nil {
		if errors.Is(err, consts.ErrAlreadyExists) {
			return nil, fmt.Errorf("admin user already has super_admin role")
		}
		return nil, fmt.Errorf("failed to assign super_admin role to admin user: %w", err)
	}

	for _, projectData := range data.Projects {
		project := projectData.ConvertToDBProject()
		if err := repository.CreateProject(tx, project); err != nil {
			if errors.Is(err, consts.ErrAlreadyExists) {
				return nil, fmt.Errorf("project %s already exists", project.Name)
			}
			return nil, fmt.Errorf("failed to create project %s: %w", project.Name, err)
		}

		if err := repository.CreateUserProject(tx, &database.UserProject{
			UserID:    adminUser.ID,
			ProjectID: project.ID,
			RoleID:    superAdminRole.ID,
			Status:    consts.CommonEnabled,
		}); err != nil {
			return nil, fmt.Errorf("failed to add admin user to project %s: %w", project.Name, err)
		}
	}

	return adminUser, nil
}

func initializeContainers(tx *gorm.DB, data *InitialData, userID int) error {
	dataPath := config.GetString("initialization.data_path")

	for _, containerData := range data.Containers {
		container := containerData.ConvertToDBContainer()
		if container.Type == consts.ContainerTypePedestal {
			system := chaos.SystemType(container.Name)
			if !system.IsValid() {
				return fmt.Errorf("invalid pedestal name: %s", container.Name)
			}
		}

		versions := make([]database.ContainerVersion, 0, len(containerData.Versions))
		for _, versionData := range containerData.Versions {
			version := versionData.ConvertToDBContainerVersion()

			if len(versionData.EnvVars) > 0 {
				params := make([]database.ParameterConfig, 0, len(versionData.EnvVars))
				for _, paramData := range versionData.EnvVars {
					param := paramData.ConvertToDBParameterConfig()
					params = append(params, *param)
				}
				version.EnvVars = params
			}

			if versionData.HelmConfig != nil {
				helmConfig := versionData.HelmConfig.ConvertToDBHelmConfig()
				if len(versionData.HelmConfig.Values) > 0 {
					params := make([]database.ParameterConfig, 0, len(versionData.HelmConfig.Values))
					for _, paramData := range versionData.HelmConfig.Values {
						param := paramData.ConvertToDBParameterConfig()
						params = append(params, *param)
					}
					helmConfig.DynamicValues = params
				}

				version.HelmConfig = helmConfig
			}

			versions = append(versions, *version)
		}

		container.Versions = versions

		createdContainer, err := producer.CreateContainerCore(tx, container, userID)
		if err != nil {
			return fmt.Errorf("failed to create container %s: %w", containerData.Name, err)
		}

		if createdContainer.Type == consts.ContainerTypePedestal {
			if err := producer.UploadHemlValueFileCore(
				tx,
				containerData.Name,
				container.Versions[0].HelmConfig,
				nil,
				filepath.Join(dataPath, fmt.Sprintf("%s.yaml", createdContainer.Name)),
			); err != nil {
				return fmt.Errorf("failed to upload helm value file for container %s: %w", containerData.Name, err)
			}
		}
	}

	return nil
}

func initializeDatasets(tx *gorm.DB, data *InitialData, userID int) error {
	for _, datasetData := range data.Datasets {
		dataset := datasetData.ConvertToDBDataset()

		versions := make([]database.DatasetVersion, 0, len(datasetData.Versions))
		for _, versionData := range datasetData.Versions {
			version := versionData.ConvertToDBDatasetVersion()
			versions = append(versions, *version)
		}

		_, err := producer.CreateDatasetCore(tx, dataset, versions, userID)
		if err != nil {
			return fmt.Errorf("failed to create dataset %s: %w", datasetData.Name, err)
		}
	}

	return nil
}

func initializeExecutionLabels(tx *gorm.DB) error {
	sourceLabels := []struct {
		value       string
		description string
	}{
		{consts.ExecutionSourceManual, consts.ExecutionManualDescription},
		{consts.ExecutionSourceSystem, consts.ExecutionSystemDescription},
	}

	for _, labelInfo := range sourceLabels {
		_, err := producer.CreateLabelCore(tx, &database.Label{
			Key:         consts.ExecutionLabelSource,
			Value:       labelInfo.value,
			Category:    consts.ExecutionCategory,
			Description: labelInfo.description,
		})
		if err != nil {
			return fmt.Errorf("failed to initialize execution label %s=%s: %w",
				consts.ExecutionLabelSource, labelInfo.value, err)
		}
	}

	return nil
}
