package containermodule

import (
	"aegis/consts"
	"aegis/model"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	containerCommonOmitFields       = "active_name"
	containerModelOmitFields        = "Versions"
	containerVersionModelOmitFields = "active_version_key,HelmConfig,EnvVars"
	helmConfigModelOmitFields       = "Values"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *Repository) withDB(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetRoleByName(name string) (*model.Role, error) {
	var role model.Role
	if err := r.db.Where("name = ? and status != ?", name, consts.CommonDeleted).First(&role).Error; err != nil {
		return nil, fmt.Errorf("failed to find role with name %s: %w", name, err)
	}
	return &role, nil
}

func (r *Repository) CreateContainer(container *model.Container) error {
	if err := r.db.Omit(containerCommonOmitFields, containerModelOmitFields).Create(container).Error; err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}
	return nil
}

func (r *Repository) CreateUserContainer(userContainer *model.UserContainer) error {
	if err := r.db.Omit("active_user_container").Create(userContainer).Error; err != nil {
		return fmt.Errorf("failed to create user-container association: %w", err)
	}
	return nil
}

func (r *Repository) BatchDeleteContainerVersions(containerID int) (int64, error) {
	result := r.db.Model(&model.ContainerVersion{}).
		Where("container_id = ? AND status != ?", containerID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to batch soft delete container versions for container %d: %w", containerID, result.Error)
	}
	return result.RowsAffected, nil
}

func (r *Repository) RemoveUsersFromContainer(containerID int) (int64, error) {
	result := r.db.Model(&model.UserContainer{}).
		Where("container_id = ? AND status != ?", containerID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to delete user-container associations for container %d: %w", containerID, err)
	}
	return result.RowsAffected, nil
}

func (r *Repository) ClearContainerLabels(containerIDs []int, labelIDs []int) error {
	if len(containerIDs) == 0 {
		return nil
	}

	query := r.db.Table("container_labels").Where("container_id IN (?)", containerIDs)
	if len(labelIDs) > 0 {
		query = query.Where("label_id IN (?)", labelIDs)
	}
	if err := query.Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to clear container-label associations: %w", err)
	}
	return nil
}

func (r *Repository) DeleteContainer(containerID int) (int64, error) {
	result := r.db.Model(&model.Container{}).
		Where("id = ? AND status != ?", containerID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to delete container %d: %w", containerID, err)
	}
	return result.RowsAffected, nil
}

func (r *Repository) GetContainerByID(containerID int) (*model.Container, error) {
	var container model.Container
	if err := r.db.Where("id = ? AND status != ?", containerID, consts.CommonDeleted).First(&container).Error; err != nil {
		return nil, fmt.Errorf("failed to find container with id %d: %w", containerID, err)
	}
	return &container, nil
}

func (r *Repository) ListContainerVersionsByContainerID(containerID int) ([]model.ContainerVersion, error) {
	var versions []model.ContainerVersion
	if err := r.db.
		Preload("Container").
		Preload("HelmConfig").
		Where("container_id = ?", containerID).
		Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to list container versions for container %d: %w", containerID, err)
	}
	return versions, nil
}

func (r *Repository) ListContainers(limit, offset int, containerType *consts.ContainerType, isPublic *bool, status *consts.StatusType) ([]model.Container, int64, error) {
	var (
		containers []model.Container
		total      int64
	)

	query := r.db.Model(&model.Container{})
	if containerType != nil {
		query = query.Where("type = ?", *containerType)
	}
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count containers: %w", err)
	}
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&containers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list containers: %w", err)
	}
	return containers, total, nil
}

func (r *Repository) ListContainerLabels(containerIDs []int) (map[int][]model.Label, error) {
	if len(containerIDs) == 0 {
		return nil, nil
	}

	type containerLabelResult struct {
		model.Label
		ContainerID int `gorm:"column:container_id"`
	}

	var flatResults []containerLabelResult
	if err := r.db.Model(&model.Label{}).
		Joins("JOIN container_labels cl ON cl.label_id = labels.id").
		Where("cl.container_id IN (?)", containerIDs).
		Select("labels.*, cl.container_id").
		Find(&flatResults).Error; err != nil {
		return nil, fmt.Errorf("failed to batch query container labels: %w", err)
	}

	labelsMap := make(map[int][]model.Label, len(containerIDs))
	for _, id := range containerIDs {
		labelsMap[id] = []model.Label{}
	}
	for _, res := range flatResults {
		labelsMap[res.ContainerID] = append(labelsMap[res.ContainerID], res.Label)
	}
	return labelsMap, nil
}

func (r *Repository) UpdateContainer(container *model.Container) error {
	if err := r.db.Omit(containerCommonOmitFields).Save(container).Error; err != nil {
		return fmt.Errorf("failed to update container: %w", err)
	}
	return nil
}

func (r *Repository) AddContainerLabels(containerLabels []model.ContainerLabel) error {
	if len(containerLabels) == 0 {
		return nil
	}
	if err := r.db.Create(&containerLabels).Error; err != nil {
		return fmt.Errorf("failed to add container-label associations: %w", err)
	}
	return nil
}

func (r *Repository) ListLabelIDsByKeyAndContainerID(containerID int, keys []string) ([]int, error) {
	var labelIDs []int
	if err := r.db.Table("labels l").
		Select("l.id").
		Joins("JOIN container_labels cl ON cl.label_id = l.id").
		Where("cl.container_id = ? AND l.label_key IN (?)", containerID, keys).
		Pluck("l.id", &labelIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to find label IDs by keys for container %d: %w", containerID, err)
	}
	return labelIDs, nil
}

