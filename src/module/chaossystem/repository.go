package chaossystemmodule

import (
	"aegis/consts"
	"aegis/model"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListSystems(limit, offset int) ([]model.System, int64, error) {
	var (
		systems []model.System
		total   int64
	)

	query := r.db.Model(&model.System{}).Where("status != ?", consts.CommonDeleted)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count systems: %w", err)
	}
	if err := query.Limit(limit).Offset(offset).Order("updated_at DESC").Find(&systems).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list systems: %w", err)
	}
	return systems, total, nil
}

func (r *Repository) GetSystemByID(id int) (*model.System, error) {
	var system model.System
	if err := r.db.Where("id = ? AND status != ?", id, consts.CommonDeleted).First(&system).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("system with id %d: %w", id, consts.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to find system with id %d: %w", id, err)
	}
	return &system, nil
}

func (r *Repository) CreateSystem(system *model.System) error {
	if err := r.db.Create(system).Error; err != nil {
		return fmt.Errorf("failed to create system: %w", err)
	}
	return nil
}

func (r *Repository) UpdateSystem(id int, updates map[string]interface{}) error {
	result := r.db.Model(&model.System{}).
		Where("id = ? AND status != ?", id, consts.CommonDeleted).
		Updates(updates)
	if err := result.Error; err != nil {
		return fmt.Errorf("failed to update system with id %d: %w", id, err)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("system with id %d: %w", id, consts.ErrNotFound)
	}
	return nil
}

func (r *Repository) DeleteSystem(id int) error {
	result := r.db.Model(&model.System{}).
		Where("id = ? AND status != ?", id, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if err := result.Error; err != nil {
		return fmt.Errorf("failed to delete system with id %d: %w", id, err)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("system with id %d: %w", id, consts.ErrNotFound)
	}
	return nil
}

func (r *Repository) UpsertSystemMetadata(meta *model.SystemMetadata) error {
	if err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "system_name"}, {Name: "metadata_type"}, {Name: "service_name"}},
		DoUpdates: clause.AssignmentColumns([]string{"data", "updated_at"}),
	}).Create(meta).Error; err != nil {
		var existing model.SystemMetadata
		if findErr := r.db.Where("system_name = ? AND metadata_type = ? AND service_name = ?",
			meta.SystemName, meta.MetadataType, meta.ServiceName).First(&existing).Error; findErr == nil {
			return r.db.Model(&existing).Updates(map[string]any{"data": meta.Data}).Error
		}
		return fmt.Errorf("failed to upsert system metadata: %w", err)
	}
	return nil
}

func (r *Repository) ListSystemMetadata(systemName, metadataType string) ([]model.SystemMetadata, error) {
	var metas []model.SystemMetadata
	query := r.db.Where("system_name = ?", systemName)
	if metadataType != "" {
		query = query.Where("metadata_type = ?", metadataType)
	}
	if err := query.Find(&metas).Error; err != nil {
		return nil, fmt.Errorf("failed to list system metadata: %w", err)
	}
	return metas, nil
}
