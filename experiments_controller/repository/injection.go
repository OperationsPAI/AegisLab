package repository

import "github.com/CUHK-SE-Group/rcabench/database"

func GetInjectionRecordByDataset(name string) (*database.FaultInjectionSchedule, error) {
	var record database.FaultInjectionSchedule
	err := database.DB.
		Select("id, config, start_time, end_time").
		Where("injection_name = ? AND status = ?", name, 1).
		First(&record).
		Error
	return &record, err
}
