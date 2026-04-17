package evaluationmodule

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

func (r *Repository) ListEvaluations(limit, offset int) ([]model.Evaluation, int64, error) {
	var (
		evaluations []model.Evaluation
		total       int64
	)

	query := r.db.Model(&model.Evaluation{}).
		Where("status != ?", consts.CommonDeleted)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count evaluations: %w", err)
	}
	if err := query.Select("id, project_id, algorithm_name, algorithm_version, datapack_name, dataset_name, dataset_version, eval_type, precision, recall, f1_score, accuracy, status, created_at, updated_at").
		Limit(limit).Offset(offset).Order("updated_at DESC").Find(&evaluations).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list evaluations: %w", err)
	}
	return evaluations, total, nil
}

func (r *Repository) GetEvaluationByID(id int) (*model.Evaluation, error) {
	var evaluation model.Evaluation
	if err := r.db.Where("id = ? AND status != ?", id, consts.CommonDeleted).First(&evaluation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("evaluation with id %d: %w", id, consts.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to find evaluation with id %d: %w", id, err)
	}
	return &evaluation, nil
}

func (r *Repository) DeleteEvaluation(id int) error {
	result := r.db.Model(&model.Evaluation{}).
		Where("id = ? AND status != ?", id, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if err := result.Error; err != nil {
		return fmt.Errorf("failed to delete evaluation with id %d: %w", id, err)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("evaluation with id %d: %w", id, consts.ErrNotFound)
	}
	return nil
}
