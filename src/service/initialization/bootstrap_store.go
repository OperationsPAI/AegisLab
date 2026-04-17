package initialization

import (
	"errors"
	"fmt"

	"aegis/consts"
	"aegis/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	resourceOmitFields   = "Parent"
	roleOmitFields       = "ActiveName"
	permissionOmitFields = "ActiveName,Resource"
	userOmitFields       = "active_username"
	teamOmitFields       = "ActiveName"
	projectOmitFields    = "ActiveName"
	userTeamOmitFields   = "active_user_team"
)

type bootstrapStore struct {
	db *gorm.DB
}

func newBootstrapStore(db *gorm.DB) *bootstrapStore {
	return &bootstrapStore{db: db}
}

func (s *bootstrapStore) listExistingConfigs() ([]model.DynamicConfig, error) {
	var configs []model.DynamicConfig
	if err := s.db.Order("config_key ASC").Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to list all existing configs: %w", err)
	}
	return configs, nil
}

func (s *bootstrapStore) upsertResources(resources []model.Resource) error {
	if len(resources) == 0 {
		return fmt.Errorf("no resources to upsert")
	}
	if err := s.db.Omit(resourceOmitFields).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{}),
	}).Create(&resources).Error; err != nil {
		return fmt.Errorf("failed to batch upsert resources: %w", err)
	}
	return nil
}

func (s *bootstrapStore) listResourcesByNames(names []consts.ResourceName) ([]model.Resource, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("no resource names provided")
	}
	var resources []model.Resource
	if err := s.db.Where("name IN ?", names).Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("failed to list resources by names: %w", err)
	}
	return resources, nil
}

func (s *bootstrapStore) upsertPermissions(permissions []model.Permission) error {
	if len(permissions) == 0 {
		return fmt.Errorf("no permissions to upsert")
	}
	if err := s.db.Omit(permissionOmitFields).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{}),
	}).Create(&permissions).Error; err != nil {
		return fmt.Errorf("failed to batch upsert permissions: %w", err)
	}
	return nil
}

func (s *bootstrapStore) upsertRoles(roles []model.Role) error {
	if len(roles) == 0 {
		return fmt.Errorf("no roles to upsert")
	}
	if err := s.db.Omit(roleOmitFields).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{}),
	}).Create(&roles).Error; err != nil {
		return fmt.Errorf("failed to batch upsert roles: %w", err)
	}
	return nil
}

func (s *bootstrapStore) getRoleByName(name string) (*model.Role, error) {
	var role model.Role
	if err := s.db.Where("name = ? AND status != ?", name, consts.CommonDeleted).First(&role).Error; err != nil {
		return nil, fmt.Errorf("failed to find role with name %s: %w", name, err)
	}
	return &role, nil
}

func (s *bootstrapStore) listSystemPermissions() ([]model.Permission, error) {
	var permissions []model.Permission
	if err := s.db.Where("is_system = ? AND status = ?", true, consts.CommonEnabled).Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get system permissions: %w", err)
	}
	return permissions, nil
}

func (s *bootstrapStore) listPermissionsByNames(names []string) ([]model.Permission, error) {
	if len(names) == 0 {
		return []model.Permission{}, nil
	}
	var permissions []model.Permission
	if err := s.db.Where("name IN ? AND status = ?", names, consts.CommonEnabled).Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	return permissions, nil
}

func (s *bootstrapStore) createRolePermissions(rolePermissions []model.RolePermission) error {
	if len(rolePermissions) == 0 {
		return nil
	}
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&rolePermissions).Error; err != nil {
		return fmt.Errorf("failed to batch create role permissions: %w", err)
	}
	return nil
}

func (s *bootstrapStore) createUser(user *model.User) error {
	if err := s.db.Omit(userOmitFields).Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("%w: user %s already exists", consts.ErrAlreadyExists, user.Username)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (s *bootstrapStore) createUserRole(userRole *model.UserRole) error {
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(userRole).Error; err != nil {
		return fmt.Errorf("failed to create user-role association: %w", err)
	}
	return nil
}

func (s *bootstrapStore) createTeam(team *model.Team) error {
	if err := s.db.Omit(teamOmitFields).Create(team).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("%w: team %s already exists", consts.ErrAlreadyExists, team.Name)
		}
		return fmt.Errorf("failed to create team: %w", err)
	}
	return nil
}

func (s *bootstrapStore) createProject(project *model.Project) error {
	if err := s.db.Omit(projectOmitFields).Create(project).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("%w: project %s already exists", consts.ErrAlreadyExists, project.Name)
		}
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

func (s *bootstrapStore) getTeamByName(name string) (*model.Team, error) {
	var team model.Team
	if err := s.db.Where("name = ? AND status != ?", name, consts.CommonDeleted).First(&team).Error; err != nil {
		return nil, fmt.Errorf("failed to find team with name %s: %w", name, err)
	}
	return &team, nil
}

func (s *bootstrapStore) getProjectByName(name string) (*model.Project, error) {
	var project model.Project
	if err := s.db.Where("name = ? AND status != ?", name, consts.CommonDeleted).First(&project).Error; err != nil {
		return nil, fmt.Errorf("failed to find project with name %s: %w", name, err)
	}
	return &project, nil
}

func (s *bootstrapStore) saveProject(project *model.Project) error {
	if err := s.db.Omit(projectOmitFields).Save(project).Error; err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

func (s *bootstrapStore) createUserTeam(userTeam *model.UserTeam) error {
	if err := s.db.Omit(userTeamOmitFields).Clauses(clause.OnConflict{DoNothing: true}).Create(userTeam).Error; err != nil {
		return fmt.Errorf("failed to create user-team association: %w", err)
	}
	return nil
}

func (s *bootstrapStore) createUserProject(userProject *model.UserProject) error {
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(userProject).Error; err != nil {
		return fmt.Errorf("failed to create user-project association: %w", err)
	}
	return nil
}

func (s *bootstrapStore) listEnabledSystems() ([]model.System, error) {
	var systems []model.System
	if err := s.db.Where("status = ?", consts.CommonEnabled).Find(&systems).Error; err != nil {
		return nil, fmt.Errorf("failed to list enabled systems: %w", err)
	}
	return systems, nil
}
