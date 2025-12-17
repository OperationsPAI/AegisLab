package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aegis/client"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ===================== Task Redis =====================

// Redis key constants for task queues and indexes
const (
	DelayedQueueKey    = "task:delayed"          // Sorted set for delayed tasks
	ReadyQueueKey      = "task:ready"            // List for ready-to-execute tasks
	DeadLetterKey      = "task:dead"             // Sorted set for failed tasks
	TaskIndexKey       = "task:index"            // Hash mapping task IDs to their queue
	ConcurrencyLockKey = "task:concurrency_lock" // Counter for concurrency control
	LastBatchInfoKey   = "last_batch_info"       // Key for batch processing information
	MaxConcurrency     = 20                      // Maximum concurrent tasks
)

// ImmediateTask

// SubmitImmediateTask sends a task to the ready queue for immediate execution
func SubmitImmediateTask(ctx context.Context, taskData []byte, taskID string) error {
	redisCli := client.GetRedisClient()
	if err := redisCli.LPush(ctx, ReadyQueueKey, taskData).Err(); err != nil {
		return err
	}

	return redisCli.HSet(ctx, TaskIndexKey, taskID, ReadyQueueKey).Err()
}

// GetTask retrieves a task from the ready queue with blocking
func GetTask(ctx context.Context, timeout time.Duration) (string, error) {
	redisCli := client.GetRedisClient()
	result, err := redisCli.BRPop(ctx, timeout, ReadyQueueKey).Result()
	if err != nil {
		return "", err
	}

	return result[1], nil
}

// HandleFailedTask moves a failed task to the dead letter queue
func HandleFailedTask(ctx context.Context, taskData []byte, backoffSec int) error {
	deadLetterTime := time.Now().Add(time.Duration(backoffSec) * time.Second).Unix()
	redisCli := client.GetRedisClient()
	return redisCli.ZAdd(ctx, DeadLetterKey, redis.Z{
		Score:  float64(deadLetterTime),
		Member: taskData,
	}).Err()
}

// Delayed Task

// SubmitDelayedTask sends a task to the delayed queue for future execution
func SubmitDelayedTask(ctx context.Context, taskData []byte, taskID string, executeTime int64) error {
	redisCli := client.GetRedisClient()
	if err := redisCli.ZAdd(ctx, DelayedQueueKey, redis.Z{
		Score:  float64(executeTime),
		Member: taskData,
	}).Err(); err != nil {
		return err
	}

	return redisCli.HSet(ctx, TaskIndexKey, taskID, DelayedQueueKey).Err()
}

// ProcessDelayedTasks moves tasks from delayed queue to ready queue when their time arrives
func ProcessDelayedTasks(ctx context.Context) ([]string, error) {
	redisCli := client.GetRedisClient()
	now := time.Now().Unix()

	delayedTaskScript := redis.NewScript(`
		local tasks = redis.call('ZRANGEBYSCORE', KEYS[1], 0, ARGV[1])
		if #tasks > 0 then
			redis.call('ZREMRANGEBYSCORE', KEYS[1], 0, ARGV[1])
			redis.call('LPUSH', KEYS[2], unpack(tasks))
			-- Update task index
			for _, task in ipairs(tasks) do
				local t = cjson.decode(task)
				redis.call('HSET', KEYS[3], t.task_id, KEYS[2])
			end
		end
		return tasks
	`)

	result, err := delayedTaskScript.Run(ctx, redisCli,
		[]string{DelayedQueueKey, ReadyQueueKey, TaskIndexKey},
		now,
	).StringSlice()

	if err != nil && err != redis.Nil {
		return nil, err
	}

	return result, nil
}

// HandleCronRescheduleFailure moves a failed cron task to the dead letter queue
func HandleCronRescheduleFailure(ctx context.Context, taskData []byte) error {
	return client.GetRedisClient().ZAdd(ctx, DeadLetterKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: taskData,
	}).Err()
}

