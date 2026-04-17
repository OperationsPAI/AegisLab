package consumer

import (
	"aegis/consts"
	"aegis/dto"
	redisinfra "aegis/infra/redis"
	"aegis/model"
	groupmodule "aegis/module/group"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// levelStatistics holds statistics for a specific level in the task tree
type levelStatistics struct {
	Total     int
	Completed int
	Failed    int
	Running   int
	Pending   int
}

var traceTypeHeightMap = map[consts.TraceType]int{
	consts.TraceTypeFullPipeline:   7,
	consts.TraceTypeFaultInjection: 5,
	consts.TraceTypeDatapackBuild:  3,
	consts.TraceTypeAlgorithmRun:   2,
}

var traceTaskEventMap = map[consts.TaskType]map[consts.TaskState]consts.EventType{
	consts.TaskTypeRestartPedestal: {
		consts.TaskRunning:     consts.EventRestartPedestalStarted,
		consts.TaskCompleted:   consts.EventRestartPedestalCompleted,
		consts.TaskError:       consts.EventRestartPedestalFailed,
		consts.TaskRescheduled: consts.EventNoNamespaceAvailable,
	},
	consts.TaskTypeFaultInjection: {
		consts.TaskRunning:   consts.EventFaultInjectionStarted,
		consts.TaskCompleted: consts.EventFaultInjectionCompleted,
		consts.TaskError:     consts.EventFaultInjectionFailed,
	},
	consts.TaskTypeBuildDatapack: {
		consts.TaskRunning:   consts.EventDatapackBuildStarted,
		consts.TaskCompleted: consts.EventDatapackBuildSucceed,
		consts.TaskError:     consts.EventDatapackBuildFailed,
	},
	consts.TaskTypeRunAlgorithm: {
		consts.TaskRunning:     consts.EventAlgoRunStarted,
		consts.TaskCompleted:   consts.EventAlgoRunSucceed,
		consts.TaskError:       consts.EventAlgoRunFailed,
		consts.TaskRescheduled: consts.EventNoTokenAvailable,
	},
}

// getEventTypeByTask maps a task type and state to the corresponding event type
func getEventTypeByTask(taskType consts.TaskType, taskState consts.TaskState) consts.EventType {
	if taskType == consts.TaskTypeCollectResult {
		return "unknown"
	}

	stateMap, exists := traceTaskEventMap[taskType]
	if !exists {
		logrus.Warnf("no event type mapping for task type: %s", consts.GetTaskTypeName(taskType))
		return "unknown"
	}

	eventType, exists := stateMap[taskState]
	if !exists {
		logrus.Warnf("no event type mapping for task state: %s", consts.GetTaskStateName(taskState))
		return "unknown"
	}

	return eventType
}

// -----------------------------------------------------------------------------
// Trace State Update Functions
// -----------------------------------------------------------------------------

// updateTraceState updates trace state based on task state change
// This function is called after task state is persisted to ensure real-time sync
func updateTraceState(redisGateway *redisinfra.Gateway, db *gorm.DB, traceID, taskID string, newState consts.TaskState, event *dto.TraceStreamEvent) error {
	logEntry := logrus.WithField("trace_id", traceID).WithField("task_id", taskID)

	// Update trace state asynchronously to avoid blocking task processing
	go func() {
		ctx := consumerDetachedContext()

		if err := performTraceStateUpdate(redisGateway, ctx, db, traceID, taskID, newState, event); err != nil {
			logEntry.Errorf("failed to update trace state: %v", err)
		}
	}()

	return nil
}

// performTraceStateUpdate performs the actual trace state update with retry logic
func performTraceStateUpdate(redisGateway *redisinfra.Gateway, ctx context.Context, db *gorm.DB, traceID, taskID string, newState consts.TaskState, event *dto.TraceStreamEvent) error {
	const maxRetries = 3
	logEntry := logrus.WithField("trace_id", traceID)

	for attempt := range maxRetries {
		err := tryUpdateTraceStateCore(redisGateway, ctx, db, traceID, taskID, newState, event)
		if err == nil {
			return nil
		}

		// Check if it's a version conflict (optimistic lock failure)
		if isOptimisticLockError(err) && attempt < maxRetries-1 {
			logEntry.Warnf("optimistic lock conflict on attempt %d, retrying...", attempt+1)
			time.Sleep(time.Millisecond * 50 * time.Duration(attempt+1)) // Exponential backoff
			continue
		}

		return err
	}

	return fmt.Errorf("failed to update trace state after %d attempts", maxRetries)
}

// tryUpdateTraceStateCore attempts to update trace state once
func tryUpdateTraceStateCore(redisGateway *redisinfra.Gateway, ctx context.Context, db *gorm.DB, traceID, taskID string, newState consts.TaskState, streamEvent *dto.TraceStreamEvent) error {
	if db == nil {
		return fmt.Errorf("trace state update db is nil")
	}

	logEntry := logrus.WithField("trace_id", traceID)

	// 1. Fetch trace with all tasks (including the just-updated task)
	trace, err := getTraceByID(db, traceID)
	if err != nil {
		return fmt.Errorf("failed to get trace: %w", err)
	}

	// Store original updated_at for optimistic locking
	originalUpdatedAt := trace.UpdatedAt

	// 2. Find the task that was just updated
	var updatedTask *model.Task
	for i := range trace.Tasks {
		if trace.Tasks[i].ID == taskID {
			updatedTask = &trace.Tasks[i]
			break
		}
	}

	if updatedTask == nil {
		return fmt.Errorf("task %s not found in trace", taskID)
	}

	// 3. Infer new trace state and event from all current tasks
	// Pass streamEvent to help distinguish early termination vs continuation scenarios
	inferredState, inferredEventType := inferTraceState(trace, trace.Tasks, streamEvent)

	// Special handling for CollectResult task: use the provided event directly
	// CollectResult tasks provide specific events like EventDatapackResultCollection, EventDatapackNoAnomaly
	// that are more accurate than inferred events
	if updatedTask.Type == consts.TaskTypeCollectResult && streamEvent != nil && streamEvent.EventName != "" {
		// Always use the explicit event from CollectResult task for more accurate event tracking
		inferredEventType = streamEvent.EventName
		logEntry.Debugf("using explicit event from CollectResult task: %s", inferredEventType)

		// For FullPipeline with early termination events, mark trace as completed
		// EventDatapackNoAnomaly and EventDatapackNoDetectorData indicate no further processing needed
		if trace.Type == consts.TraceTypeFullPipeline &&
			(streamEvent.EventName == consts.EventDatapackNoAnomaly || streamEvent.EventName == consts.EventDatapackNoDetectorData) {
			inferredState = consts.TraceCompleted
			logEntry.Debugf("FullPipeline early termination detected, marking trace as completed")
		}
	}

	logEntry.Debugf("inferred trace state: %s, event: %s (triggered by task %s: %s)",
		consts.GetTraceStateName(inferredState),
		inferredEventType,
		taskID,
		consts.GetTaskStateName(newState))

	// 4. Check if update is necessary (skip if state unchanged to reduce DB writes)
	if trace.State == inferredState && trace.LastEvent == inferredEventType {
		logEntry.Debugf("trace state unchanged, skipping update")
		return nil
	}

	// 5. Prepare update data
	updates := map[string]any{
		"state":      inferredState,
		"last_event": inferredEventType,
		"updated_at": time.Now(),
	}

	// Set end time for terminal states
	if (inferredState == consts.TraceCompleted || inferredState == consts.TraceFailed) && trace.EndTime == nil {
		now := time.Now()
		updates["end_time"] = &now

		// Publish to group-level stream for real-time group progress SSE
		if trace.GroupID != "" {
			publishGroupStreamEvent(redisGateway, ctx, trace.GroupID, traceID, inferredState, inferredEventType)
		}
	}

	// 6. Execute optimistic locking update
	result := db.Model(&model.Trace{}).
		Where("id = ? AND updated_at = ?", traceID, originalUpdatedAt).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update trace: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("optimistic lock conflict: trace was modified by another job")
	}

	logEntry.Infof("trace state updated: %s -> %s, event: %s (triggered by task state change)",
		consts.GetTraceStateName(trace.State),
		consts.GetTraceStateName(inferredState),
		inferredEventType)

	return nil
}

