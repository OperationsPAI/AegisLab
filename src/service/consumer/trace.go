package consumer

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/service/common"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// levelStatistics holds statistics for a specific level in the task tree
type levelStatistics struct {
	Total     int
	Completed int
	Failed    int
	Running   int
	Pending   int
}

// -----------------------------------------------------------------------------
// Trace State Update Functions
// -----------------------------------------------------------------------------

// updateTraceState updates trace state based on task state change
// This function is called after task state is persisted to ensure real-time sync
func updateTraceState(traceID, taskID string, newState consts.TaskState, event *dto.StreamEvent) error {
	logEntry := logrus.WithField("trace_id", traceID).WithField("task_id", taskID)

	// Update trace state asynchronously to avoid blocking task processing
	go func() {
		// Use background context since this is async
		ctx := context.Background()

		if err := performTraceStateUpdate(ctx, traceID, taskID, newState, event); err != nil {
			logEntry.Errorf("failed to update trace state: %v", err)
		}
	}()

	return nil
}

// performTraceStateUpdate performs the actual trace state update with retry logic
func performTraceStateUpdate(ctx context.Context, traceID, taskID string, newState consts.TaskState, event *dto.StreamEvent) error {
	const maxRetries = 3
	logEntry := logrus.WithField("trace_id", traceID)

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := tryUpdateTraceStateCore(ctx, traceID, taskID, newState, event)
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
func tryUpdateTraceStateCore(ctx context.Context, traceID, taskID string, newState consts.TaskState, event *dto.StreamEvent) error {
	logEntry := logrus.WithField("trace_id", traceID)

	// 1. Fetch trace with all tasks (including the just-updated task)
	trace, err := repository.GetTraceByID(database.DB, traceID)
	if err != nil {
		return fmt.Errorf("failed to get trace: %w", err)
	}

	// Store original updated_at for optimistic locking
	originalUpdatedAt := trace.UpdatedAt

	// 2. Find the task that was just updated
	var updatedTask *database.Task
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
	inferredState, inferredEvent := inferTraceState(trace, trace.Tasks)

	// Special handling for CollectResult task: use the provided event directly
	// CollectResult tasks provide specific events like EventDatapackResultCollection, EventDatapackNoAnomaly
	// that are more accurate than inferred events
	if updatedTask.Type == consts.TaskTypeCollectResult && event != nil && event.EventName != "" {
		inferredEvent = event.EventName
		logEntry.Debugf("using explicit event from CollectResult task: %s", inferredEvent)
	}

	logEntry.Debugf("inferred trace state: %s, event: %s (triggered by task %s: %s)",
		consts.GetTraceStateName(inferredState),
		inferredEvent,
		taskID,
		consts.GetTaskStateName(newState))

	// 4. Check if update is necessary (skip if state unchanged to reduce DB writes)
	if trace.State == inferredState && trace.LastEvent == inferredEvent {
		logEntry.Debugf("trace state unchanged, skipping update")
		return nil
	}

	// 5. Prepare update data
	updates := map[string]interface{}{
		"state":      inferredState,
		"last_event": inferredEvent,
		"updated_at": time.Now(),
	}

	// Set end time for terminal states
	if (inferredState == consts.TraceCompleted || inferredState == consts.TraceFailed) && trace.EndTime == nil {
		now := time.Now()
		updates["end_time"] = &now
	}

	// 6. Execute optimistic locking update
	result := database.DB.Model(&database.Trace{}).
		Where("id = ? AND updated_at = ?", traceID, originalUpdatedAt).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update trace: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("optimistic lock conflict: trace was modified by another job")
	}

	// 7. Publish trace state change event to Redis Stream
	stream := fmt.Sprintf(consts.StreamLogKey, traceID)
	traceEvent := dto.StreamEvent{
		TaskID:    traceID,             // Use trace ID for trace-level events
		TaskType:  consts.TaskType(-1), // Special marker for trace-level events
		EventName: inferredEvent,
		Payload:   nil,
	}

	if event != nil && event.Payload != nil {
		traceEvent.Payload = event.Payload
	}

	publishEvent(ctx, stream, traceEvent, withCallerLevel(4))

	logEntry.Infof("trace state updated: %s -> %s, event: %s (triggered by task state change)",
		consts.GetTraceStateName(trace.State),
		consts.GetTraceStateName(inferredState),
		inferredEvent)

	return nil
}

// buildLevelStatistics constructs level statistics from task list
func buildLevelStatistics(tasks []database.Task, treeHeight int) map[int]*levelStatistics {
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

// selectBestLastEvent selects the most appropriate last event from completed leaf tasks
func selectBestLastEvent(tasks []database.Task, leafLevel int) consts.EventType {
	// Event priority map: higher value = higher priority
	eventPriority := map[consts.EventType]int{
		consts.EventDatapackResultCollection: 100,
		consts.EventDatapackNoAnomaly:        90,
		consts.EventFaultInjectionCompleted:  80,
		consts.EventAlgoRunSucceed:           70,
		consts.EventDatapackBuildSucceed:     60,
		consts.EventRestartPedestalCompleted: 50,
	}

	var bestEvent consts.EventType
	var bestPriority int = -1
	var latestTime time.Time

	for _, task := range tasks {
		if task.Level != leafLevel || task.State != consts.TaskCompleted {
			continue
		}

		// Get event type from task type and state mapping
		eventType := common.GetEventTypeByTask(task.Type, task.State)
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
func inferTraceState(trace *database.Trace, tasks []database.Task) (consts.TraceState, consts.EventType) {
	treeHeight := consts.TraceTypeHeightMap[trace.Type]
	stats := buildLevelStatistics(tasks, treeHeight)

	// State inference with priority: Failed > Completed > Running > Pending

	// Priority 1: Check if any level has all tasks failed
	for level := range treeHeight {
		levelStat := stats[level]
		if levelStat.Total > 0 && levelStat.Failed == levelStat.Total {
			// All tasks at this level failed -> Trace failed
			lastEvent := selectBestLastEvent(tasks, level)
			if lastEvent == consts.EventTaskStateUpdate {
				// Find any failed task at this level to get its event
				for _, task := range tasks {
					if task.Level == level && task.State == consts.TaskError {
						lastEvent = common.GetEventTypeByTask(task.Type, task.State)
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
		lastEvent := selectBestLastEvent(tasks, leafLevel)
		return consts.TraceCompleted, lastEvent
	}

	// Priority 3: Check if any task is running
	for level := range treeHeight {
		if stats[level].Running > 0 {
			// Find the first running task to get its event
			var lastEvent consts.EventType
			for _, task := range tasks {
				if task.State == consts.TaskRunning {
					lastEvent = common.GetEventTypeByTask(task.Type, task.State)
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
					lastEvent = common.GetEventTypeByTask(task.Type, task.State)
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

// isOptimisticLockError checks if an error is due to optimistic lock failure
func isOptimisticLockError(err error) bool {
	return err != nil && err.Error() == "optimistic lock conflict: trace was modified by another job"
}
