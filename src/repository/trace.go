package repository

import (
	"fmt"
	"time"

	"aegis/consts"
	"aegis/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// =====================================================================
// Database Operations
// =====================================================================

// GetTraceByID retrieves a trace by its trace ID
func GetTraceByID(db *gorm.DB, traceID string) (*database.Trace, error) {
	var trace database.Trace
	if err := db.Model(&database.Trace{}).
		Preload("Tasks", func(db *gorm.DB) *gorm.DB {
			return db.Order("level ASC, sequence ASC")
		}).
		Where("id = ? AND status != ?", traceID, consts.CommonDeleted).
		First(&trace).Error; err != nil {
		return nil, err
	}
	return &trace, nil
}

// GetTracesByGroupID retrieves all traces belonging to a specific group
func GetTracesByGroupID(db *gorm.DB, groupID string) ([]database.Trace, error) {
	var traces []database.Trace
	if err := db.Model(&database.Trace{}).
		Preload("Tasks").
		Where("group_id = ? AND status != ?", groupID, consts.CommonDeleted).
		Order("start_time DESC").
		Find(&traces).Error; err != nil {
		return nil, err
	}
	return traces, nil
}

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

// UpsertTrace inserts or updates a trace in the database
func UpsertTrace(db *gorm.DB, trace *database.Trace) error {
	if err := db.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"last_event",
				"end_time",
				"state",
				"updated_at",
			}),
		},
	).Create(trace).Error; err != nil {
		return fmt.Errorf("failed to upsert task: %w", err)
	}
	return nil
}
