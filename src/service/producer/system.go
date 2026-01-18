package producer

import (
	"aegis/client"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// InspectLock retrieves the current lock status of all namespaces
func InspectLock(ctx context.Context) (*dto.ListNamespaceLockResp, error) {
	redisClient := client.GetRedisClient()

	// Get all namespaces
	namespaces, err := redisClient.SMembers(ctx, consts.NamespacesKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces from Redis: %v", err)
	}

	nsMap := make(map[string]dto.NsMonitorItem, len(namespaces))

	// Get data for each namespace
	for _, ns := range namespaces {
		nsKey := fmt.Sprintf(consts.NamespaceKeyPattern, ns)
		values, err := redisClient.HGetAll(ctx, nsKey).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get data for namespace %s: %v", ns, err)
		}

		endTimeUnix, err := strconv.ParseInt(values["end_time"], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time format for namespace %s: %v", ns, err)
		}

		// Get status, default to enabled for backward compatibility
		status := consts.CommonEnabled
		if statusStr, ok := values["status"]; ok {
			statusInt, err := strconv.Atoi(statusStr)
			if err == nil {
				status = consts.StatusType(statusInt)
			}
		}

		nsMap[ns] = dto.NsMonitorItem{
			LockedBy: values["trace_id"],
			EndTime:  time.Unix(endTimeUnix, 0),
			Status:   consts.GetStatusTypeName(status),
		}
	}

	resp := &dto.ListNamespaceLockResp{
		Items: nsMap,
	}
	return resp, nil
}

// ListQueuedTasks lists tasks currently in the ready and delayed queues
func ListQueuedTasks(ctx context.Context) (*dto.QueuedTasksResp, error) {
	readyTaskDatas, err := repository.ListReadyTasks(ctx)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("%w: no ready tasks found", consts.ErrNotFound)
		}
		return nil, err
	}

	readyTask := make([]dto.TaskResp, 0, len(readyTaskDatas))
	for _, taskData := range readyTaskDatas {
		var task database.Task
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			return nil, err
		}

		readyTask = append(readyTask, *dto.NewTaskResp(&task))
	}

	delayedTaskDatas, err := repository.ListDelayedTasks(ctx, 1000)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("%w: no delayed tasks found", consts.ErrNotFound)
		}
		return nil, err
	}

	delayedTask := make([]dto.TaskResp, 0, len(delayedTaskDatas))
	for _, taskData := range delayedTaskDatas {
		var task database.Task
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			return nil, err
		}

		delayedTask = append(delayedTask, *dto.NewTaskResp(&task))
	}

	resp := &dto.QueuedTasksResp{
		ReadyTasks:   readyTask,
		DelayedTasks: delayedTask,
	}
	return resp, nil
}

// GetSystemMetrics retrieves current system metrics
func GetSystemMetrics(ctx context.Context) (*dto.SystemMetricsResp, error) {
	now := time.Now()

	// Get CPU usage
	cpuPercent, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %v", err)
	}
	cpuUsage := 0.0
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	// Get memory usage
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory usage: %v", err)
	}

	// Get disk usage
	diskInfo, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %v", err)
	}

	resp := &dto.SystemMetricsResp{
		CPU: dto.MetricValue{
			Value:     cpuUsage,
			Timestamp: now,
			Unit:      "%",
		},
		Memory: dto.MetricValue{
			Value:     memInfo.UsedPercent,
			Timestamp: now,
			Unit:      "%",
		},
		Disk: dto.MetricValue{
			Value:     diskInfo.UsedPercent,
			Timestamp: now,
			Unit:      "%",
		},
	}

	return resp, nil
}

// GetSystemMetricsHistory retrieves historical system metrics (24 hours)
func GetSystemMetricsHistory(ctx context.Context) (*dto.SystemMetricsHistoryResp, error) {
	redisClient := client.GetRedisClient()
	now := time.Now()

	// Get last 24 hours of metrics from Redis
	startTime := now.Add(-24 * time.Hour).Unix()
	endTime := now.Unix()

	cpuKey := "system:metrics:cpu"
	memKey := "system:metrics:memory"

	// Get CPU history
	cpuData, err := redisClient.ZRangeByScore(ctx, cpuKey, &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", startTime),
		Max: fmt.Sprintf("%d", endTime),
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("failed to get CPU history: %v", err)
	}

	// Get memory history
	memData, err := redisClient.ZRangeByScore(ctx, memKey, &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", startTime),
		Max: fmt.Sprintf("%d", endTime),
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("failed to get memory history: %v", err)
	}

	// Parse CPU data
	cpuMetrics := make([]dto.MetricValue, 0, len(cpuData))
	for _, data := range cpuData {
		var metric dto.MetricValue
		if err := json.Unmarshal([]byte(data), &metric); err == nil {
			cpuMetrics = append(cpuMetrics, metric)
		}
	}

	// Parse memory data
	memMetrics := make([]dto.MetricValue, 0, len(memData))
	for _, data := range memData {
		var metric dto.MetricValue
		if err := json.Unmarshal([]byte(data), &metric); err == nil {
			memMetrics = append(memMetrics, metric)
		}
	}

	// If no historical data, generate current metrics
	if len(cpuMetrics) == 0 || len(memMetrics) == 0 {
		current, err := GetSystemMetrics(ctx)
		if err != nil {
			return nil, err
		}

		if len(cpuMetrics) == 0 {
			cpuMetrics = []dto.MetricValue{current.CPU}
		}
		if len(memMetrics) == 0 {
			memMetrics = []dto.MetricValue{current.Memory}
		}
	}

	resp := &dto.SystemMetricsHistoryResp{
		CPU:    cpuMetrics,
		Memory: memMetrics,
	}

	return resp, nil
}

// StoreSystemMetrics stores current system metrics in Redis for historical tracking
func StoreSystemMetrics(ctx context.Context) error {
	metrics, err := GetSystemMetrics(ctx)
	if err != nil {
		return err
	}

	redisClient := client.GetRedisClient()
	now := time.Now().Unix()

	// Store CPU metric
	cpuData, _ := json.Marshal(metrics.CPU)
	if err := redisClient.ZAdd(ctx, "system:metrics:cpu", redis.Z{
		Score:  float64(now),
		Member: cpuData,
	}).Err(); err != nil {
		return fmt.Errorf("failed to store CPU metric: %v", err)
	}

	// Store memory metric
	memData, _ := json.Marshal(metrics.Memory)
	if err := redisClient.ZAdd(ctx, "system:metrics:memory", redis.Z{
		Score:  float64(now),
		Member: memData,
	}).Err(); err != nil {
		return fmt.Errorf("failed to store memory metric: %v", err)
	}

	// Clean up old metrics (older than 24 hours)
	oldTime := time.Now().Add(-24 * time.Hour).Unix()
	redisClient.ZRemRangeByScore(ctx, "system:metrics:cpu", "0", fmt.Sprintf("%d", oldTime))
	redisClient.ZRemRangeByScore(ctx, "system:metrics:memory", "0", fmt.Sprintf("%d", oldTime))

	return nil
}

func init() {
	// Start background goroutine to collect metrics every minute
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			ctx := context.Background()
			if err := StoreSystemMetrics(ctx); err != nil {
				// Log error but don't crash
				runtime.Gosched()
			}
		}
	}()
}
