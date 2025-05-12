package repository

import (
	"errors"
	"fmt"

	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"gorm.io/gorm"
)

// TODO
func CreateDetectorForMatchingInjection(prefix string) (bool, error) {
	query := database.DB.Model(&database.FaultInjectionSchedule{}).
		Select("id").
		Where("injection_name LIKE ?", prefix+"%").
		Order("created_at ASC")

	var record database.FaultInjectionSchedule
	if err := query.First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, err
	}

	query = database.DB.Model(&database.Detector{}).
		Select("").
		Where("")

	return true, nil
}

func GetDetectorRecordByDatasetID(datasetID int) (dto.DetectorRecord, error) {
	var record dto.DetectorRecord

	selectFields := `
        detectors.span_name AS span_name,
        detectors.issues AS issues,
        detectors.abnormal_avg_duration AS abnormal_avg_duration,
        detectors.normal_avg_duration AS normal_avg_duration,
        detectors.abnormal_succ_rate AS abnormal_succ_rate,
        detectors.normal_succ_rate AS normal_succ_rate,
        detectors.abnormal_p90 AS abnormal_p90,
        detectors.normal_p90 AS normal_p90,
        detectors.abnormal_p95 AS abnormal_p95,
        detectors.normal_p95 AS normal_p95,
        detectors.abnormal_p99 AS abnormal_p99,
        detectors.normal_p99 AS normal_p99
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