func (r *Repository) BatchDecreaseLabelUsages(labelIDs []int, decrement int) error {
	if len(labelIDs) == 0 {
		return nil
	}

	expr := gorm.Expr("GREATEST(0, usage_count - ?)", decrement)
	if err := r.db.Model(&model.Label{}).
		Where("id IN (?)", labelIDs).
		Clauses(clause.Returning{}).
		UpdateColumn("usage_count", expr).Error; err != nil {
		return fmt.Errorf("failed to batch decrease label usages: %w", err)
	}
	return nil
}

func (r *Repository) ListLabelsByContainerID(containerID int) ([]model.Label, error) {
	var labels []model.Label
	if err := r.db.Model(&model.Label{}).
		Joins("JOIN container_labels cl ON cl.label_id = labels.id").
		Where("cl.container_id = ?", containerID).
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to list labels for container %d: %w", containerID, err)
	}
	return labels, nil
}

func (r *Repository) BatchCreateContainerVersions(versions []model.ContainerVersion) error {
	if len(versions) == 0 {
		return fmt.Errorf("no container versions to create")
	}
	if err := r.db.Omit(containerVersionModelOmitFields).Create(&versions).Error; err != nil {
		return fmt.Errorf("failed to batch create container versions: %w", err)
	}
	return nil
}

func (r *Repository) BatchCreateOrFindParameterConfigs(params []model.ParameterConfig) error {
	if len(params) == 0 {
		return nil
	}
	if err := r.db.Clauses(clause.OnConflict{OnConstraint: "idx_unique_config", DoNothing: true}).Create(&params).Error; err != nil {
		return fmt.Errorf("failed to batch create parameter configs: %w", err)
	}
	return nil
}

func (r *Repository) ListParameterConfigsByKeys(configs []model.ParameterConfig) ([]model.ParameterConfig, error) {
	if len(configs) == 0 {
		return []model.ParameterConfig{}, nil
	}

	var results []model.ParameterConfig
	query := r.db.Model(&model.ParameterConfig{})
	conditions := r.db.Where("1 = 0")
	for _, cfg := range configs {
		conditions = conditions.Or(r.db.Where("config_key = ? AND type = ? AND category = ?", cfg.Key, cfg.Type, cfg.Category))
	}
	if err := query.Where(conditions).Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to list parameter configs by keys: %w", err)
	}
	return results, nil
}

func (r *Repository) AddContainerVersionEnvVars(envVars []model.ContainerVersionEnvVar) error {
	if len(envVars) == 0 {
		return nil
	}
	if err := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&envVars).Error; err != nil {
		return fmt.Errorf("failed to add container version env vars: %w", err)
	}
	return nil
}

func (r *Repository) BatchCreateHelmConfigs(helmConfigs []*model.HelmConfig) error {
	if len(helmConfigs) == 0 {
		return fmt.Errorf("no helm configs to create")
	}
	if err := r.db.Omit(helmConfigModelOmitFields).Create(helmConfigs).Error; err != nil {
		return fmt.Errorf("failed to batch create helm configs: %v", err)
	}
	return nil
}

func (r *Repository) AddHelmConfigValues(helmValues []model.HelmConfigValue) error {
	if len(helmValues) == 0 {
		return nil
	}
	if err := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&helmValues).Error; err != nil {
		return fmt.Errorf("failed to add helm config values: %w", err)
	}
	return nil
}

func (r *Repository) DeleteContainerVersion(versionID int) (int64, error) {
	result := r.db.Model(&model.ContainerVersion{}).
		Where("id = ? AND status != ?", versionID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to soft delete container version %d: %w", versionID, result.Error)
	}
	return result.RowsAffected, nil
}

func (r *Repository) GetContainerVersionByID(versionID int) (*model.ContainerVersion, error) {
	var version model.ContainerVersion
	if err := r.db.
		Preload("Container").
		Preload("HelmConfig").
		Where("id = ?", versionID).
		First(&version).Error; err != nil {
		return nil, fmt.Errorf("failed to find container version with id %d: %w", versionID, err)
	}
	return &version, nil
}

func (r *Repository) ListContainerVersions(limit, offset int, containerID int, status *consts.StatusType) ([]model.ContainerVersion, int64, error) {
	var (
		versions []model.ContainerVersion
		total    int64
	)

	query := r.db.Model(&model.ContainerVersion{}).Where("container_id = ?", containerID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count container versions: %v", err)
	}
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&versions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list container versions: %v", err)
	}
	return versions, total, nil
}

func (r *Repository) UpdateContainerVersion(version *model.ContainerVersion) error {
	if err := r.db.Omit(containerVersionModelOmitFields).Save(version).Error; err != nil {
		return fmt.Errorf("failed to update container version: %w", err)
	}
	return nil
}

func (r *Repository) GetHelmConfigByContainerVersionID(versionID int) (*model.HelmConfig, error) {
	var helmConfig model.HelmConfig
	if err := r.db.Preload("ContainerVersion").Where("container_version_id = ?", versionID).First(&helmConfig).Error; err != nil {
		return nil, fmt.Errorf("failed to find helm config for version id %d: %w", versionID, err)
	}
	return &helmConfig, nil
}

func (r *Repository) UpdateHelmConfig(helmConfig *model.HelmConfig) error {
	if err := r.db.Save(helmConfig).Error; err != nil {
		return fmt.Errorf("failed to update helm config: %w", err)
	}
	return nil
}