// AcquireConcurrencyLock attempts to acquire a lock for task execution
func AcquireConcurrencyLock(ctx context.Context) bool {
	redisCli := client.GetRedisClient()
	currentCount, _ := redisCli.Get(ctx, ConcurrencyLockKey).Int64()
	if currentCount >= MaxConcurrency {
		return false
	}
	return redisCli.Incr(ctx, ConcurrencyLockKey).Err() == nil
}

// InitConcurrencyLock initializes the concurrency lock counter
func InitConcurrencyLock(ctx context.Context) error {
	redisCli := client.GetRedisClient()
	return redisCli.Set(ctx, ConcurrencyLockKey, 0, 0).Err()
}

// ReleaseConcurrencyLock releases a lock after task execution
func ReleaseConcurrencyLock(ctx context.Context) {
	redisCli := client.GetRedisClient()
	if err := redisCli.Decr(ctx, ConcurrencyLockKey).Err(); err != nil {
		logrus.Warnf("error releasing concurrency lock: %v", err)
	}
}

// GetTaskQueue retrieves the queue a task is in
func GetTaskQueue(ctx context.Context, taskID string) (string, error) {
	return client.GetRedisClient().HGet(ctx, TaskIndexKey, taskID).Result()
}

// ListDelayedTasks lists all tasks in the delayed queue
func ListDelayedTasks(ctx context.Context, limit int64) ([]string, error) {
	delayedTasksWithScore, err := client.GetRedisZRangeByScoreWithScores(ctx, DelayedQueueKey, limit)
	if err != nil {
		return nil, err
	}

	taskDatas := make([]string, 0, len(delayedTasksWithScore))
	for _, z := range delayedTasksWithScore {
		taskData, ok := z.Member.(string)
		if !ok {
			return nil, fmt.Errorf("invalid delayed task data")
		}
		taskDatas = append(taskDatas, taskData)
	}

	return taskDatas, nil
}

// ListReadyTasks lists all tasks in the ready queue
func ListReadyTasks(ctx context.Context) ([]string, error) {
	return client.GetRedisListRange(ctx, ReadyQueueKey)
}

// RemoveFromList removes a task from a Redis list using Lua script
func RemoveFromList(ctx context.Context, key, taskID string) (bool, error) {
	cli := client.GetRedisClient()
	// Efficient list removal Lua script
	removeFromListScript := redis.NewScript(`
		local key = KEYS[1]
		local taskID = ARGV[1]
		local count = 0

		for i=0, redis.call('LLEN', key)-1 do
			local item = redis.call('LINDEX', key, i)
			if item then
				local task = cjson.decode(item)
				if task.task_id == taskID then
					redis.call('LSET', key, i, "__DELETED__")
					count = count + 1
				end
			end
		end

		if count > 0 then
			redis.call('LREM', key, count, "__DELETED__")
		end
		
		return count
	`)
	result, err := removeFromListScript.Run(ctx, cli, []string{key}, taskID).Int()
	if err != nil {
		return false, fmt.Errorf("failed to remove from list: %w", err)
	}

	return result > 0, nil
}

// RemoveFromZSet removes a task from a Redis sorted set
func RemoveFromZSet(ctx context.Context, key, taskID string) bool {
	cli := client.GetRedisClient()
	members, err := cli.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	}).Result()
	if err != nil {
		return false
	}

	for _, member := range members {
		var t dto.UnifiedTask
		if json.Unmarshal([]byte(member), &t) == nil && t.TaskID == taskID {
			if err := cli.ZRem(ctx, key, member).Err(); err != nil {
				logrus.Warnf("failed to remove from ZSet: %v", err)
				return false
			}
			return true
		}
	}

	return false
}

// DeleteTaskIndex removes a task from the task index
func DeleteTaskIndex(ctx context.Context, taskID string) error {
	return client.GetRedisClient().HDel(ctx, TaskIndexKey, taskID).Err()
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
		Preload("Project").
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

	query := db.Model(&database.Task{}).Preload("Project")
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
