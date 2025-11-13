package repository

import (
	"fmt"
	"time"

	"aegis/database"
	"aegis/dto"

	"gorm.io/gorm"
)

// CreateAuditLog creates a new audit log entry
func CreateAuditLog(db *gorm.DB, log *database.AuditLog) error {
	if err := db.Create(log).Error; err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

// GetAuditLogByID retrieves a single audit log by ID
func GetAuditLogByID(db *gorm.DB, id int) (*database.AuditLog, error) {
	var log database.AuditLog
	err := db.Where("id = ?", id).First(&log).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}
	return &log, nil
}

// ListAuditLogs retrieves audit logs with pagination and filtering
func ListAuditLogs(db *gorm.DB, limit, offset int, filterOptions *dto.ListAuditLogFilters) ([]database.AuditLog, int64, error) {
	var logs []database.AuditLog
	var total int64

	query := db.Model(&database.AuditLog{}).Preload("User").Preload("Resource")
	if filterOptions != nil {
		if filterOptions.Action != "" {
			query = query.Where("action = ?", filterOptions.Action)
		}
		if filterOptions.IpAddress != "" {
			query = query.Where("ip_address = ?", filterOptions.IpAddress)
		}
		if filterOptions.UserID != 0 {
			query = query.Where("user_id = ?", filterOptions.UserID)
		}
		if filterOptions.ResourceID != 0 {
			query = query.Where("resource_id = ?", filterOptions.ResourceID)
		}
		if filterOptions.State != nil {
			query = query.Where("state = ?", *filterOptions.State)
		}
		if filterOptions.Status != nil {
			query = query.Where("status = ?", *filterOptions.Status)
		}
		if filterOptions.StartTime != nil {
			query = query.Where("created_at >= ?", *filterOptions.StartTime)
		}
		if filterOptions.EndTime != nil {
			query = query.Where("created_at <= ?", *filterOptions.EndTime)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return logs, total, nil
}

// GetAuditLogStatistics returns statistics about audit logs
func GetAuditLogStatistics() (map[string]any, error) {
	stats := make(map[string]interface{})

	// Total logs count
	var totalCount int64
	if err := database.DB.Model(&database.AuditLog{}).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count total audit logs: %w", err)
	}
	stats["total"] = totalCount

	// Count by status
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var statusCounts []StatusCount
	err := database.DB.Model(&database.AuditLog{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}

	statusMap := make(map[string]int64)
	for _, sc := range statusCounts {
		statusMap[sc.Status] = sc.Count
	}
	stats["by_status"] = statusMap

	// Count by action
	type ActionCount struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}
	var actionCounts []ActionCount
	err = database.DB.Model(&database.AuditLog{}).
		Select("action, COUNT(*) as count").
		Group("action").
		Find(&actionCounts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get action counts: %w", err)
	}

	actionMap := make(map[string]int64)
	for _, ac := range actionCounts {
		actionMap[ac.Action] = ac.Count
	}
	stats["by_action"] = actionMap

	// Recent activity (last 24 hours)
	last24h := time.Now().Add(-24 * time.Hour)
	var recentCount int64
	if err := database.DB.Model(&database.AuditLog{}).Where("created_at >= ?", last24h).Count(&recentCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count recent audit logs: %w", err)
	}
	stats["last_24h"] = recentCount

	return stats, nil
}
