package repository

import (
	"fmt"

	"aegis/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetSystemMetadata(db *gorm.DB, systemName, metadataType, serviceName string) (*database.SystemMetadata, error) {
	var meta database.SystemMetadata
	if err := db.
		Where("system_name = ? AND metadata_type = ? AND service_name = ?", systemName, metadataType, serviceName).
		First(&meta).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get system metadata: %w", err)
	}
	return &meta, nil
}

func ListSystemMetadata(db *gorm.DB, systemName, metadataType string) ([]database.SystemMetadata, error) {
	var metas []database.SystemMetadata
	query := db.Where("system_name = ?", systemName)
	if metadataType != "" {
		query = query.Where("metadata_type = ?", metadataType)
	}
	if err := query.Find(&metas).Error; err != nil {
		return nil, fmt.Errorf("failed to list system metadata: %w", err)
	}
	return metas, nil
}

func UpsertSystemMetadata(db *gorm.DB, meta *database.SystemMetadata) error {
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "system_name"}, {Name: "metadata_type"}, {Name: "service_name"}},
		DoUpdates: clause.AssignmentColumns([]string{"data", "updated_at"}),
	}).Create(meta).Error; err != nil {
		// Fallback: try find and update
		var existing database.SystemMetadata
		if findErr := db.Where("system_name = ? AND metadata_type = ? AND service_name = ?",
			meta.SystemName, meta.MetadataType, meta.ServiceName).First(&existing).Error; findErr == nil {
			return db.Model(&existing).Updates(map[string]interface{}{
				"data": meta.Data,
			}).Error
		}
		return fmt.Errorf("failed to upsert system metadata: %w", err)
	}
	return nil
}

func DeleteSystemMetadata(db *gorm.DB, systemName string) error {
	if err := db.Where("system_name = ?", systemName).Delete(&database.SystemMetadata{}).Error; err != nil {
		return fmt.Errorf("failed to delete system metadata for %s: %w", systemName, err)
	}
	return nil
}

func ListServiceNames(db *gorm.DB, systemName, metadataType string) ([]string, error) {
	var names []string
	query := db.Model(&database.SystemMetadata{}).
		Where("system_name = ?", systemName)
	if metadataType != "" {
		query = query.Where("metadata_type = ?", metadataType)
	}
	if err := query.Distinct("service_name").Pluck("service_name", &names).Error; err != nil {
		return nil, fmt.Errorf("failed to list service names: %w", err)
	}
	return names, nil
}
