package repository

import (
	"fmt"

	"aegis/consts"
	"aegis/database"

	"gorm.io/gorm"
)

func ListEvaluations(db *gorm.DB, limit, offset int) ([]database.Evaluation, int64, error) {
	var evaluations []database.Evaluation
	var total int64

	query := db.Model(&database.Evaluation{}).
		Where("status != ?", consts.CommonDeleted)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count evaluations: %w", err)
	}

	if err := query.Select("id, project_id, algorithm_name, algorithm_version, datapack_name, dataset_name, dataset_version, eval_type, precision, recall, f1_score, accuracy, status, created_at, updated_at").Limit(limit).Offset(offset).Order("updated_at DESC").Find(&evaluations).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list evaluations: %w", err)
	}

	return evaluations, total, nil
}

func GetEvaluationByID(db *gorm.DB, id int) (*database.Evaluation, error) {
	var evaluation database.Evaluation
	if err := db.
		Where("id = ? AND status != ?", id, consts.CommonDeleted).
		First(&evaluation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("evaluation with id %d: %w", id, consts.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to find evaluation with id %d: %w", id, err)
	}
	return &evaluation, nil
}

func CreateEvaluation(db *gorm.DB, eval *database.Evaluation) error {
	if err := db.Create(eval).Error; err != nil {
		return fmt.Errorf("failed to create evaluation: %w", err)
	}
	return nil
}

func DeleteEvaluation(db *gorm.DB, id int) error {
	result := db.Model(&database.Evaluation{}).
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
