package repository

import (
	"context"
	"fmt"
	"time"

	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	taskqueue "aegis/service/queue"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ===================== Task Redis =====================

// Redis key constants for task queues and indexes
const (
	DelayedQueueKey    = taskqueue.DelayedQueueKey
	ReadyQueueKey      = taskqueue.ReadyQueueKey
	DeadLetterKey      = taskqueue.DeadLetterKey
	TaskIndexKey       = taskqueue.TaskIndexKey
	ConcurrencyLockKey = taskqueue.ConcurrencyLockKey
	LastBatchInfoKey   = taskqueue.LastBatchInfoKey
	MaxConcurrency     = taskqueue.MaxConcurrency
)

// ImmediateTask

// SubmitImmediateTask sends a task to the ready queue for immediate execution
func SubmitImmediateTask(ctx context.Context, taskData []byte, taskID string) error {
	return taskqueue.SubmitImmediateTask(ctx, taskData, taskID)
}

// GetTask retrieves a task from the ready queue with blocking
func GetTask(ctx context.Context, timeout time.Duration) (string, error) {
	return taskqueue.GetTask(ctx, timeout)
}

// HandleFailedTask moves a failed task to the dead letter queue
func HandleFailedTask(ctx context.Context, taskData []byte, backoffSec int) error {
	return taskqueue.HandleFailedTask(ctx, taskData, backoffSec)
}

// Delayed Task

// SubmitDelayedTask sends a task to the delayed queue for future execution
func SubmitDelayedTask(ctx context.Context, taskData []byte, taskID string, executeTime int64) error {
	return taskqueue.SubmitDelayedTask(ctx, taskData, taskID, executeTime)
}

// ProcessDelayedTasks moves tasks from delayed queue to ready queue when their time arrives
func ProcessDelayedTasks(ctx context.Context) ([]string, error) {
	return taskqueue.ProcessDelayedTasks(ctx)
}

// HandleCronRescheduleFailure moves a failed cron task to the dead letter queue
func HandleCronRescheduleFailure(ctx context.Context, taskData []byte) error {
	return taskqueue.HandleCronRescheduleFailure(ctx, taskData)
}

// AcquireConcurrencyLock attempts to acquire a lock for task execution
func AcquireConcurrencyLock(ctx context.Context) bool {
	return taskqueue.AcquireConcurrencyLock(ctx)
}

// InitConcurrencyLock initializes the concurrency lock counter
func InitConcurrencyLock(ctx context.Context) error {
	return taskqueue.InitConcurrencyLock(ctx)
}

// ReleaseConcurrencyLock releases a lock after task execution
func ReleaseConcurrencyLock(ctx context.Context) {
	taskqueue.ReleaseConcurrencyLock(ctx)
}

// GetTaskQueue retrieves the queue a task is in
func GetTaskQueue(ctx context.Context, taskID string) (string, error) {
	return taskqueue.GetTaskQueue(ctx, taskID)
}

// ListDelayedTasks lists all tasks in the delayed queue
func ListDelayedTasks(ctx context.Context, limit int64) ([]string, error) {
	return taskqueue.ListDelayedTasks(ctx, limit)
}

// ListReadyTasks lists all tasks in the ready queue
func ListReadyTasks(ctx context.Context) ([]string, error) {
	return taskqueue.ListReadyTasks(ctx)
}

// RemoveFromList removes a task from a Redis list using Lua script
func RemoveFromList(ctx context.Context, key, taskID string) (bool, error) {
	return taskqueue.RemoveFromList(ctx, key, taskID)
}

// RemoveFromZSet removes a task from a Redis sorted set
func RemoveFromZSet(ctx context.Context, key, taskID string) bool {
	return taskqueue.RemoveFromZSet(ctx, key, taskID)
}

// DeleteTaskIndex removes a task from the task index
func DeleteTaskIndex(ctx context.Context, taskID string) error {
	return taskqueue.DeleteTaskIndex(ctx, taskID)
}

// ===================== Task Database =====================

// BatchDeleteTasks marks multiple tasks as deleted in batch
func BatchDeleteTasks(db *gorm.DB, taskIDs []string) error {
	if len(taskIDs) == 0 {
		return nil
	}

	if err := db.Model(&database.Task{}).
		Where("id IN (?) AND status != ?", taskIDs, consts.CommonDeleted).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return fmt.Errorf("failed to batch delete tasks: %w", err)
	}
	return nil
}

