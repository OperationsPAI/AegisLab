package initialization

import (
	"errors"
	"fmt"
	"path/filepath"

	"aegis/config"
	"aegis/consts"
	redisinfra "aegis/infra/redis"
	"aegis/model"
	containermodule "aegis/module/container"
	datasetmodule "aegis/module/dataset"
	labelmodule "aegis/module/label"
	"aegis/service/common"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
)

type permMeta struct {
	action        consts.ActionName
	resourceID    int
	resourceName  consts.ResourceName
	resourceScope consts.ResourceScope
}

func (r permMeta) String() string {
	return fmt.Sprintf("%v %v %v", r.action, r.resourceScope, r.resourceName)
}

func InitializeProducer(db *gorm.DB, publisher *redisinfra.Gateway, listener *common.ConfigUpdateListener) error {
	producerData, err := newConfigDataWithDB(db, consts.ConfigScopeProducer)
	if err != nil {
		return fmt.Errorf("failed to load producer config metadata: %w", err)
	}

	if len(producerData.configs) == 0 {
		logrus.Info("Seeding initial system data for producer...")
		if err := initializeProducer(db); err != nil {
			return fmt.Errorf("failed to initialize system data for producer: %w", err)
		}
		logrus.Info("Successfully seeded initial system data for producer")
	} else {
		logrus.Info("Initial system data for producer already seeded, skipping initialization")
	}

	// Initialize systems (seed builtins, register with chaos-experiment, set MetadataStore)
	if err := InitializeSystems(db); err != nil {
		return fmt.Errorf("failed to initialize systems: %w", err)
	}
	common.RegisterGlobalHandlers(publisher)
	if err := activateConfigScope(producerData.scope, listener); err != nil {
		return err
	}

	return nil
}

