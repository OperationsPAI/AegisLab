package k8s

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/repository"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Redis keys
const (
	// Set of all monitored namespaces
	namespacesKey = "monitor:namespaces"
	// Hash pattern for namespace status (will be monitor:ns:{namespace})
	namespaceKeyPattern = "monitor:ns:%s"
	// Used to coordinate initialization across processes
	monitorInitKey = "monitor:initialized"
)

// MonitorItem represents the state of a namespace lock
type MonitorItem struct {
	EndTime time.Time `json:"end_time"`
	TraceID string    `json:"trace_id"`
}

// Monitor manages namespace locks and status using Redis
type Monitor struct {
	redisClient *redis.Client
	ctx         context.Context
}

// Singleton instance and initialization control
var (
	monitorInstance *Monitor
	monitorOnce     sync.Once
)

// GetMonitor returns the singleton Monitor instance,
// ensuring initialization is only performed once across all processes
func GetMonitor() *Monitor {
	// Local process singleton pattern
	monitorOnce.Do(func() {
		redisClient := client.GetRedisClient()
		ctx := context.Background()

		// Try to set the initialization flag in Redis
		// If we're the first to set this, we'll get true and handle initialization
		// If we're not the first, we'll get false and skip initialization
		isFirstInit, err := redisClient.SetNX(ctx, monitorInitKey, time.Now().String(), 0).Result()
		if err != nil {
			logrus.Warnf("Failed to check initialization status: %v, proceeding with local init", err)
			isFirstInit = true // Default to initializing if we can't check
		}

		if isFirstInit {
			logrus.Info("First process initializing the Monitor")
			// We're the first process to initialize Monitor
			monitorInstance = initMonitor()
		} else {
			logrus.Info("Monitor already initialized by another process")
			// Another process already initialized, just create the local instance
			monitorInstance = &Monitor{
				redisClient: redisClient,
				ctx:         ctx,
			}

		}
	})

	return monitorInstance
}

// initMonitor creates and initializes a new Monitor instance
func initMonitor() *Monitor {
	initialNamespaces, err := config.GetAllNamespaces()
	if err != nil {
		logrus.Fatalf("Failed to get namespaces for initialization: %v", err)
	}

	redisClient := client.GetRedisClient()
	ctx := context.Background()

	// Add namespaces to Redis set
	if len(initialNamespaces) > 0 {
		members := make([]any, len(initialNamespaces))
		for i, ns := range initialNamespaces {
			members[i] = ns
		}
		redisClient.SAdd(ctx, namespacesKey, members...)
	}

	// Initialize namespace data in Redis
	now := time.Now().Unix()
	for _, namespace := range initialNamespaces {
		nsKey := fmt.Sprintf(namespaceKeyPattern, namespace)

		redisClient.HSetNX(ctx, nsKey, "end_time", now)
		redisClient.HSetNX(ctx, nsKey, "trace_id", "")
	}

	return &Monitor{
		redisClient: redisClient,
		ctx:         ctx,
	}
}

// acquireNamespaceLock attempts to acquire a lock on a namespace
// Returns nil on success, error if the lock cannot be acquired
func (m *Monitor) acquireNamespaceLock(namespace string, endTime time.Time, traceID string) (err error) {

	defer func() {
		repository.PublishEvent(context.Background(), fmt.Sprintf(consts.StreamLogKey, namespace), dto.StreamEvent{
			TaskType:  consts.TaskTypeRestartService,
			EventName: consts.EventAcquireLock,
			Payload:   map[string]any{"trace_id": traceID, "end_time": endTime, "error": err},
		})
	}()

	nsKey := fmt.Sprintf(namespaceKeyPattern, namespace)
	nowTime := time.Now().Unix()

	err = m.redisClient.Watch(m.ctx, func(tx *redis.Tx) error {
		// Check if the lock is still available
		currentEndTimeStr, e := tx.HGet(m.ctx, nsKey, "end_time").Result()
		if e != nil && e != redis.Nil {
			return e
		}

		currentEndTime, e := strconv.ParseInt(currentEndTimeStr, 10, 64)
		if e != nil {
			return e
		}

		currentTraceID, e := tx.HGet(m.ctx, nsKey, "trace_id").Result()
		if e != nil && e != redis.Nil {
			return e
		}

		// If lock is held by someone else and not expired
		if currentTraceID != "" && currentTraceID != traceID && nowTime < currentEndTime {
			return redis.TxFailedErr // Lock was taken by someone else
		}

		// Try to acquire the lock
		_, e = tx.TxPipelined(m.ctx, func(pipe redis.Pipeliner) error {
			pipe.HSet(m.ctx, nsKey, "end_time", endTime.Unix())
			pipe.HSet(m.ctx, nsKey, "trace_id", traceID)
			return nil
		})
		return e
	}, nsKey)

	return err
}

