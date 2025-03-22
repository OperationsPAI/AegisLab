package repository

import (
	"errors"
	"fmt"

	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"gorm.io/gorm"
)

func GetDetectorRecordByDatasetID(datasetID int) (dto.DetectorRecord, error) {
	var record dto.DetectorRecord

	selectFields := `
        detectors.span_name AS span_name,
        detectors.issues AS issues,
        detectors.avg_duration AS avg_duration,
        detectors.succ_rate AS succ_rate,
        detectors.p90 AS p90,
        detectors.p95 AS p95,
        detectors.p99 AS p99
    `

	query := database.DB.
		Table("detectors").
		Select(selectFields).
		Joins(`
            LEFT JOIN execution_results 
            ON detectors.execution_id = execution_results.id
        `).
		Where("execution_results.dataset = ?", datasetID).
		Order("detectors.created_at DESC").Limit(1)

	if err := query.Find(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.DetectorRecord{}, nil
		}

		return dto.DetectorRecord{}, fmt.Errorf("database query error: %w", err)
	}

	return record, nil
}
