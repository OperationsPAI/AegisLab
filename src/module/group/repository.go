package group

import (
	"aegis/consts"
	"aegis/model"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetTracesByGroupID(groupID string) ([]model.Trace, error) {
	var traces []model.Trace
	if err := r.db.Model(&model.Trace{}).
		Preload("Tasks").
		Where("group_id = ? AND status != ?", groupID, consts.CommonDeleted).
		Order("start_time DESC").
		Find(&traces).Error; err != nil {
		return nil, err
	}
	return traces, nil
}

func (r *Repository) CountTracesByGroupID(groupID string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Trace{}).
		Where("group_id = ? AND status != ?", groupID, consts.CommonDeleted).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
