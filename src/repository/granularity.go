package repository

import (
	"errors"
	"fmt"

	"aegis/consts"
	"aegis/database"

	"gorm.io/gorm"
)

// ListGranularityResultsByExecutionID lists granularity results for a specific execution ID
func ListGranularityResultsByExecutionID(db *gorm.DB, executionID int) ([]database.GranularityResult, error) {
	var results []database.GranularityResult
	if err := db.
		Where("execution_id = ? AND status != ?", executionID, consts.CommonDeleted).
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to list granularity results for execution %d: %w", executionID, err)
	}
	return results, nil
}

// SaveGranularityResults saves multiple granularity results
func SaveGranularityResults(db *gorm.DB, results []database.GranularityResult) error {
	if len(results) == 0 {
		return fmt.Errorf("no granularity results to create")
	}

	for i := range results {
		resultPtr := &results[i]
		err := db.Omit(containerVersionOmitFields).Create(resultPtr).Error
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: index %d", consts.ErrAlreadyExists, i)
			}
			return fmt.Errorf("failed to create record index %d: %w", i, err)
		}
	}

	return nil
}
