package repository

import (
	"time"

	"aegis/database"

	"gorm.io/gorm"
)

// =====================================================================
// Database Operations
// =====================================================================

// ListTraceIDs retrieves distinct trace IDs from tasks within the specified time range
func ListTraceIDs(db *gorm.DB, startTime, endTime *time.Time) ([]string, error) {
	var traceIDs []string

	query := db.Model(&database.Task{}).Select("DISTINCT trace_id")
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	if err := query.Find(&traceIDs).Error; err != nil {
		return nil, err
	}

	return traceIDs, nil
}
