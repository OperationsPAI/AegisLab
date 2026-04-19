package pedestal

import (
	"fmt"

	"aegis/model"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetHelmConfigByContainerVersionID returns the helm_configs row bound to the
// given container version. Returns gorm.ErrRecordNotFound if absent.
func (r *Repository) GetHelmConfigByContainerVersionID(versionID int) (*model.HelmConfig, error) {
	var cfg model.HelmConfig
	if err := r.db.Where("container_version_id = ?", versionID).First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

// UpsertHelmConfig creates a new row if versionID has none, else updates
// the existing row in place. Returns the (fresh) row.
func (r *Repository) UpsertHelmConfig(versionID int, fields *model.HelmConfig) (*model.HelmConfig, error) {
	existing, err := r.GetHelmConfigByContainerVersionID(versionID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("query existing helm config: %w", err)
	}
	if existing != nil && existing.ID != 0 {
		existing.ChartName = fields.ChartName
		existing.Version = fields.Version
		existing.RepoURL = fields.RepoURL
		existing.RepoName = fields.RepoName
		existing.ValueFile = fields.ValueFile
		existing.LocalPath = fields.LocalPath
		if err := r.db.Save(existing).Error; err != nil {
			return nil, fmt.Errorf("update helm config: %w", err)
		}
		return existing, nil
	}
	fields.ContainerVersionID = versionID
	if err := r.db.Create(fields).Error; err != nil {
		return nil, fmt.Errorf("create helm config: %w", err)
	}
	return fields, nil
}
