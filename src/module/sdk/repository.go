package sdk

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListSDKEvaluations(expID, stage string, limit, offset int) ([]SDKEvaluationSample, int64, error) {
	var (
		items []SDKEvaluationSample
		total int64
	)

	query := r.db.Model(&SDKEvaluationSample{})
	if expID != "" {
		query = query.Where("exp_id = ?", expID)
	}
	if stage != "" {
		query = query.Where("stage = ?", stage)
	}

	if err := query.Count(&total).Error; err != nil {
		if isTableNotExistError(err) {
			return []SDKEvaluationSample{}, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to count SDK evaluation samples: %w", err)
	}
	if err := query.Limit(limit).Offset(offset).Order("id DESC").Find(&items).Error; err != nil {
		if isTableNotExistError(err) {
			return []SDKEvaluationSample{}, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to list SDK evaluation samples: %w", err)
	}

	return items, total, nil
}

func (r *Repository) GetSDKEvaluationByID(id int) (*SDKEvaluationSample, error) {
	var item SDKEvaluationSample
	if err := r.db.Where("id = ?", id).First(&item).Error; err != nil {
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

func (r *Repository) ListSDKExperiments() ([]string, error) {
	var expIDs []string
	if err := r.db.Model(&SDKEvaluationSample{}).Distinct("exp_id").Pluck("exp_id", &expIDs).Error; err != nil {
		if isTableNotExistError(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list SDK experiments: %w", err)
	}
	return expIDs, nil
}

func (r *Repository) ListSDKDatasetSamples(dataset string, limit, offset int) ([]SDKDatasetSample, int64, error) {
	var (
		items []SDKDatasetSample
		total int64
	)

	query := r.db.Model(&SDKDatasetSample{})
	if dataset != "" {
		query = query.Where("dataset = ?", dataset)
	}

	if err := query.Count(&total).Error; err != nil {
		if isTableNotExistError(err) {
			return []SDKDatasetSample{}, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to count SDK dataset samples: %w", err)
	}
	if err := query.Limit(limit).Offset(offset).Order("id DESC").Find(&items).Error; err != nil {
		if isTableNotExistError(err) {
			return []SDKDatasetSample{}, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to list SDK dataset samples: %w", err)
	}

	return items, total, nil
}

func isTableNotExistError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "doesn't exist") ||
		strings.Contains(msg, "does not exist") ||
		strings.Contains(msg, "no such table")
}
