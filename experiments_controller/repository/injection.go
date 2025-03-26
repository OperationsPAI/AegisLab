package repository

import (
	"fmt"
	"time"

	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
)

func GetInjectionRecordByDataset(name string, status ...int) (*database.FaultInjectionSchedule, error) {
	var record database.FaultInjectionSchedule
	query := database.DB.
		Select("id, config, status, start_time, end_time").
		Where("injection_name = ?", name)

	if len(status) == 0 {
		query = query.Where("status != ?", consts.DatasetDeleted)
	} else if len(status) == 1 {
		query = query.Where("status = ?", status[0])
	} else {
		query = query.Where("status IN ?", status)
	}

	err := query.First(&record).Error
	return &record, err
}

func UpdateStatusByDataset(name string, status int) error {
	return updateRecord(name, map[string]any{
		"status": status,
	})
}

func UpdateTimeByDataset(name string, startTime, endTime time.Time) error {
	return updateRecord(name, map[string]any{
		"start_time": startTime,
		"end_time":   endTime,
		"status":     consts.DatasetInjectSuccess,
	})
}

func updateRecord(name string, updates map[string]any) error {
	if len(updates) == 0 {
		return fmt.Errorf("empty update fields")
	}

	var record database.FaultInjectionSchedule
	result := database.DB.
		Model(&record).
		Where("injection_name = ?", name).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update record: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no records updated")
	}

	return nil
}