// releaseNamespaceLock releases a lock on a namespace if it's owned by the specified traceID
func (m *Monitor) releaseNamespaceLock(namespace string, traceID string) (err error) {

	defer func() {
		repository.PublishEvent(context.Background(), fmt.Sprintf(consts.StreamLogKey, namespace), dto.StreamEvent{
			TaskType:  consts.TaskTypeRestartService,
			EventName: consts.EventReleaseLock,
			Payload:   map[string]any{"trace_id": traceID, "error": err},
		})
	}()

	nsKey := fmt.Sprintf(namespaceKeyPattern, namespace)

	// Check if namespace exists
	var exists int64
	exists, err = m.redisClient.Exists(m.ctx, nsKey).Result()
	if err != nil {
		err = fmt.Errorf("failed to check namespace existence: %v", err)
		return
	}

	if exists == 0 {
		err = fmt.Errorf("namespace %s not found", namespace)
		return
	}

	// Check if the lock is actually held by this traceID
	currentTraceID, err := m.redisClient.HGet(m.ctx, nsKey, "trace_id").Result()
	if err != nil && err != redis.Nil {
		err = fmt.Errorf("failed to get current trace_id: %v", err)
		return
	}

	// If the lock is held by someone else or is already released
	if currentTraceID != traceID {
		err = fmt.Errorf("cannot release lock: namespace %s is not owned by trace_id %s (current owner: %s)",
			namespace, traceID, currentTraceID)
		return
	}

	// Update namespace lock info - release by setting current time and empty trace ID
	_, err = m.redisClient.Pipelined(m.ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(m.ctx, nsKey, "end_time", time.Now().Unix())
		pipe.HSet(m.ctx, nsKey, "trace_id", "")
		return nil
	})

	return
}

// Update the ReleaseLock method to pass traceID
func (m *Monitor) ReleaseLock(namespace string, traceID string) {
	if namespace == "" || traceID == "" {
		logrus.WithFields(logrus.Fields{
			"namespace": namespace,
			"trace_id":  traceID,
		}).Error("namespace or trace_id is empty")
		return
	}
	err := m.releaseNamespaceLock(namespace, traceID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"namespace": namespace,
			"trace_id":  traceID,
		}).Errorf("Failed to release namespace lock: %v", err)
	} else {
		logrus.WithFields(logrus.Fields{
			"namespace": namespace,
			"trace_id":  traceID,
		}).Info("released namespace lock")
	}
}

func (m *Monitor) AcquireLock(endTime time.Time, traceID string) string {
	return m.getNamespaceToRestart(endTime, traceID)
}

func (m *Monitor) InspectLock() (map[string]*MonitorItem, error) {
	// Get all namespaces
	namespaces, err := m.redisClient.SMembers(m.ctx, namespacesKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces from Redis: %v", err)
	}

	nsMap := make(map[string]*MonitorItem, len(namespaces))

	// Get data for each namespace
	for _, ns := range namespaces {
		nsKey := fmt.Sprintf(namespaceKeyPattern, ns)
		values, err := m.redisClient.HGetAll(m.ctx, nsKey).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get data for namespace %s: %v", ns, err)
		}

		endTimeUnix, err := strconv.ParseInt(values["end_time"], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time format for namespace %s: %v", ns, err)
		}

		nsMap[ns] = &MonitorItem{
			EndTime: time.Unix(endTimeUnix, 0),
			TraceID: values["trace_id"],
		}
	}

	return nsMap, nil
}