// buildLevelStatistics constructs level statistics from task list
func buildLevelStatistics(tasks []model.Task, treeHeight int) map[int]*levelStatistics {
	stats := make(map[int]*levelStatistics)

	// Initialize statistics for each level
	for i := range treeHeight {
		stats[i] = &levelStatistics{}
	}

	// Aggregate task states by level
	for _, task := range tasks {
		if task.Level < 0 || task.Level >= treeHeight {
			logrus.Warnf("task %s has invalid level %d (tree height: %d)", task.ID, task.Level, treeHeight)
			continue
		}

		levelStat := stats[task.Level]
		levelStat.Total++

		switch task.State {
		case consts.TaskCompleted:
			levelStat.Completed++
		case consts.TaskError:
			levelStat.Failed++
		case consts.TaskRunning:
			levelStat.Running++
		case consts.TaskPending, consts.TaskRescheduled:
			levelStat.Pending++
		}
	}

	return stats
}

// hasEarlyTerminationEvent checks if any CollectResult task has completed with early termination events
// Now also checks the streamEvent to accurately determine if it's truly an early termination
func hasEarlyTerminationEvent(tasks []model.Task, streamEvent *dto.TraceStreamEvent) bool {
	// Priority 1: Check if streamEvent explicitly indicates early termination
	if streamEvent != nil && streamEvent.EventName != "" {
		// These events indicate early termination - no further processing needed
		if streamEvent.EventName == consts.EventDatapackNoAnomaly ||
			streamEvent.EventName == consts.EventDatapackNoDetectorData {
			return true
		}
		// EventDatapackResultCollection means anomaly detected, need to continue with RCA
		if streamEvent.EventName == consts.EventDatapackResultCollection {
			return false
		}
	}

	// Priority 2: Fallback to checking task structure
	// This handles cases where streamEvent is not available
	for _, task := range tasks {
		if task.Type == consts.TaskTypeCollectResult && task.State == consts.TaskCompleted {
			// Check task metadata or level to determine if it's a detector CollectResult
			// Detector CollectResult is at level 4 in FullPipeline:
			// L0: RestartPedestal, L1: FaultInjection, L2: BuildDatapack,
			// L3: RunAlgorithm(detector), L4: CollectResult(detector) <- early termination point
			if task.Level == 4 {
				// Without streamEvent, we can't determine if it's early termination
				// Need to check if RCA tasks exist
				hasRCAAlgorithmTasks := false
				for _, t := range tasks {
					if t.Type == consts.TaskTypeRunAlgorithm && t.Level >= 5 {
						hasRCAAlgorithmTasks = true
						break
					}
				}
				// Only consider it early termination if no RCA tasks were created
				return !hasRCAAlgorithmTasks
			}
		}
	}
	return false
}

