package consumer

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"sync"
	"time"

	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	redisinfra "aegis/infra/redis"
	"aegis/utils"

	goredis "github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type LockMessage struct {
	TraceID string    `json:"trace_id"`
	EndTime time.Time `json:"end_time,omitempty"`
	Error   error     `json:"err"`
}

// MonitorItem represents the state of a namespace lock
type MonitorItem struct {
	EndTime time.Time `json:"end_time"`
	TraceID string    `json:"trace_id"`
}

// NamespaceRefreshResult contains detailed results of namespace refresh operation
type NamespaceRefreshResult struct {
	Added     []string // Newly added namespaces (new in config)
	Recovered []string // Namespaces that were disabled/deleted but now enabled again
	Disabled  []string // Namespaces removed from config but have active locks
	Deleted   []string // Namespaces removed from config with no active locks
}

// NamespaceInitResult contains results of namespace initialization on startup
type NamespaceInitResult struct {
	Refreshed   *NamespaceRefreshResult // Result of configuration refresh
	Initialized []string                // Namespaces that were re-initialized (all enabled namespaces)
}

type NamespaceMonitor interface {
	SetContext(ctx context.Context)
	InitializeNamespaces() ([]string, error)
	RefreshNamespaces() (*NamespaceRefreshResult, error)
	ReleaseLock(ctx context.Context, namespace string, traceID string) error
	CheckNamespaceToInject(namespace string, executeTime time.Time, traceID string) error
	GetNamespaceToRestart(endTime time.Time, nsPattern, traceID string) string
}

// monitor manages namespace locks and status using Redis
type monitor struct {
	ctx          context.Context
	redisGateway *redisinfra.Gateway
	namespaces   namespaceCatalogStore
	locks        namespaceLockStore
	status       namespaceStatusStore
	mu           sync.RWMutex // Protects namespace operations
}

func NewMonitor(gateway *redisinfra.Gateway) NamespaceMonitor {
	return &monitor{
		ctx:          context.TODO(),
		redisGateway: gateway,
		namespaces:   newNamespaceCatalogStore(gateway),
		locks:        newNamespaceLockStore(gateway),
		status:       newNamespaceStatusStore(gateway),
	}
}

func (m *monitor) SetContext(ctx context.Context) {
	if ctx == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ctx = ctx
}

func (m *monitor) currentContext() context.Context {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.ctx != nil {
		return m.ctx
	}
	return context.TODO()
}

func (m *monitor) listNamespaces() ([]string, error) {
	return m.namespaces.list(m.currentContext())
}

func (m *monitor) namespaceExists(namespace string) (bool, error) {
	return m.namespaces.exists(m.currentContext(), namespace)
}

func (m *monitor) seedNamespace(namespace string, endTime time.Time) error {
	return m.namespaces.seed(m.currentContext(), namespace, endTime)
}

// AcquireLock attempts to acquire a lock on a namespace
// Returns nil on success, error if the lock cannot be acquired
func (m *monitor) AcquireLock(namespace string, endTime time.Time, traceID string, taskType consts.TaskType) (err error) {
	defer func() {
		publishEvent(m.redisGateway, m.currentContext(), fmt.Sprintf(consts.StreamTraceLogKey, namespace), dto.TraceStreamEvent{
			TaskType:  taskType,
			EventName: consts.EventAcquireLock,
			Payload: LockMessage{
				TraceID: traceID,
				EndTime: endTime,
				Error:   err,
			},
		})
	}()

	nowTime := time.Now().Unix()

	// Check if namespace exists
	exists, err := m.namespaceExists(namespace)
	if err != nil {
		return fmt.Errorf("failed to check namespace existence: %v", err)
	}

	if !exists {
		// Lazy loading: verify namespace is valid in current configuration
		latestNamespaces, err := config.GetAllNamespaces()
		if err != nil {
			return fmt.Errorf("failed to validate namespace: %w", err)
		}

		isValid := slices.Contains(latestNamespaces, namespace)
		if !isValid {
			return fmt.Errorf("namespace %s not found in current configuration", namespace)
		}

		// Namespace is valid but not in Redis, auto-add it
		logrus.Infof("Lazy-loading namespace: %s", namespace)
		if err := m.addNamespace(namespace, time.Now()); err != nil {
			return fmt.Errorf("failed to lazy-load namespace: %w", err)
		}
	}

	// Check namespace status (reject if disabled or deleted)
	status, err := m.getNamespaceStatus(namespace)
	if err != nil {
		return fmt.Errorf("failed to check namespace status: %v", err)
	}
	if status == consts.CommonDisabled {
		return fmt.Errorf("namespace %s is disabled and not accepting new locks", namespace)
	}
	if status == consts.CommonDeleted {
		return fmt.Errorf("namespace %s has been deleted", namespace)
	}

	// All lock checking and acquisition happens in a single atomic transaction
	err = m.locks.acquire(m.currentContext(), namespace, endTime, traceID, time.Unix(nowTime, 0))

	logEntry := logrus.WithFields(
		logrus.Fields{
			"namespace": namespace,
			"trace_id":  traceID,
			"end_time":  endTime,
		},
	)

	if err == nil {
		logEntry.Info("acquired namespace lock")
	} else if err != goredis.TxFailedErr {
		logEntry.Warn("failed to acquire namespace lock")
	}

	return err
}