func initializeProducer(db *gorm.DB) error {
	dataPath := config.GetString("initialization.data_path")
	filePath := filepath.Join(dataPath, consts.InitialFilename)
	initialData, err := loadInitialDataFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load initial data from file: %w", err)
	}

	// System resources (following the order in system.go)
	resources := []model.Resource{
		{Name: consts.ResourceSystem, Type: consts.ResourceTypeSystem, Category: consts.ResourceCategorySystem},
		{Name: consts.ResourceAudit, Type: consts.ResourceTypeTable, Category: consts.ResourceCategorySystem},
		{Name: consts.ResourceConfiguration, Type: consts.ResourceTypeTable, Category: consts.ResourceCategorySystem},
		{Name: consts.ResourceContainer, Type: consts.ResourceTypeTable, Category: consts.ResourceCategoryAsset},
		{Name: consts.ResourceContainerVersion, Type: consts.ResourceTypeTable, Category: consts.ResourceCategoryAsset},
		{Name: consts.ResourceDataset, Type: consts.ResourceTypeTable, Category: consts.ResourceCategoryAsset},
		{Name: consts.ResourceDatasetVersion, Type: consts.ResourceTypeTable, Category: consts.ResourceCategoryAsset},
		{Name: consts.ResourceProject, Type: consts.ResourceTypeTable, Category: consts.ResourceCategoryPlatform},
		{Name: consts.ResourceTeam, Type: consts.ResourceTypeTable, Category: consts.ResourceCategoryPlatform},
		{Name: consts.ResourceLabel, Type: consts.ResourceTypeTable, Category: consts.ResourceCategoryAsset},
		{Name: consts.ResourceUser, Type: consts.ResourceTypeTable, Category: consts.ResourceCategorySystem},
		{Name: consts.ResourceRole, Type: consts.ResourceTypeTable, Category: consts.ResourceCategorySystem},
		{Name: consts.ResourcePermission, Type: consts.ResourceTypeTable, Category: consts.ResourceCategorySystem},
		{Name: consts.ResourceTask, Type: consts.ResourceTypeTable, Category: consts.ResourceCategoryChaos},
		{Name: consts.ResourceTrace, Type: consts.ResourceTypeTable, Category: consts.ResourceCategoryChaos},
		{Name: consts.ResourceInjection, Type: consts.ResourceTypeTable, Category: consts.ResourceCategoryChaos},
		{Name: consts.ResourceExecution, Type: consts.ResourceTypeTable, Category: consts.ResourceCategoryChaos},
	}

	for i := range resources {
		resources[i].DisplayName = consts.GetResourceDisplayName(resources[i].Name)
	}

	systemRoles := make([]model.Role, 0)
	for role, displayName := range consts.SystemRoleDisplayNames {
		systemRoles = append(systemRoles, model.Role{
			Name:        role.String(),
			DisplayName: displayName,
			IsSystem:    true,
			Status:      consts.CommonEnabled,
		})
	}

	return withOptimizedDBSettings(db, func() error {
		return db.Transaction(func(tx *gorm.DB) error {
			txStore := newBootstrapStore(tx)

			if err := txStore.upsertResources(resources); err != nil {
				return fmt.Errorf("failed to create system resources: %w", err)
			}

			resourceNames := make([]consts.ResourceName, 0, len(resources))
			for _, res := range resources {
				resourceNames = append(resourceNames, res.Name)
			}

			allResourcesInDB, err := txStore.listResourcesByNames(resourceNames)
			if err != nil {
				return fmt.Errorf("failed to get system resources from database: %w", err)
			}

			if len(allResourcesInDB) != len(resources) {
				return fmt.Errorf("mismatch in number of resources created and fetched")
			}

			resourceMap := make(map[consts.ResourceName]*model.Resource, len(allResourcesInDB))
			resourceIDMap := make(map[consts.ResourceName]int, len(allResourcesInDB))
			for _, res := range allResourcesInDB {
				resourceIDMap[res.Name] = res.ID
				resourceMap[res.Name] = &res
			}

			resourceMap[consts.ResourceContainerVersion].ParentID = utils.IntPtr(resourceIDMap[consts.ResourceContainer])
			resourceMap[consts.ResourceDatasetVersion].ParentID = utils.IntPtr(resourceIDMap[consts.ResourceDataset])

			toUpdatedResources := []model.Resource{
				*resourceMap[consts.ResourceContainerVersion],
				*resourceMap[consts.ResourceDatasetVersion],
			}

			if err := txStore.upsertResources(toUpdatedResources); err != nil {
				return fmt.Errorf("failed to update resource parent IDs: %w", err)
			}

			// Extract unique permissions from SystemRolePermissions to avoid creating unused permissions
			uniquePermissions := make(map[string]permMeta)

			for _, permissionRules := range consts.SystemRolePermissions {
				for _, rule := range permissionRules {
					resourceID, ok := resourceIDMap[rule.Resource]
					if !ok {
						return fmt.Errorf("resource %s not found in resourceIDMap", rule.Resource)
					}

					key := rule.String()
					if _, exists := uniquePermissions[key]; !exists {
						uniquePermissions[key] = permMeta{
							action:        rule.Action,
							resourceID:    resourceID,
							resourceName:  rule.Resource,
							resourceScope: rule.Scope,
						}
					}
				}
			}

			var permissionsToCreate []model.Permission
			for permName, permData := range uniquePermissions {
				resource, ok := resourceMap[permData.resourceName]
				if !ok {
					for _, res := range allResourcesInDB {
						if res.ID == permData.resourceID {
							resource = &res
							break
						}
					}
					if resource == nil {
						return fmt.Errorf("resource with ID %d not found", permData.resourceID)
					}
				}

				permission := model.Permission{
					Name:        permName,
					DisplayName: permData.String(),
					Action:      permData.action,
					Scope:       permData.resourceScope,
					ResourceID:  permData.resourceID,
					IsSystem:    true,
					Status:      consts.CommonEnabled,
				}
				permissionsToCreate = append(permissionsToCreate, permission)
			}

			if err := txStore.upsertPermissions(permissionsToCreate); err != nil {
				return fmt.Errorf("failed to create system permissions: %w", err)
			}

			if err := txStore.upsertRoles(systemRoles); err != nil {
				return fmt.Errorf("failed to create system roles: %w", err)
			}

			if err := assignSystemRolePermissions(txStore); err != nil {
				return fmt.Errorf("failed to assign system role permissions: %w", err)
			}

			adminUser, err := initializeAdminUser(txStore, initialData)
			if err != nil {
				return fmt.Errorf("failed to initialize admin user: %w", err)
			}

			if err := initializeProjectsAndTeams(txStore, initialData); err != nil {
				return fmt.Errorf("failed to initialize admin user, projects and teams: %w", err)
			}

			if err := initializeUsers(txStore, initialData); err != nil {
				return fmt.Errorf("failed to initialize users: %w", err)
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

func assignSystemRolePermissions(store *bootstrapStore) error {
	for roleName, permissionRules := range consts.SystemRolePermissions {
		role, err := store.getRoleByName(roleName.String())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("role %s not found", roleName)
			}
			return err
		}

		if roleName == consts.RoleSuperAdmin {
			permissions, err := store.listSystemPermissions()
			if err != nil {
				return fmt.Errorf("failed to list system permissions: %w", err)
			}

			var rolePermissions []model.RolePermission
			for _, perm := range permissions {
				rolePermissions = append(rolePermissions, model.RolePermission{
					RoleID:       role.ID,
					PermissionID: perm.ID,
				})
			}

			if err := store.createRolePermissions(rolePermissions); err != nil {
				return fmt.Errorf("failed to assign all permissions to super admin role: %w", err)
			}
		} else {
			var permissionStrs []string
			for _, rule := range permissionRules {
				permissionStrs = append(permissionStrs, rule.String())
			}

			permissions, err := store.listPermissionsByNames(permissionStrs)
			if err != nil {
				return fmt.Errorf("failed to list permissions for role %s: %w", roleName, err)
			}

			var rolePermissions []model.RolePermission
			for _, perm := range permissions {
				rolePermissions = append(rolePermissions, model.RolePermission{
					RoleID:       role.ID,
					PermissionID: perm.ID,
				})
			}

			if err := store.createRolePermissions(rolePermissions); err != nil {
				return fmt.Errorf("failed to assign permissions to role %s: %w", roleName, err)
			}
		}
	}

	return nil
}

func initializeAdminUser(store *bootstrapStore, data *InitialData) (*model.User, error) {
	adminUser := data.AdminUser.ConvertToDBUser()
	if err := store.createUser(adminUser); err != nil {
		if errors.Is(err, consts.ErrAlreadyExists) {
			return nil, fmt.Errorf("admin user already exists")
		}
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	superAdminRole, err := store.getRoleByName("super_admin")
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("super_admin role not found, ensure system roles are initialized first")
		}
		return nil, fmt.Errorf("failed to get super_admin role: %w", err)
	}

	userRole := model.UserRole{
		UserID: adminUser.ID,
		RoleID: superAdminRole.ID,
	}
	if err := store.createUserRole(&userRole); err != nil {
		if errors.Is(err, consts.ErrAlreadyExists) {
			return nil, fmt.Errorf("admin user already has super_admin role")
		}
		return nil, fmt.Errorf("failed to assign super_admin role to admin user: %w", err)
	}

	return adminUser, nil
}

func initializeProjectsAndTeams(store *bootstrapStore, data *InitialData) error {
	for _, teamData := range data.Teams {
		team := teamData.ConvertToDBTeam()
		if err := store.createTeam(team); err != nil {
			if errors.Is(err, consts.ErrAlreadyExists) {
				return fmt.Errorf("team %s already exists", team.Name)
			}
			return fmt.Errorf("failed to create team %s: %w", team.Name, err)
		}
	}

	for _, projectData := range data.Projects {
		project := projectData.ConvertToDBProject()
		if err := store.createProject(project); err != nil {
			if errors.Is(err, consts.ErrAlreadyExists) {
				return fmt.Errorf("project %s already exists", project.Name)
			}
			return fmt.Errorf("failed to create project %s: %w", project.Name, err)
		}
	}

	return nil
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

		versions := make([]model.ContainerVersion, 0, len(containerData.Versions))
		for _, versionData := range containerData.Versions {
			version := versionData.ConvertToDBContainerVersion()

			if len(versionData.EnvVars) > 0 {
				params := make([]model.ParameterConfig, 0, len(versionData.EnvVars))
				for _, paramData := range versionData.EnvVars {
					param := paramData.ConvertToDBParameterConfig()
					params = append(params, *param)
				}
				version.EnvVars = params
			}

			if versionData.HelmConfig != nil {
				helmConfig := versionData.HelmConfig.ConvertToDBHelmConfig()
				if len(versionData.HelmConfig.Values) > 0 {
					params := make([]model.ParameterConfig, 0, len(versionData.HelmConfig.Values))
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

		createdContainer, err := containermodule.CreateContainerCore(tx, container, userID)
		if err != nil {
			return fmt.Errorf("failed to create container %s: %w", containerData.Name, err)
		}

		if createdContainer.Type == consts.ContainerTypePedestal {
			if err := containermodule.UploadHelmValueFileFromPath(
				tx,
				containerData.Name,
				container.Versions[0].HelmConfig,
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

		versions := make([]model.DatasetVersion, 0, len(datasetData.Versions))
		for _, versionData := range datasetData.Versions {
			version := versionData.ConvertToDBDatasetVersion()
			versions = append(versions, *version)
		}

		_, err := datasetmodule.CreateDatasetCore(tx, dataset, versions, userID)
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
		_, err := labelmodule.CreateLabelCore(tx, &model.Label{
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

func initializeUsers(store *bootstrapStore, data *InitialData) error {
	if len(data.Users) == 0 {
		return nil
	}

	role, err := store.getRoleByName(consts.RoleUser.String())
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return fmt.Errorf("user role not found, ensure system roles are initialized first")
		}
		return fmt.Errorf("failed to get user role: %w", err)
	}

	for _, userData := range data.Users {
		user := userData.ConvertToDBUser()

		if err := store.createUser(user); err != nil {
			if errors.Is(err, consts.ErrAlreadyExists) {
				logrus.Warnf("User %s already exists, skipping", user.Username)
				continue
			}
			return fmt.Errorf("failed to create user %s: %w", user.Username, err)
		}

		if err := store.createUserRole(&model.UserRole{
			UserID: user.ID,
			RoleID: role.ID,
		}); err != nil {
			return fmt.Errorf("failed to assign default role to user %s: %w", user.Username, err)
		}

		// Bind user to specified teams with their roles
		if len(userData.Teams) > 0 {
			for _, teamBinding := range userData.Teams {
				// Get team by name
				team, err := store.getTeamByName(teamBinding.Name)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return fmt.Errorf("team %s not found for user %s", teamBinding.Name, user.Username)
					}
					return fmt.Errorf("failed to get team %s: %w", teamBinding.Name, err)
				}

				// Get role by name for user-team binding
				teamRole, err := store.getRoleByName(teamBinding.Role)
				if err != nil {
					if errors.Is(err, consts.ErrNotFound) {
						return fmt.Errorf("role %s not found for user %s in team %s", teamBinding.Role, user.Username, teamBinding.Name)
					}
					return fmt.Errorf("failed to get role %s: %w", teamBinding.Role, err)
				}

				// Bind user to team with role
				if err := store.createUserTeam(&model.UserTeam{
					UserID: user.ID,
					TeamID: team.ID,
					RoleID: teamRole.ID,
					Status: consts.CommonEnabled,
				}); err != nil {
					return fmt.Errorf("failed to bind user %s to team %s with role %s: %w", user.Username, teamBinding.Name, teamBinding.Role, err)
				}

				// Bind projects to this team and user if specified
				if len(teamBinding.Projects) > 0 {
					for _, projectBinding := range teamBinding.Projects {
						project, err := store.getProjectByName(projectBinding.Name)
						if err != nil {
							if errors.Is(err, gorm.ErrRecordNotFound) {
								return fmt.Errorf("project %s not found for team %s", projectBinding.Name, teamBinding.Name)
							}
							return fmt.Errorf("failed to get project %s: %w", projectBinding.Name, err)
						}

						// Update project's team_id to bind project to team
						project.TeamID = &team.ID
						if err := store.saveProject(project); err != nil {
							return fmt.Errorf("failed to bind project %s to team %s: %w", projectBinding.Name, teamBinding.Name, err)
						}

						logrus.Infof("Bound project %s to team %s", projectBinding.Name, teamBinding.Name)

						// Get role for user-project binding
						projectRole, err := store.getRoleByName(projectBinding.Role)
						if err != nil {
							if errors.Is(err, consts.ErrNotFound) {
								return fmt.Errorf("role %s not found for user %s in project %s", projectBinding.Role, user.Username, projectBinding.Name)
							}
							return fmt.Errorf("failed to get role %s: %w", projectBinding.Role, err)
						}

						// Bind user to project with role
						if err := store.createUserProject(&model.UserProject{
							UserID:    user.ID,
							ProjectID: project.ID,
							RoleID:    projectRole.ID,
							Status:    consts.CommonEnabled,
						}); err != nil {
							return fmt.Errorf("failed to bind user %s to project %s with role %s: %w", user.Username, projectBinding.Name, projectBinding.Role, err)
						}
					}
				}
			}
		} else {
			logrus.Infof("Created user %s without team bindings", user.Username)
		}

		// Bind user to specified projects directly (not through teams)
		if len(userData.Projects) > 0 {
			for _, projectBinding := range userData.Projects {
				// Get project by name
				project, err := store.getProjectByName(projectBinding.Name)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return fmt.Errorf("project %s not found for user %s", projectBinding.Name, user.Username)
					}
					return fmt.Errorf("failed to get project %s: %w", projectBinding.Name, err)
				}

				// Get role by name
				projectRole, err := store.getRoleByName(projectBinding.Role)
				if err != nil {
					if errors.Is(err, consts.ErrNotFound) {
						return fmt.Errorf("role %s not found for user %s in project %s", projectBinding.Role, user.Username, projectBinding.Name)
					}
					return fmt.Errorf("failed to get role %s: %w", projectBinding.Role, err)
				}

				// Bind user to project with role
				if err := store.createUserProject(&model.UserProject{
					UserID:    user.ID,
					ProjectID: project.ID,
					RoleID:    projectRole.ID,
					Status:    consts.CommonEnabled,
				}); err != nil {
					return fmt.Errorf("failed to bind user %s to project %s with role %s: %w", user.Username, projectBinding.Name, projectBinding.Role, err)
				}
			}
		}
	}

	return nil
}