// findEarlyTerminationEvent finds and returns the early termination event from completed CollectResult tasks
// Now uses streamEvent to accurately return the correct event type
func findEarlyTerminationEvent(tasks []model.Task, streamEvent *dto.TraceStreamEvent) consts.EventType {
	// Priority 1: If streamEvent is provided with early termination events, use it directly
	if streamEvent != nil && streamEvent.EventName != "" {
		if streamEvent.EventName == consts.EventDatapackNoAnomaly ||
			streamEvent.EventName == consts.EventDatapackNoDetectorData {
			return streamEvent.EventName
		}
		// EventDatapackResultCollection is NOT an early termination event
		if streamEvent.EventName == consts.EventDatapackResultCollection {
			return "" // Not early termination
		}
	}

	// Priority 2: Fallback to checking task structure
	// Look for completed CollectResult tasks at level 4 (detector CollectResult level)
	for _, task := range tasks {
		if task.Type == consts.TaskTypeCollectResult && task.State == consts.TaskCompleted && task.Level == 4 {
			// Check if there are no subsequent RCA algorithm tasks (level 5+)
			// Detector RunAlgorithm is at level 3, RCA algorithms are at level 5
			hasRCAAlgorithmTasks := false
			for _, t := range tasks {
				if t.Type == consts.TaskTypeRunAlgorithm && t.Level >= 5 {
					hasRCAAlgorithmTasks = true
					break
				}
			}

			if !hasRCAAlgorithmTasks {
				// Early termination confirmed, but without streamEvent we can't determine exact event
				// Return a conservative default (this case should be rare with streamEvent passing)
				return consts.EventDatapackNoAnomaly
			}
		}
	}
	return ""
}

