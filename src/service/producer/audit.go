package producer

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// GetAuditLogDetail retrieves detailed information about a specific audit log by ID
func GetAuditLogDetail(id int) (*dto.AuditLogDetailResp, error) {
	log, err := repository.GetAuditLogByID(database.DB, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: audit log with ID %d not found", consts.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	return dto.NewAuditLogDetailResp(log), nil
}

// ListAuditLogs retrieves audit logs with pagination and filtering
func ListAuditLogs(req *dto.ListAuditLogReq) (*dto.ListResp[dto.AuditLogResp], error) {
	limit, offset := req.ToGormParams()
	filterOptions := req.ToFilterOptions()

	logs, total, err := repository.ListAuditLogs(database.DB, limit, offset, filterOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}

	logResps := make([]dto.AuditLogResp, 0, len(logs))
	for i := range logs {
		logResps = append(logResps, *dto.NewAuditLogResp(&logs[i]))
	}

	resp := dto.ListResp[dto.AuditLogResp]{
		Items:      logResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// LogFailedAction logs a failed action with error message
func LogFailedAction(ipAddress, userAgent, action, errorMsg string, duration, userID int, resourceName consts.ResourceName) error {
	if resourceName == "" {
		return fmt.Errorf("resource name cannot be empty")
	}

	log := &database.AuditLog{
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Duration:  duration,
		Action:    action,
		ErrorMsg:  errorMsg,
		UserID:    userID,
		State:     consts.AuditLogStateFailed,
		Status:    consts.CommonEnabled,
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		resource, err := repository.GetResourceByName(tx, resourceName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: resource %s not found", consts.ErrNotFound, resourceName)
			}
			return fmt.Errorf("failed to get resource: %w", err)
		}

		log.ResourceID = resource.ID

		if err := repository.CreateAuditLog(database.DB, log); err != nil {
			return fmt.Errorf("failed to log failed action: %w", err)
		}
		return nil
	})
}

// LogSystemAction logs a system action (no user involved)
func LogSystemAction(action, details string, resourceName consts.ResourceName) error {
	if resourceName == "" {
		return fmt.Errorf("resource name cannot be empty")
	}

	log := &database.AuditLog{
		IPAddress: "127.0.0.1",
		UserAgent: "SYSTEM",
		Action:    action,
		Details:   details,
		State:     consts.AuditLogStateSuccess,
		Status:    consts.CommonEnabled,
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		resource, err := repository.GetResourceByName(tx, resourceName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: resource %s not found", consts.ErrNotFound, resourceName)
			}
			return fmt.Errorf("failed to get resource: %w", err)
		}

		log.ResourceID = resource.ID

		if err := repository.CreateAuditLog(database.DB, log); err != nil {
			return fmt.Errorf("failed to log system action: %w", err)
		}
		return nil
	})
}

// LogUserAction logs an action performed by a user
func LogUserAction(ipAddress, userAgent, action, details string, duration, userID int, resourceName consts.ResourceName) error {
	if resourceName == "" {
		return fmt.Errorf("resource name cannot be empty")
	}

	log := &database.AuditLog{
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Duration:  duration,
		Action:    action,
		Details:   details,
		UserID:    userID,
		State:     consts.AuditLogStateSuccess,
		Status:    consts.CommonEnabled,
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		resource, err := repository.GetResourceByName(tx, resourceName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: resource %s not found", consts.ErrNotFound, resourceName)
			}
			return fmt.Errorf("failed to get resource: %w", err)
		}

		log.ResourceID = resource.ID

		if err := repository.CreateAuditLog(database.DB, log); err != nil {
			return fmt.Errorf("failed to log user action: %w", err)
		}
		return nil
	})
}
