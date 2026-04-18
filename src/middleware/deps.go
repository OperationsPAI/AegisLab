package middleware

import (
	"context"
	"errors"
	"fmt"
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	"aegis/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TokenVerifier interface {
	VerifyToken(ctx context.Context, token string) (*utils.Claims, error)
	VerifyServiceToken(ctx context.Context, token string) (*utils.ServiceClaims, error)
}

type permissionChecker interface {
	CheckUserPermission(context.Context, *dto.CheckPermissionParams) (bool, error)
	IsUserTeamAdmin(context.Context, int, int) (bool, error)
	IsUserInTeam(context.Context, int, int) (bool, error)
	IsTeamPublic(context.Context, int) (bool, error)
	IsUserProjectAdmin(context.Context, int, int) (bool, error)
	IsUserInProject(context.Context, int, int) (bool, error)
}

type auditLogger interface {
	LogFailedAction(ipAddress, userAgent, action, errorMsg string, duration, userID int, resourceName consts.ResourceName) error
	LogUserAction(ipAddress, userAgent, action, details string, duration, userID int, resourceName consts.ResourceName) error
}

type Service interface {
	TokenVerifier
	permissionChecker
	auditLogger
}

const middlewareServiceContextKey = "middleware.service"

func NewService(db *gorm.DB, verifier TokenVerifier) Service {
	return &dbBackedMiddlewareService{db: db, verifier: verifier}
}

func InjectService(service Service) gin.HandlerFunc {
	if service == nil {
		service = noopMiddlewareService{}
	}

	return func(c *gin.Context) {
		c.Set(middlewareServiceContextKey, service)
		c.Next()
	}
}

func permissionCheckerFromContext(c *gin.Context) permissionChecker { return serviceFromContext(c) }

func auditLoggerFromContext(c *gin.Context) auditLogger { return serviceFromContext(c) }

func serviceFromContext(c *gin.Context) Service {
	if c == nil {
		return noopMiddlewareService{}
	}
	service, ok := c.Get(middlewareServiceContextKey)
	if !ok {
		return noopMiddlewareService{}
	}
	middlewareService, ok := service.(Service)
	if !ok || middlewareService == nil {
		return noopMiddlewareService{}
	}
	return middlewareService
}

type dbBackedMiddlewareService struct {
	db       *gorm.DB
	verifier TokenVerifier
}

func (s *dbBackedMiddlewareService) VerifyToken(ctx context.Context, token string) (*utils.Claims, error) {
	if s.verifier == nil {
		return nil, fmt.Errorf("token verifier not initialized")
	}
	return s.verifier.VerifyToken(ctx, token)
}

func (s *dbBackedMiddlewareService) VerifyServiceToken(ctx context.Context, token string) (*utils.ServiceClaims, error) {
	if s.verifier == nil {
		return nil, fmt.Errorf("token verifier not initialized")
	}
	return s.verifier.VerifyServiceToken(ctx, token)
}

func (s *dbBackedMiddlewareService) CheckUserPermission(_ context.Context, params *dto.CheckPermissionParams) (bool, error) {
	if err := params.Validate(); err != nil {
		return false, fmt.Errorf("invalid request: %w", err)
	}

	permission, err := s.getPermissionByActionAndResource(params.Action, params.Scope, params.ResourceName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to find target permission: %w", err)
	}

	return s.checkUserHasPermission(params, permission.ID)
}

func (s *dbBackedMiddlewareService) IsUserInTeam(_ context.Context, userID, teamID int) (bool, error) {
	ut, err := s.getUserTeamRole(userID, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return ut != nil, nil
}

func (s *dbBackedMiddlewareService) IsUserTeamAdmin(_ context.Context, userID, teamID int) (bool, error) {
	ut, err := s.getUserTeamRole(userID, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return ut != nil && ut.Role != nil && ut.Role.Name == consts.RoleTeamAdmin.String(), nil
}

func (s *dbBackedMiddlewareService) IsTeamPublic(_ context.Context, teamID int) (bool, error) {
	team, err := s.getTeamByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return team.IsPublic, nil
}

func (s *dbBackedMiddlewareService) IsUserInProject(_ context.Context, userID, projectID int) (bool, error) {
	up, err := s.getUserProjectRole(userID, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return up != nil, nil
}

func (s *dbBackedMiddlewareService) IsUserProjectAdmin(_ context.Context, userID, projectID int) (bool, error) {
	up, err := s.getUserProjectRole(userID, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return up != nil && up.Role != nil && up.Role.Name == consts.RoleProjectAdmin.String(), nil
}

func (s *dbBackedMiddlewareService) LogFailedAction(ipAddress, userAgent, action, errorMsg string, duration, userID int, resourceName consts.ResourceName) error {
	if resourceName == "" {
		return fmt.Errorf("resource name cannot be empty")
	}

	log := &model.AuditLog{
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Duration:  duration,
		Action:    action,
		ErrorMsg:  errorMsg,
		UserID:    userID,
		State:     consts.AuditLogStateFailed,
		Status:    consts.CommonEnabled,
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		resource, err := s.getResourceByName(tx, resourceName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: resource %s not found", consts.ErrNotFound, resourceName)
			}
			return fmt.Errorf("failed to get resource: %w", err)
		}

		log.ResourceID = resource.ID
		return s.createAuditLog(tx, log)
	})
}

func (s *dbBackedMiddlewareService) LogUserAction(ipAddress, userAgent, action, details string, duration, userID int, resourceName consts.ResourceName) error {
	if resourceName == "" {
		return fmt.Errorf("resource name cannot be empty")
	}

	log := &model.AuditLog{
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Duration:  duration,
		Action:    action,
		Details:   details,
		UserID:    userID,
		State:     consts.AuditLogStateSuccess,
		Status:    consts.CommonEnabled,
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		resource, err := s.getResourceByName(tx, resourceName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: resource %s not found", consts.ErrNotFound, resourceName)
			}
			return fmt.Errorf("failed to get resource: %w", err)
		}

		log.ResourceID = resource.ID
		return s.createAuditLog(tx, log)
	})
}

