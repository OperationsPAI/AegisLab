package trace

import (
	"aegis/consts"
	"aegis/model"
	"fmt"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetTraceByID(traceID string) (*model.Trace, error) {
	var trace model.Trace
	if err := r.db.Model(&model.Trace{}).
		Preload("Project").
		Preload("Tasks", func(db *gorm.DB) *gorm.DB {
			return db.Order("level ASC, sequence ASC")
		}).
		Where("id = ? AND status != ?", traceID, consts.CommonDeleted).
		First(&trace).Error; err != nil {
		return nil, err
	}
	return &trace, nil
}

func (r *Repository) ListTraces(limit, offset int, filterOptions *ListTraceFilters) ([]model.Trace, int64, error) {
	var (
		traces []model.Trace
		total  int64
	)

	query := r.db.Model(&model.Trace{}).Preload("Project")
	if filterOptions.TraceType != nil {
		query = query.Where("type = ?", *filterOptions.TraceType)
	}
	if filterOptions.GroupID != "" {
		query = query.Where("group_id = ?", filterOptions.GroupID)
	}
	if filterOptions.ProjectID > 0 {
		query = query.Where("project_id = ?", filterOptions.ProjectID)
	}
	if filterOptions.State != nil {
		query = query.Where("state = ?", *filterOptions.State)
	}
	if filterOptions.Status != nil {
		query = query.Where("status = ?", *filterOptions.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count traces: %w", err)
	}
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&traces).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list traces: %w", err)
	}
	return traces, total, nil
}
