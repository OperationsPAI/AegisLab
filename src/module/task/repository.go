package task

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

func (r *Repository) BatchDelete(taskIDs []string) error {
	if len(taskIDs) == 0 {
		return nil
	}

	if err := r.db.Model(&model.Task{}).
		Where("id IN (?) AND status != ?", taskIDs, consts.CommonDeleted).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return fmt.Errorf("failed to batch delete tasks: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(taskID string) (*model.Task, error) {
	var task model.Task
	if err := r.db.
		Preload("FaultInjection.Benchmark.Container").
		Preload("FaultInjection.Pedestal.Container").
		Preload("Execution.AlgorithmVersion.Container").
		Preload("Execution.Datapack").
		Preload("Execution.DatasetVersion").
		Where("id = ? AND status != ?", taskID, consts.CommonDeleted).
		First(&task).Error; err != nil {
		return nil, fmt.Errorf("failed to find task with id %s: %w", taskID, err)
	}
	return &task, nil
}

// UpdateExecuteTime updates the execute_time column of a task row.
func (r *Repository) UpdateExecuteTime(taskID string, executeTime int64) error {
	return r.db.Model(&model.Task{}).
		Where("id = ?", taskID).
		Update("execute_time", executeTime).Error
}

func (r *Repository) List(limit, offset int, filters *ListTaskFilters) ([]model.Task, int64, error) {
	var (
		tasks []model.Task
		total int64
	)

	query := r.db.Model(&model.Task{})
	if filters.Immediate != nil {
		query = query.Where("immediate = ?", *filters.Immediate)
	}
	if filters.TaskType != nil {
		query = query.Where("type = ?", *filters.TaskType)
	}
	if filters.TraceID != "" {
		query = query.Where("trace_id = ?", filters.TraceID)
	}
	if filters.GroupID != "" {
		query = query.Where("group_id = ?", filters.GroupID)
	}
	if filters.ProjectID > 0 {
		query = query.Where("project_id = ?", filters.ProjectID)
	}
	if filters.State != nil {
		query = query.Where("state = ?", *filters.State)
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}
	return tasks, total, nil
}