func (s *dbBackedMiddlewareService) getPermissionByActionAndResource(action consts.ActionName, scope consts.ResourceScope, resourceName consts.ResourceName) (*model.Permission, error) {
	var permission model.Permission
	if err := s.db.
		Select("permissions.*").
		Joins("JOIN resources ON permissions.resource_id = resources.id").
		Where("permissions.action = ? AND permissions.scope = ? AND resources.name = ?", action, scope, resourceName).
		Where("permissions.status != ?", consts.CommonDeleted).
		First(&permission).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

func (s *dbBackedMiddlewareService) checkUserHasPermission(params *dto.CheckPermissionParams, permissionID int) (bool, error) {
	directQuery := s.buildDirectPermissionQuery(params.UserID, permissionID, params.ProjectID, params.ContainerID, params.DatasetID)
	globalRoleQuery := s.buildGlobalRolePermissionQuery(params.UserID, permissionID)
	finalQuery := s.db.Table("(? UNION ALL ?) as base", directQuery, globalRoleQuery)

	if params.TeamID != nil {
		finalQuery = s.db.Table("(? UNION ALL ?) as combined", finalQuery, s.buildTeamRolePermissionQuery(params.UserID, permissionID, *params.TeamID))
	}
	if params.ProjectID != nil {
		finalQuery = s.db.Table("(? UNION ALL ?) as combined", finalQuery, s.buildProjectRolePermissionQuery(params.UserID, permissionID, *params.ProjectID))
	}
	if params.ContainerID != nil {
		finalQuery = s.db.Table("(? UNION ALL ?) as combined", finalQuery, s.buildContainerRolePermissionQuery(params.UserID, permissionID, *params.ContainerID))
	}
	if params.DatasetID != nil {
		finalQuery = s.db.Table("(? UNION ALL ?) as combined", finalQuery, s.buildDatasetRolePermissionQuery(params.UserID, permissionID, *params.DatasetID))
	}

	var count int64
	if err := finalQuery.Limit(1).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user permission: %w", err)
	}
	return count > 0, nil
}

func (s *dbBackedMiddlewareService) buildDirectPermissionQuery(userID int, permissionID int, projectID, containerID, datasetID *int) *gorm.DB {
	query := s.db.
		Select("up.permission_id").
		Table("user_permissions up").
		Where("up.user_id = ? AND up.permission_id = ?", userID, permissionID).
		Where("up.grant_type = ?", consts.GrantTypeGrant).
		Where("up.expires_at IS NULL OR up.expires_at > ?", time.Now())

	if projectID != nil {
		query = query.Where("up.project_id IS NULL OR up.project_id = ?", *projectID)
	} else {
		query = query.Where("up.project_id IS NULL")
	}
	if containerID != nil {
		query = query.Where("up.container_id IS NULL OR up.container_id = ?", *containerID)
	} else {
		query = query.Where("up.container_id IS NULL")
	}
	if datasetID != nil {
		query = query.Where("up.dataset_id IS NULL OR up.dataset_id = ?", *datasetID)
	} else {
		query = query.Where("up.dataset_id IS NULL")
	}

	return query
}

func (s *dbBackedMiddlewareService) buildGlobalRolePermissionQuery(userID int, permissionID int) *gorm.DB {
	return s.db.
		Select("rp.permission_id").
		Table("role_permissions rp").
		Joins("JOIN user_roles ur ON rp.role_id = ur.role_id").
		Where("ur.user_id = ? AND rp.permission_id = ?", userID, permissionID)
}

func (s *dbBackedMiddlewareService) buildTeamRolePermissionQuery(userID int, permissionID int, teamID int) *gorm.DB {
	return s.db.
		Select("rp.permission_id").
		Table("role_permissions rp").
		Joins("JOIN user_teams ut ON rp.role_id = ut.role_id").
		Where("ut.user_id = ? AND ut.team_id = ? AND rp.permission_id = ?", userID, teamID, permissionID).
		Where("ut.status = ?", consts.CommonEnabled)
}

