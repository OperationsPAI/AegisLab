package repository

import (
	"fmt"
	"strings"

	"aegis/database"

	"gorm.io/gorm"
)

// isTableNotExistError checks if the error indicates the table does not exist.
// This handles the case where the SDK tables have not been created yet.
func isTableNotExistError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "doesn't exist") ||
		strings.Contains(msg, "does not exist") ||
		strings.Contains(msg, "no such table")
}

// ListSDKEvaluations returns paginated SDK evaluation samples filtered by exp_id and stage.
func ListSDKEvaluations(db *gorm.DB, expID string, stage string, limit, offset int) ([]database.SDKEvaluationSample, int64, error) {
	var items []database.SDKEvaluationSample
	var total int64

	query := db.Model(&database.SDKEvaluationSample{})

	if expID != "" {
		query = query.Where("exp_id = ?", expID)
	}
	if stage != "" {
		query = query.Where("stage = ?", stage)
	}

	if err := query.Count(&total).Error; err != nil {
		if isTableNotExistError(err) {
			return []database.SDKEvaluationSample{}, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to count SDK evaluation samples: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("id DESC").Find(&items).Error; err != nil {
		if isTableNotExistError(err) {
			return []database.SDKEvaluationSample{}, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to list SDK evaluation samples: %w", err)
	}

	return items, total, nil
}

// GetSDKEvaluationByID returns a single SDK evaluation sample by its ID.
func GetSDKEvaluationByID(db *gorm.DB, id int) (*database.SDKEvaluationSample, error) {
	var item database.SDKEvaluationSample
	if err := db.Where("id = ?", id).First(&item).Error; err != nil {
		if isTableNotExistError(err) {
			return nil, fmt.Errorf("SDK evaluation sample with id %d not found (table does not exist)", id)
		}
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("SDK evaluation sample with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to find SDK evaluation sample with id %d: %w", id, err)
	}
	return &item, nil
}

// ListSDKExperiments returns all distinct exp_id values from the evaluation_data table.
func ListSDKExperiments(db *gorm.DB) ([]string, error) {
	var expIDs []string
	if err := db.Model(&database.SDKEvaluationSample{}).Distinct("exp_id").Pluck("exp_id", &expIDs).Error; err != nil {
		if isTableNotExistError(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list SDK experiments: %w", err)
	}
	return expIDs, nil
}

// ListSDKDatasetSamples returns paginated SDK dataset samples filtered by dataset name.
func ListSDKDatasetSamples(db *gorm.DB, dataset string, limit, offset int) ([]database.SDKDatasetSample, int64, error) {
	var items []database.SDKDatasetSample
	var total int64

	query := db.Model(&database.SDKDatasetSample{})

	if dataset != "" {
		query = query.Where("dataset = ?", dataset)
	}

	if err := query.Count(&total).Error; err != nil {
		if isTableNotExistError(err) {
			return []database.SDKDatasetSample{}, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to count SDK dataset samples: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("id DESC").Find(&items).Error; err != nil {
		if isTableNotExistError(err) {
			return []database.SDKDatasetSample{}, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to list SDK dataset samples: %w", err)
	}

	return items, total, nil
}