// GetTaskByID retrieves a task by its ID with preloaded associations
func GetTaskByID(db *gorm.DB, taskID string) (*database.Task, error) {
	var result database.Task
	if err := db.
		Preload("FaultInjection.Benchmark.Container").
		Preload("FaultInjection.Pedestal.Container").
		Preload("Execution.AlgorithmVersion.Container").
		Preload("Execution.Datapack").
		Preload("Execution.DatasetVersion").
		Where("id = ? AND status != ?", taskID, consts.CommonDeleted).
		First(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to find task with id %s: %w", taskID, err)
	}
	return &result, nil
}

// GetTaskWithParentByID retrieves a task along with its parent task by ID
func GetTaskWithParentByID(db *gorm.DB, taskID string) (*database.Task, error) {
	var result database.Task
	if err := db.
		Preload("ParentTask").
		Where("id = ? AND status != ?", taskID, consts.CommonDeleted).
		First(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to find task with id %s: %w", taskID, err)
	}
	return &result, nil
}

// GetParentTaskLevelByID retrieves the level of a parent task by its ID
func GetParentTaskLevelByID(db *gorm.DB, parentTaskID string) (int, error) {
	var result database.Task
	if err := db.
		Select("level").
		Where("id = ? AND status != ?", parentTaskID, consts.CommonDeleted).
		First(&result).Error; err != nil {
		return 0, fmt.Errorf("failed to find parent task with id %s: %w", parentTaskID, err)
	}
	return result.Level, nil
}

// ListTasks lists tasks based on filter and pagination with preloaded associations
func ListTasks(db *gorm.DB, limit, offset int, filterOptions *dto.ListTaskFilters) ([]database.Task, int64, error) {
	var tasks []database.Task
	var total int64

	query := db.Model(&database.Task{})
	if filterOptions.Immediate != nil {
		query = query.Where("immediate = ?", *filterOptions.Immediate)
	}
	if filterOptions.TaskType != nil {
		query = query.Where("type = ?", *filterOptions.TaskType)
	}
	if filterOptions.TraceID != "" {
		query = query.Where("trace_id = ?", filterOptions.TraceID)
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
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, total, nil
}

// ListTasksByTimeRange retrieves tasks created within a specific time range
func ListTasksByTimeRange(db *gorm.DB, startTime, endTime time.Time) ([]database.Task, error) {
	var tasks []database.Task
	err := database.DB.Model(&database.Task{}).
		Where("created_at >= ? AND created_at <= ? AND status != ?", startTime, endTime, consts.CommonDeleted).
		Find(&tasks).Error
	return tasks, err
}

// UpdateTaskState updates the task state in the database
func UpdateTaskState(db *gorm.DB, ctx context.Context, taskID string, state consts.TaskState) error {
	return db.WithContext(ctx).Model(&database.Task{}).
		Where("id = ?", taskID).
		Update("state", state).Error
}

// UpdateTaskStatus updates the task status in the database
func UpdateTaskStatus(db *gorm.DB, ctx context.Context, taskID string, status int) error {
	return db.WithContext(ctx).Model(&database.Task{}).
		Where("id = ?", taskID).
		Update("status", status).Error
}

// UpsertTask inserts or updates a task in the database
func UpsertTask(db *gorm.DB, task *database.Task) error {
	if err := db.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"execute_time",
				"state",
				"updated_at",
			}),
		},
	).Create(task).Error; err != nil {
		return fmt.Errorf("failed to upsert task: %w", err)
	}
	return nil
}

// GetTaskStatistics returns statistics about tasks
func GetTaskStatistics() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total tasks
	var total int64
	if err := database.DB.Model(&database.Task{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total tasks: %v", err)
	}
	stats["total"] = total

	// Tasks by status
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	var statusCounts []StatusCount
	err := database.DB.Model(&database.Task{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count tasks by status: %v", err)
	}

	for _, sc := range statusCounts {
		stats[sc.Status] = sc.Count
	}

	// Tasks by type
	type TypeCount struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}

	var typeCounts []TypeCount
	err = database.DB.Model(&database.Task{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Find(&typeCounts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count tasks by type: %v", err)
	}

	for _, tc := range typeCounts {
		stats[tc.Type+"_tasks"] = tc.Count
	}

	return stats, nil
}

// GetRecentTaskActivity returns task activity for the last N days
func GetRecentTaskActivity(days int) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Last N days activity
	startDate := time.Now().AddDate(0, 0, -days)
	var recentCount int64
	if err := database.DB.Model(&database.Task{}).Where("created_at >= ?", startDate).Count(&recentCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count recent tasks: %v", err)
	}
	stats[fmt.Sprintf("last_%d_days", days)] = recentCount

	// Today's tasks
	today := time.Now().Truncate(24 * time.Hour)
	var todayCount int64
	if err := database.DB.Model(&database.Task{}).Where("created_at >= ?", today).Count(&todayCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count today's tasks: %v", err)
	}
	stats["today"] = todayCount

	return stats, nil
}
