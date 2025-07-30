package repository

import (
	"fmt"
	"time"

	"github.com/LGU-SE-Internal/rcabench/database"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     *int      `gorm:"index" json:"user_id"`                   // User who performed the action (nullable for system actions)
	Username   string    `gorm:"index" json:"username"`                  // Username for quick reference
	Action     string    `gorm:"not null;index" json:"action"`           // Action performed (CREATE, UPDATE, DELETE, etc.)
	Resource   string    `gorm:"not null;index" json:"resource"`         // Resource type (USER, ROLE, PERMISSION, etc.)
	ResourceID *int      `json:"resource_id"`                            // ID of the affected resource
	Details    string    `gorm:"type:text" json:"details"`               // Additional details in JSON format
	IPAddress  string    `gorm:"index" json:"ip_address"`                // IP address of the client
	UserAgent  string    `gorm:"type:text" json:"user_agent"`            // User agent of the client
	Status     string    `gorm:"not null;index" json:"status"`           // SUCCESS, FAILED, WARNING
	ErrorMsg   string    `gorm:"type:text" json:"error_msg,omitempty"`   // Error message if status is FAILED
	Duration   int64     `json:"duration"`                               // Duration in milliseconds
	CreatedAt  time.Time `gorm:"autoCreateTime;index" json:"created_at"` // When the action was performed
}

// CreateAuditLogTable creates the audit log table if it doesn't exist
func CreateAuditLogTable() error {
	return database.DB.AutoMigrate(&AuditLog{})
}

// CreateAuditLog creates a new audit log entry
func CreateAuditLog(log *AuditLog) error {
	if err := database.DB.Create(log).Error; err != nil {
		return fmt.Errorf("failed to create audit log: %v", err)
	}
	return nil
}

// LogUserAction logs a user action with basic information
func LogUserAction(userID int, username, action, resource string, resourceID *int, details string, ipAddress, userAgent string) error {
	log := &AuditLog{
		UserID:     &userID,
		Username:   username,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Status:     "SUCCESS",
	}

	return CreateAuditLog(log)
}

// LogSystemAction logs a system action (no user involved)
func LogSystemAction(action, resource string, resourceID *int, details string) error {
	log := &AuditLog{
		Username:   "SYSTEM",
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
		IPAddress:  "127.0.0.1",
		UserAgent:  "SYSTEM",
		Status:     "SUCCESS",
	}

	return CreateAuditLog(log)
}

// LogFailedAction logs a failed action with error message
func LogFailedAction(userID *int, username, action, resource string, resourceID *int, errorMsg, ipAddress, userAgent string) error {
	log := &AuditLog{
		UserID:     userID,
		Username:   username,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Status:     "FAILED",
		ErrorMsg:   errorMsg,
	}

	return CreateAuditLog(log)
}

// GetAuditLogs retrieves audit logs with pagination and filtering
func GetAuditLogs(page, pageSize int, userID *int, action, resource, status string, startTime, endTime *time.Time) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	query := database.DB.Model(&AuditLog{})

	// Apply filters
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %v", err)
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get audit logs: %v", err)
	}

	return logs, total, nil
}

// GetAuditLogsByUser retrieves audit logs for a specific user
func GetAuditLogsByUser(userID int, page, pageSize int) ([]AuditLog, int64, error) {
	return GetAuditLogs(page, pageSize, &userID, "", "", "", nil, nil)
}

// GetAuditLogsByResource retrieves audit logs for a specific resource
func GetAuditLogsByResource(resource string, resourceID *int, page, pageSize int) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	query := database.DB.Model(&AuditLog{}).Where("resource = ?", resource)
	if resourceID != nil {
		query = query.Where("resource_id = ?", *resourceID)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs for resource %s: %v", resource, err)
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get audit logs for resource %s: %v", resource, err)
	}

	return logs, total, nil
}

// GetAuditLogStatistics returns statistics about audit logs
func GetAuditLogStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total logs count
	var totalCount int64
	if err := database.DB.Model(&AuditLog{}).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count total audit logs: %v", err)
	}
	stats["total"] = totalCount

	// Count by status
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var statusCounts []StatusCount
	err := database.DB.Model(&AuditLog{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get status counts: %v", err)
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
	err = database.DB.Model(&AuditLog{}).
		Select("action, COUNT(*) as count").
		Group("action").
		Find(&actionCounts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get action counts: %v", err)
	}

	actionMap := make(map[string]int64)
	for _, ac := range actionCounts {
		actionMap[ac.Action] = ac.Count
	}
	stats["by_action"] = actionMap

	// Recent activity (last 24 hours)
	last24h := time.Now().Add(-24 * time.Hour)
	var recentCount int64
	if err := database.DB.Model(&AuditLog{}).Where("created_at >= ?", last24h).Count(&recentCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count recent audit logs: %v", err)
	}
	stats["last_24h"] = recentCount

	return stats, nil
}

// CleanupOldAuditLogs removes audit logs older than specified days
func CleanupOldAuditLogs(days int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	result := database.DB.Where("created_at < ?", cutoffDate).Delete(&AuditLog{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old audit logs: %v", result.Error)
	}
	return result.RowsAffected, nil
}
