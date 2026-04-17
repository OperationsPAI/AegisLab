package metricmodule

import (
	"aegis/model"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListFaultInjections(query func(*gorm.DB) *gorm.DB) ([]model.FaultInjection, error) {
	var items []model.FaultInjection
	if err := query(r.db).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) ListExecutions(query func(*gorm.DB) *gorm.DB) ([]model.Execution, error) {
	var items []model.Execution
	if err := query(r.db).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) ListAlgorithmContainers() ([]model.Container, error) {
	var items []model.Container
	if err := r.db.Where("type = ?", 2).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