// selectBestLastEvent selects the most appropriate last event from completed leaf tasks
func selectBestLastEvent(tasks []model.Task, leafLevel int, streamEvent *dto.TraceStreamEvent) consts.EventType {
	// Event priority map: higher value = higher priority
	eventPriority := map[consts.EventType]int{
		consts.EventFaultInjectionCompleted:  80,
		consts.EventAlgoRunSucceed:           70,
		consts.EventDatapackBuildSucceed:     60,
		consts.EventRestartPedestalCompleted: 50,
	}

	var bestEvent consts.EventType
	bestPriority := -1
	var latestTime time.Time

	// First, check for early termination events at any level (not just leaf level)
	// This handles FullPipeline cases where CollectResult at level 4 is the actual end
	earlyTerminationEvent := findEarlyTerminationEvent(tasks, streamEvent)
	if earlyTerminationEvent != "" {
		return earlyTerminationEvent
	}

	// Then check leaf level tasks as before
	for _, task := range tasks {
		if task.Level != leafLevel || task.State != consts.TaskCompleted {
			continue
		}

		// Get event type from task type and state mapping
		eventType := getEventTypeByTask(task.Type, task.State)
		priority, exists := eventPriority[eventType]

		if !exists {
			priority = 0
		}

		// Select by priority, or by latest update time if priority is same
		if priority > bestPriority || (priority == bestPriority && task.UpdatedAt.After(latestTime)) {
			bestEvent = eventType
			bestPriority = priority
			latestTime = task.UpdatedAt
		}
	}

	// Fallback to task state update event if no specific event found
	if bestEvent == "" {
		bestEvent = consts.EventTaskStateUpdate
	}

	return bestEvent
}