func (m *Monitor) CheckNamespaceToInject(namespace string, executeTime time.Time, traceID string) error {
	nsKey := fmt.Sprintf(namespaceKeyPattern, namespace)

	// Get namespace data from Redis
	values, err := m.redisClient.HGetAll(m.ctx, nsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get namespace data from Redis: %v", err)
	}

	if len(values) == 0 {
		return fmt.Errorf("failed to find the item of the namespace %s", namespace)
	}

	// Check end time
	endTimeUnix, err := strconv.ParseInt(values["end_time"], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid end_time format in Redis: %v", err)
	}
	endTime := time.Unix(endTimeUnix, 0)

	// Check if the lock is held by current client
	isOwnLock := values["trace_id"] == traceID

	// Check if the lock is expired (regardless of who owns it)
	isLockExpired := time.Now().After(endTime)

	// If the lock is held by another client and hasn't expired yet
	if !isOwnLock && !isLockExpired && values["trace_id"] != "" {
		return fmt.Errorf("cannot inject fault: namespace %s is occupied by %v at %v (current execution time: %v)",
			namespace, values["trace_id"], endTime.Format(time.RFC3339), executeTime.Format(time.RFC3339))
	}

	// At this point, either:
	// 1. We own the lock already
	// 2. The lock is expired
	// 3. The lock is not assigned to anyone
	// In all cases, we should be able to acquire/refresh the lock

	// Refresh or acquire the lock
	proposedEndTime := executeTime.Add(time.Duration(5) * time.Minute)

	// Use the extracted helper function
	err = m.acquireNamespaceLock(namespace, proposedEndTime, traceID)

	if err != nil {
		if err == redis.TxFailedErr {
			return fmt.Errorf("cannot inject fault: namespace %s was concurrently acquired by another client", namespace)
		}
		return fmt.Errorf("failed to acquire lock: %v", err)
	}

	logrus.WithFields(
		logrus.Fields{
			"namespace": namespace,
			"trace_id":  traceID,
			"end_time":  proposedEndTime,
		},
	).Info("refreshed or acquired namespace lock")

	return nil
}

func (m *Monitor) getNamespaceToRestart(endTime time.Time, traceID string) string {
	// Get all namespaces
	namespaces, err := m.redisClient.SMembers(m.ctx, namespacesKey).Result()
	if err != nil {
		logrus.Errorf("Failed to get namespaces from Redis: %v", err)
		return ""
	}

	// Try to acquire an available namespace
	for _, ns := range namespaces {
		nsKey := fmt.Sprintf(namespaceKeyPattern, ns)

		// 直接尝试获取锁，在锁内部进行过期检查
		err := m.redisClient.Watch(m.ctx, func(tx *redis.Tx) error {
			// 获取当前end_time和trace_id
			endTimeStr, err := tx.HGet(m.ctx, nsKey, "end_time").Result()
			if err != nil {
				return err
			}

			nsEndTime, err := strconv.ParseInt(endTimeStr, 10, 64)
			if err != nil {
				return err
			}

			currentTraceID, err := tx.HGet(m.ctx, nsKey, "trace_id").Result()
			if err != nil && err != redis.Nil {
				return err
			}

			// 检查锁是否过期或空闲
			nowTime := time.Now().Unix()
			if nsEndTime >= nowTime && currentTraceID != "" {
				return fmt.Errorf("namespace lock not available")
			}

			// 尝试获取锁
			_, err = tx.TxPipelined(m.ctx, func(pipe redis.Pipeliner) error {
				pipe.HSet(m.ctx, nsKey, "end_time", endTime.Unix())
				pipe.HSet(m.ctx, nsKey, "trace_id", traceID)
				return nil
			})
			return err
		}, nsKey)

		if err == nil {
			logrus.WithFields(logrus.Fields{
				"namespace": ns,
				"trace_id":  traceID,
			}).Info("acquired namespace lock")
			return ns
		} else if err != redis.TxFailedErr {
			logrus.WithFields(logrus.Fields{
				"namespace": ns,
				"error":     err,
				"trace_id":  traceID,
			}).Warn("failed to acquire namespace lock")
		}
	}

	return ""
}