func (s *dbBackedMiddlewareService) buildProjectRolePermissionQuery(userID int, permissionID int, projectID int) *gorm.DB {
	return s.db.
		Select("rp.permission_id").
		Table("role_permissions rp").
		Joins("JOIN user_projects upr ON rp.role_id = upr.role_id").
		Where("upr.user_id = ? AND upr.project_id = ? AND rp.permission_id = ?", userID, projectID, permissionID).
		Where("upr.status = ?", consts.CommonEnabled)
}

func (s *dbBackedMiddlewareService) buildContainerRolePermissionQuery(userID int, permissionID int, containerID int) *gorm.DB {
	return s.db.
		Select("rp.permission_id").
		Table("role_permissions rp").
		Joins("JOIN user_containers uc ON rp.role_id = uc.role_id").
		Where("uc.user_id = ? AND uc.container_id = ? AND rp.permission_id = ?", userID, containerID, permissionID).
		Where("uc.status = ?", consts.CommonEnabled)
}

func (s *dbBackedMiddlewareService) buildDatasetRolePermissionQuery(userID int, permissionID int, datasetID int) *gorm.DB {
	return s.db.
		Select("rp.permission_id").
		Table("role_permissions rp").
		Joins("JOIN user_datasets ud ON rp.role_id = ud.role_id").
		Where("ud.user_id = ? AND ud.dataset_id = ? AND rp.permission_id = ?", userID, datasetID, permissionID).
		Where("ud.status = ?", consts.CommonEnabled)
}

func (s *dbBackedMiddlewareService) getUserTeamRole(userID, teamID int) (*model.UserTeam, error) {
	var userTeam model.UserTeam
	if err := s.db.Preload("Role").
		Where("user_id = ? AND team_id = ? AND status = ?", userID, teamID, consts.CommonEnabled).
		First(&userTeam).Error; err != nil {
		return nil, err
	}
	return &userTeam, nil
}

func (s *dbBackedMiddlewareService) getTeamByID(teamID int) (*model.Team, error) {
	var team model.Team
	if err := s.db.Where("id = ?", teamID).First(&team).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

func (s *dbBackedMiddlewareService) getUserProjectRole(userID, projectID int) (*model.UserProject, error) {
	var userProject model.UserProject
	if err := s.db.Preload("Role").
		Where("user_id = ? AND project_id = ? AND status = ?", userID, projectID, consts.CommonEnabled).
		First(&userProject).Error; err != nil {
		return nil, err
	}
	return &userProject, nil
}

func (s *dbBackedMiddlewareService) getResourceByName(db *gorm.DB, resourceName consts.ResourceName) (*model.Resource, error) {
	var resource model.Resource
	if err := db.Where("name = ? AND status != ?", resourceName, consts.CommonDeleted).First(&resource).Error; err != nil {
		return nil, err
	}
	return &resource, nil
}

func (s *dbBackedMiddlewareService) createAuditLog(db *gorm.DB, log *model.AuditLog) error {
	return db.Create(log).Error
}

type noopMiddlewareService struct{}

func (noopMiddlewareService) VerifyToken(context.Context, string) (*utils.Claims, error) {
	return nil, fmt.Errorf("token verifier not initialized")
}
func (noopMiddlewareService) VerifyServiceToken(context.Context, string) (*utils.ServiceClaims, error) {
	return nil, fmt.Errorf("token verifier not initialized")
}

func (noopMiddlewareService) CheckUserPermission(context.Context, *dto.CheckPermissionParams) (bool, error) {
	return false, fmt.Errorf("permission checker not initialized")
}
func (noopMiddlewareService) IsUserTeamAdmin(context.Context, int, int) (bool, error) {
	return false, fmt.Errorf("permission checker not initialized")
}
func (noopMiddlewareService) IsUserInTeam(context.Context, int, int) (bool, error) {
	return false, fmt.Errorf("permission checker not initialized")
}
func (noopMiddlewareService) IsTeamPublic(context.Context, int) (bool, error) {
	return false, fmt.Errorf("permission checker not initialized")
}
func (noopMiddlewareService) IsUserProjectAdmin(context.Context, int, int) (bool, error) {
	return false, fmt.Errorf("permission checker not initialized")
}
func (noopMiddlewareService) IsUserInProject(context.Context, int, int) (bool, error) {
	return false, fmt.Errorf("permission checker not initialized")
}

func (noopMiddlewareService) LogFailedAction(string, string, string, string, int, int, consts.ResourceName) error {
	return fmt.Errorf("audit logger not initialized")
}
func (noopMiddlewareService) LogUserAction(string, string, string, string, int, int, consts.ResourceName) error {
	return fmt.Errorf("audit logger not initialized")
}
