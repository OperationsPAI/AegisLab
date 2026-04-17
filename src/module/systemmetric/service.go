package systemmetricmodule

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"aegis/consts"
	redisinfra "aegis/infra/redis"
	"aegis/model"
	taskmodule "aegis/module/task"

	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type Service struct {
	repo  *Repository
	redis *redisinfra.Gateway
}

func NewService(repo *Repository, redis *redisinfra.Gateway) *Service {
	return &Service{repo: repo, redis: redis}
}

func (s *Service) GetSystemMetrics(ctx context.Context) (*SystemMetricsResp, error) {
	now := time.Now()

	cpuPercent, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %v", err)
	}
	cpuUsage := 0.0
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory usage: %v", err)
	}

	diskInfo, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %v", err)
	}

	return &SystemMetricsResp{
		CPU: MetricValue{
			Value:     cpuUsage,
			Timestamp: now,
			Unit:      "%",
		},
		Memory: MetricValue{
			Value:     memInfo.UsedPercent,
			Timestamp: now,
			Unit:      "%",
		},
		Disk: MetricValue{
			Value:     diskInfo.UsedPercent,
			Timestamp: now,
			Unit:      "%",
		},
	}, nil
}

func (s *Service) GetSystemMetricsHistory(ctx context.Context) (*SystemMetricsHistoryResp, error) {
	now := time.Now()
	startTime := now.Add(-24 * time.Hour).Unix()
	endTime := now.Unix()

	cpuData, err := s.redis.ZRangeByScore(ctx, "system:metrics:cpu", fmt.Sprintf("%d", startTime), fmt.Sprintf("%d", endTime))
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("failed to get CPU history: %v", err)
	}

	memData, err := s.redis.ZRangeByScore(ctx, "system:metrics:memory", fmt.Sprintf("%d", startTime), fmt.Sprintf("%d", endTime))
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("failed to get memory history: %v", err)
	}

	cpuMetrics := parseMetricValues(cpuData)
	memMetrics := parseMetricValues(memData)

	if len(cpuMetrics) == 0 || len(memMetrics) == 0 {
		current, err := s.GetSystemMetrics(ctx)
		if err != nil {
			return nil, err
		}
		if len(cpuMetrics) == 0 {
			cpuMetrics = []MetricValue{current.CPU}
		}
		if len(memMetrics) == 0 {
			memMetrics = []MetricValue{current.Memory}
		}
	}

	return &SystemMetricsHistoryResp{
		CPU:    cpuMetrics,
		Memory: memMetrics,
	}, nil
}

func (s *Service) ListNamespaceLocks(ctx context.Context) (*ListNamespaceLockResp, error) {
	namespaces, err := s.redis.SetMembers(ctx, consts.NamespacesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces from Redis: %v", err)
	}

	items := make(map[string]NsMonitorItem, len(namespaces))
	for _, namespace := range namespaces {
		nsKey := fmt.Sprintf(consts.NamespaceKeyPattern, namespace)
		values, err := s.redis.HashGetAll(ctx, nsKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get data for namespace %s: %v", namespace, err)
		}

		endTimeUnix, err := strconv.ParseInt(values["end_time"], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time format for namespace %s: %v", namespace, err)
		}

		status := consts.CommonEnabled
		if statusStr, ok := values["status"]; ok {
			statusInt, err := strconv.Atoi(statusStr)
			if err == nil {
				status = consts.StatusType(statusInt)
			}
		}

		items[namespace] = NsMonitorItem{
			LockedBy: values["trace_id"],
			EndTime:  time.Unix(endTimeUnix, 0),
			Status:   consts.GetStatusTypeName(status),
		}
	}

	return &ListNamespaceLockResp{Items: items}, nil
}

func (s *Service) ListQueuedTasks(ctx context.Context) (*taskmodule.QueuedTasksResp, error) {
	readyTaskDatas, err := s.redis.ListReadyTasks(ctx)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("%w: no ready tasks found", consts.ErrNotFound)
		}
		return nil, err
	}

	readyTasks := make([]taskmodule.TaskResp, 0, len(readyTaskDatas))
	for _, taskData := range readyTaskDatas {
		var task model.Task
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			return nil, err
		}
		readyTasks = append(readyTasks, *taskmodule.NewTaskResp(&task))
	}

	delayedTaskDatas, err := s.redis.ListDelayedTasks(ctx, 1000)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("%w: no delayed tasks found", consts.ErrNotFound)
		}
		return nil, err
	}

	delayedTasks := make([]taskmodule.TaskResp, 0, len(delayedTaskDatas))
	for _, taskData := range delayedTaskDatas {
		var task model.Task
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			return nil, err
		}
		delayedTasks = append(delayedTasks, *taskmodule.NewTaskResp(&task))
	}

	return &taskmodule.QueuedTasksResp{
		ReadyTasks:   readyTasks,
		DelayedTasks: delayedTasks,
	}, nil
}

func parseMetricValues(items []string) []MetricValue {
	metrics := make([]MetricValue, 0, len(items))
	for _, item := range items {
		var metric MetricValue
		if err := json.Unmarshal([]byte(item), &metric); err == nil {
			metrics = append(metrics, metric)
		}
	}
	return metrics
}

func (s *Service) StoreSystemMetrics(ctx context.Context) error {
	metrics, err := s.GetSystemMetrics(ctx)
	if err != nil {
		return err
	}

	now := time.Now().Unix()

	cpuData, _ := json.Marshal(metrics.CPU)
	if err := s.redis.ZAdd(ctx, "system:metrics:cpu", redis.Z{
		Score:  float64(now),
		Member: cpuData,
	}); err != nil {
		return fmt.Errorf("failed to store CPU metric: %v", err)
	}

	memData, _ := json.Marshal(metrics.Memory)
	if err := s.redis.ZAdd(ctx, "system:metrics:memory", redis.Z{
		Score:  float64(now),
		Member: memData,
	}); err != nil {
		return fmt.Errorf("failed to store memory metric: %v", err)
	}

	oldTime := time.Now().Add(-24 * time.Hour).Unix()
	_ = s.redis.ZRemRangeByScore(ctx, "system:metrics:cpu", "0", fmt.Sprintf("%d", oldTime))
	_ = s.redis.ZRemRangeByScore(ctx, "system:metrics:memory", "0", fmt.Sprintf("%d", oldTime))

	return nil
}