// inferTraceState infers trace state and last event from all tasks
// streamEvent parameter helps distinguish between early termination vs continuation scenarios
func inferTraceState(trace *model.Trace, tasks []model.Task, streamEvent *dto.TraceStreamEvent) (consts.TraceState, consts.EventType) {
	treeHeight := traceTypeHeightMap[trace.Type]
	stats := buildLevelStatistics(tasks, treeHeight)

	// State inference with priority: Failed > Completed > Running > Pending

	// Priority 0: Check for early termination in FullPipeline
	// When CollectResult task completes with no_anomaly or no_detector_data,
	// the pipeline ends early without executing algorithm tasks
	if trace.Type == consts.TraceTypeFullPipeline {
		if hasEarlyTerminationEvent(tasks, streamEvent) {
			// Found early termination event, trace should complete
			lastEvent := findEarlyTerminationEvent(tasks, streamEvent)
			if lastEvent != "" {
				return consts.TraceCompleted, lastEvent
			}
		}
	}

	// Priority 1: Check if any level has all tasks failed
	for level := range treeHeight {
		levelStat := stats[level]
		if levelStat.Total > 0 && levelStat.Failed == levelStat.Total {
			// All tasks at this level failed -> Trace failed
			lastEvent := selectBestLastEvent(tasks, level, streamEvent)
			if lastEvent == consts.EventTaskStateUpdate {
				// Find any failed task at this level to get its event
				for _, task := range tasks {
					if task.Level == level && task.State == consts.TaskError {
						lastEvent = getEventTypeByTask(task.Type, task.State)
						break
					}
				}
			}
			return consts.TraceFailed, lastEvent
		}
	}

	// Priority 2: Check if any leaf node completed (success condition)
	leafLevel := treeHeight - 1
	leafStat := stats[leafLevel]

	// For FullPipeline: LeafNum might be > 1, only need one path to succeed
	// For other types: LeafNum should be 1
	if leafStat.Completed > 0 {
		// Check if there are any tasks still running or pending at any level
		hasActiveOrPendingTasks := false
		for level := range treeHeight {
			if stats[level].Running > 0 || stats[level].Pending > 0 {
				hasActiveOrPendingTasks = true
				break
			}
		}

		// Only mark as completed if no active or pending tasks remain
		if !hasActiveOrPendingTasks {
			lastEvent := selectBestLastEvent(tasks, leafLevel, streamEvent)
			return consts.TraceCompleted, lastEvent
		}
	}

	// Priority 3: Check if any task is running
	for level := range treeHeight {
		if stats[level].Running > 0 {
			// Find the first running task to get its event
			var lastEvent consts.EventType
			for _, task := range tasks {
				if task.State == consts.TaskRunning {
					lastEvent = getEventTypeByTask(task.Type, task.State)
					if lastEvent != "" && lastEvent != "unknown" {
						break
					}
				}
			}
			if lastEvent == "" || lastEvent == "unknown" {
				lastEvent = consts.EventTaskStateUpdate
			}
			return consts.TraceRunning, lastEvent
		}
	}

	// Priority 4: Check if any task has completed (trace has started)
	// Once trace starts running, it should never go back to Pending
	for level := range treeHeight {
		if stats[level].Completed > 0 {
			// Trace has started and is waiting for next tasks
			// Use the last completed task's event
			var lastEvent consts.EventType
			var latestTime time.Time
			for _, task := range tasks {
				if task.State == consts.TaskCompleted && task.UpdatedAt.After(latestTime) {
					lastEvent = getEventTypeByTask(task.Type, task.State)
					latestTime = task.UpdatedAt
				}
			}
			if lastEvent == "" || lastEvent == "unknown" {
				lastEvent = consts.EventTaskStateUpdate
			}
			return consts.TraceRunning, lastEvent
		}
	}

	// Default: Pending (only if no tasks have started or completed)
	return consts.TracePending, consts.EventTaskStateUpdate
}

func getTraceByID(db *gorm.DB, traceID string) (*model.Trace, error) {
	var trace model.Trace
	if err := db.Model(&model.Trace{}).
		Preload("Project").
		Preload("Tasks", func(db *gorm.DB) *gorm.DB {
			return db.Order("level ASC, sequence ASC")
		}).
		Where("id = ? AND status != ?", traceID, consts.CommonDeleted).
		First(&trace).Error; err != nil {
		return nil, err
	}
	return &trace, nil
}

// isOptimisticLockError checks if an error is due to optimistic lock failure
func isOptimisticLockError(err error) bool {
	return err != nil && err.Error() == "optimistic lock conflict: trace was modified by another job"
}

// publishGroupStreamEvent publishes a lightweight event to the group-level Redis stream
// when a trace reaches a terminal state (Completed/Failed).
// This enables real-time SSE updates for group progress tracking on the frontend.
func publishGroupStreamEvent(redisGateway *redisinfra.Gateway, ctx context.Context, groupID, traceID string, state consts.TraceState, lastEvent consts.EventType) {
	streamKey := fmt.Sprintf(consts.StreamGroupLogKey, groupID)
	logEntry := logrus.WithField("group_id", groupID).WithField("trace_id", traceID)

	event := &groupmodule.GroupStreamEvent{
		TraceID:   traceID,
		State:     state,
		LastEvent: lastEvent,
	}

	if err := publishRedisStreamEvent(redisGateway, ctx, streamKey, event); err != nil {
		logEntry.Errorf("failed to publish group stream event: %v", err)
		return
	}
}
