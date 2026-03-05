package repository

import (
	"fmt"
	"time"

	"aegis/consts"
	"aegis/database"
	"aegis/dto"

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
		Preload("Project").
		Preload("Tasks", func(db *gorm.DB) *gorm.DB {
			return db.Order("level ASC, sequence ASC")
		}).
		Where("id = ? AND status != ?", traceID, consts.CommonDeleted).
		First(&trace).Error; err != nil {
		return nil, err
	}
	return &trace, nil
}

// ListTraces lists traces based on filter and pagination with preloaded associations
func ListTraces(db *gorm.DB, limit, offset int, filterOptions *dto.ListTraceFilters) ([]database.Trace, int64, error) {
	var traces []database.Trace
	var total int64

	query := db.Model(&database.Trace{}).Preload("Project")
	if filterOptions.TraceType != nil {
		query = query.Where("type = ?", *filterOptions.TraceType)
	}
	if filterOptions.GroupID != "" {
		query = query.Where("group_id = ?", filterOptions.GroupID)
	}
	if filterOptions.ProjectID > 0 {
		query = query.Where("project_id = ?", filterOptions.ProjectID)
	}
	if filterOptions.State != nil {
		query = query.Where("state = ?", *filterOptions.State)
	}
	if filterOptions.Status != nil {
		query = query.Where("status = ?", *filterOptions.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count traces: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&traces).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list traces: %w", err)
	}

	return traces, total, nil
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

// CountTracesByGroupID counts the total number of non-deleted traces in a group
func CountTracesByGroupID(db *gorm.DB, groupID string) (int64, error) {
	var count int64
	if err := db.Model(&database.Trace{}).
		Where("group_id = ? AND status != ?", groupID, consts.CommonDeleted).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
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