// ReleaseLock releases a lock on a namespace if it's owned by the specified traceID
func (m *monitor) ReleaseLock(ctx context.Context, namespace string, traceID string) (err error) {
	defer func() {
		publishEvent(m.redisGateway, ctx, fmt.Sprintf(consts.StreamTraceLogKey, namespace), dto.TraceStreamEvent{
			TaskType:  consts.TaskTypeRestartPedestal,
			EventName: consts.EventReleaseLock,
			Payload: LockMessage{
				TraceID: traceID,
				Error:   err,
			},
		})
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
	}()

	if namespace == "" || traceID == "" {
		return fmt.Errorf("namespace or trace_id is empty")
	}

	// Check if namespace exists
	exists, existsErr := m.namespaceExists(namespace)
	err = existsErr
	if err != nil {
		err = fmt.Errorf("failed to check namespace existence: %v", err)
		return
	}

	if !exists {
		err = fmt.Errorf("namespace %s not found", namespace)
		return
	}

	// Check if the lock is actually held by this traceID
	err = m.locks.release(m.currentContext(), namespace, traceID, time.Now())

	return
}

// CheckNamespaceToInject checks if a specific namespace is available for injection and acquires it
func (m *monitor) CheckNamespaceToInject(namespace string, executeTime time.Time, traceID string) error {
	// Calculate proposed end time for the lock (5 minutes after execution time)
	proposedEndTime := executeTime.Add(time.Duration(5) * time.Minute)

	// Try to acquire the lock - all availability checking is done inside acquireNamespaceLock
	err := m.AcquireLock(namespace, proposedEndTime, traceID, consts.TaskTypeFaultInjection)
	if err != nil {
		if err == goredis.TxFailedErr {
			return fmt.Errorf("cannot inject fault: namespace %s was concurrently acquired by another client", namespace)
		}
		return fmt.Errorf("cannot inject fault: %v", err)
	}

	return nil
}

// GetNamespaceToRestart finds an available namespace for restart and acquires it
func (m *monitor) GetNamespaceToRestart(endTime time.Time, nsPattern, traceID string) string {
	namespaces, err := m.listNamespaces()
	if err != nil {
		logrus.Errorf("failed to get namespaces from Redis: %v", err)
		return ""
	}

	// Compile the pattern as regex
	var pattern *regexp.Regexp
	if nsPattern != "" {
		pattern, err = regexp.Compile(nsPattern)
		if err != nil {
			logrus.Errorf("failed to compile namespace pattern '%s': %v", nsPattern, err)
			return ""
		}
	}

	for _, ns := range namespaces {
		// Check namespace status - only allocate enabled namespaces
		status, err := m.getNamespaceStatus(ns)
		if err != nil {
			logrus.Errorf("Failed to get status for namespace %s: %v", ns, err)
			continue
		}

		if status != consts.CommonEnabled {
			logrus.Debugf("Skipping namespace %s (status: %s)", ns, consts.GetStatusTypeName(status))
			continue
		}

		// Match namespace against pattern
		if pattern != nil && pattern.MatchString(ns) {
			if err := m.AcquireLock(ns, endTime, traceID, consts.TaskTypeRestartPedestal); err == nil {
				return ns
			}
		}
	}

	return ""
}

// InitializeNamespaces should be called on program startup to ensure all enabled namespaces
// are properly initialized, even if the program was restarted
func (m *monitor) InitializeNamespaces() ([]string, error) {
	_, err := m.RefreshNamespaces()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh namespaces during initialization: %w", err)
	}

	// Get all enabled namespaces from Redis
	allNamespaces, err := m.listNamespaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces from Redis: %w", err)
	}

	initialized := make([]string, 0)
	for _, ns := range allNamespaces {
		status, err := m.getNamespaceStatus(ns)
		if err != nil {
			logrus.Errorf("Failed to get status for namespace %s: %v", ns, err)
			continue
		}

		if status == consts.CommonEnabled {
			if err := m.addNamespace(ns, time.Now()); err != nil {
				logrus.Errorf("Failed to initialize namespace %s: %v", ns, err)
			} else {
				initialized = append(initialized, ns)
				logrus.Debugf("Initialized namespace on startup: %s", ns)
			}
		}
	}

	return initialized, nil
}

