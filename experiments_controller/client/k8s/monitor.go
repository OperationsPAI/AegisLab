package k8s

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/config"
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

func GetNS2Monitor() ([]string, error) {
	m := config.GetMap("injection.namespace_target_map")
	namespaces := make([]string, 0)

	for ns, value := range m {

		vInt, ok := value.(int64)
		if !ok {
			return nil, fmt.Errorf("invalid namespace value for %s", ns)
		}

		for idx := range vInt {
			namespaces = append(namespaces, fmt.Sprintf("%s%d", ns, idx))
		}
	}
	return namespaces, nil
}

// initMonitor creates and initializes a new Monitor instance
func initMonitor() *Monitor {
	initialNamespaces, err := GetNS2Monitor()
	if err != nil {
		logrus.Fatalf("Failed to get namespaces for initialization: %v", err)
	}
	redisClient := client.GetRedisClient()
	ctx := context.Background()

	// Add namespaces to Redis set
	if len(initialNamespaces) > 0 {
		members := make([]interface{}, len(initialNamespaces))
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
		redisClient.HSetNX(ctx, nsKey, "status", "true")
		redisClient.HSetNX(ctx, nsKey, "trace_id", "")
	}

	return &Monitor{
		redisClient: redisClient,
		ctx:         ctx,
	}
}

func (m *Monitor) checkNamespaceToInject(namespace string, executeTime time.Time, traceID string) error {
	nsKey := fmt.Sprintf(namespaceKeyPattern, namespace)

	// Get namespace data from Redis
	values, err := m.redisClient.HGetAll(m.ctx, nsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get namespace data from Redis: %v", err)
	}

	if len(values) == 0 {
		return fmt.Errorf("failed to find the item of the namespace %s", namespace)
	}

	// Check status
	if values["status"] != "true" {
		return fmt.Errorf("the service in namespace %s is not yet fully ready for fault injection", namespace)
	}

	// Check end time
	endTimeUnix, err := strconv.ParseInt(values["end_time"], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid end_time format in Redis: %v", err)
	}

	endTime := time.Unix(endTimeUnix, 0)
	if !endTime.After(executeTime) {
		return fmt.Errorf("cannot inject fault: namespace %s is locked until %v (current execution time: %v)",
			namespace, endTime.Format(time.RFC3339), executeTime.Format(time.RFC3339))
	}

	// Check trace ID
	if values["trace_id"] != "" && values["trace_id"] != traceID {
		return fmt.Errorf("namespace %s is currently locked by another trace (trace_id: %s)", namespace, values["trace_id"])
	}

	return nil
}

func (m *Monitor) getNamespaceToRestart(endTime time.Time, traceID string) string {
	// Get all namespaces
	namespaces, err := m.redisClient.SMembers(m.ctx, namespacesKey).Result()
	if err != nil {
		logrus.Errorf("Failed to get namespaces from Redis: %v", err)
		return ""
	}

	nowTime := time.Now().Unix()

	// Use a Redis transaction to atomically check and update namespace locks
	for _, ns := range namespaces {
		nsKey := fmt.Sprintf(namespaceKeyPattern, ns)

		// Get the end time
		endTimeStr, err := m.redisClient.HGet(m.ctx, nsKey, "end_time").Result()
		if err != nil {
			logrus.WithField("namespace", ns).Errorf("Failed to get end_time: %v", err)
			continue
		}

		nsEndTime, err := strconv.ParseInt(endTimeStr, 10, 64)
		if err != nil {
			logrus.WithField("namespace", ns).Errorf("Invalid end_time format: %v", err)
			continue
		}

		// If the lock has expired
		if nsEndTime < nowTime {
			// Try to acquire the lock using WATCH/MULTI/EXEC for atomicity
			err := m.redisClient.Watch(m.ctx, func(tx *redis.Tx) error {
				// Check if the lock is still available
				currentEndTime, err := tx.HGet(m.ctx, nsKey, "end_time").Int64()
				if err != nil {
					return err
				}

				if nowTime < currentEndTime {
					return redis.TxFailedErr // Lock was taken by someone else
				}

				// Try to acquire the lock
				_, err = tx.TxPipelined(m.ctx, func(pipe redis.Pipeliner) error {
					pipe.HSet(m.ctx, nsKey, "end_time", endTime.Unix())
					pipe.HSet(m.ctx, nsKey, "trace_id", traceID)
					return nil
				})
				return err
			}, nsKey)

			if err == nil {
				logrus.WithFields(
					logrus.Fields{
						"namespace": ns,
						"trace_id":  traceID,
					},
				).Info("acquire namespace lock")
				return ns
			}
		}
	}

	return ""
}

func (m *Monitor) setTime(namespace string, endTime time.Time, traceID string) {
	nsKey := fmt.Sprintf(namespaceKeyPattern, namespace)

	// Check if namespace exists
	exists, err := m.redisClient.Exists(m.ctx, nsKey).Result()
	if err != nil {
		logrus.WithField("namespace", namespace).Errorf("Failed to check namespace existence: %v", err)
		return
	}

	if exists == 0 {
		logrus.WithField("namespace", namespace).Warn("Namespace not found in Redis")
		return
	}

	// Update namespace lock info
	_, err = m.redisClient.Pipelined(m.ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(m.ctx, nsKey, "end_time", endTime.Unix())
		pipe.HSet(m.ctx, nsKey, "trace_id", traceID)
		return nil
	})

	if err != nil {
		logrus.WithField("namespace", namespace).Errorf("Failed to update namespace lock: %v", err)
	} else {
		logrus.WithField("namespace", namespace).Info("release namespace lock")
	}
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