// RefreshNamespaces updates the namespace list based on current configuration
// Returns detailed results of namespace state changes
func (m *monitor) RefreshNamespaces() (*NamespaceRefreshResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := &NamespaceRefreshResult{
		Added:     make([]string, 0),
		Recovered: make([]string, 0),
		Disabled:  make([]string, 0),
		Deleted:   make([]string, 0),
	}

	// Get latest namespaces from configuration
	latestNamespaces, err := config.GetAllNamespaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest namespaces: %w", err)
	}

	// Get existing namespaces from Redis
	existingNamespaces, err := m.listNamespaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get existing namespaces: %w", err)
	}

	latestSet := utils.MakeSet(latestNamespaces)
	existingSet := utils.MakeSet(existingNamespaces)

	// Handle namespaces in latest config
	for ns := range latestSet {
		if _, exists := existingSet[ns]; !exists {
			// Brand new namespace, add it
			if err := m.addNamespace(ns, time.Now()); err != nil {
				logrus.Errorf("Failed to add namespace %s: %v", ns, err)
			} else {
				result.Added = append(result.Added, ns)
				logrus.Infof("Added new namespace: %s", ns)
			}
		} else {
			// Existing namespace, check if it needs recovery
			currentStatus, err := m.getNamespaceStatus(ns)
			if err != nil {
				logrus.Errorf("Failed to get status for namespace %s: %v", ns, err)
				continue
			}

			if currentStatus != consts.CommonEnabled {
				// Namespace was disabled/deleted but is back in config, recover it
				if err := m.setNamespaceStatus(ns, consts.CommonEnabled); err != nil {
					logrus.Errorf("Failed to recover namespace %s: %v", ns, err)
				} else {
					result.Recovered = append(result.Recovered, ns)
					logrus.Infof("Recovered namespace: %s (was %s)", ns, consts.GetStatusTypeName(currentStatus))
				}
			}
			// If already enabled, no action needed
		}
	}

	// Handle namespaces removed from config
	for ns := range existingSet {
		if _, exists := latestSet[ns]; !exists {
			// Namespace removed from config
			currentStatus, err := m.getNamespaceStatus(ns)
			if err != nil {
				logrus.Errorf("Failed to get status for namespace %s: %v", ns, err)
				continue
			}

			// Skip if already disabled or deleted
			if currentStatus == consts.CommonDisabled {
				logrus.Debugf("Namespace %s already disabled, skipping", ns)
				continue
			}
			if currentStatus == consts.CommonDeleted {
				logrus.Debugf("Namespace %s already deleted, skipping", ns)
				continue
			}

			// Check if namespace has active lock
			isLocked, err := m.isNamespaceLocked(ns)
			if err != nil {
				logrus.Errorf("Failed to check lock status for %s: %v", ns, err)
				continue
			}

			if isLocked {
				// Has active lock, mark as disabled
				if err := m.setNamespaceStatus(ns, consts.CommonDisabled); err != nil {
					logrus.Errorf("Failed to set namespace %s status to disabled: %v", ns, err)
				} else {
					result.Disabled = append(result.Disabled, ns)
					logrus.Warnf("Namespace %s marked as disabled (has active lock)", ns)
				}
			} else {
				// No active lock, mark as deleted
				if err := m.setNamespaceStatus(ns, consts.CommonDeleted); err != nil {
					logrus.Errorf("Failed to set namespace %s status to deleted: %v", ns, err)
				} else {
					result.Deleted = append(result.Deleted, ns)
					logrus.Infof("Namespace %s marked as deleted (no active lock)", ns)
				}
			}
		}
	}

	return result, nil
}

// addNamespace adds a new namespace to Redis with initial state (idempotent)
func (m *monitor) addNamespace(namespace string, endTime time.Time) error {
	return m.seedNamespace(namespace, endTime)
}

// isNamespaceLocked checks if a namespace currently has an active lock
func (m *monitor) isNamespaceLocked(namespace string) (bool, error) {
	return m.locks.isActive(m.currentContext(), namespace, time.Now())
}

// getNamespaceStatus gets the status of a namespace
func (m *monitor) getNamespaceStatus(namespace string) (consts.StatusType, error) {
	return m.status.get(m.currentContext(), namespace)
}

// setNamespaceStatus sets the status of a namespace
func (m *monitor) setNamespaceStatus(namespace string, status consts.StatusType) error {
	return m.status.set(m.currentContext(), namespace, status)
}
