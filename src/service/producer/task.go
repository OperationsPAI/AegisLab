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
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	// WebSocket timing configuration
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 54 * time.Second // Must be less than pongWait
	maxMsgSize = 512              // Max size of incoming messages (control frames)

	// Task polling interval for completion detection
	taskPollInterval = 5 * time.Second

	// Flush delay after task completion to catch remaining logs
	completionFlushDelay = 5 * time.Second
)

// TaskLogStreamer manages WebSocket-based real-time log streaming for a task.
type TaskLogStreamer struct {
	conn   *websocket.Conn
	mu     sync.Mutex
	log    *logrus.Entry
	taskID string
}

// NewTaskLogStreamer creates a new TaskLogStreamer for the given WebSocket connection and task.
func NewTaskLogStreamer(conn *websocket.Conn, taskID string) *TaskLogStreamer {
	return &TaskLogStreamer{
		conn:   conn,
		taskID: taskID,
		log:    logrus.WithField("task_id", taskID),
	}
}

// BatchDeleteTasks deletes multiple tasks by their IDs
func BatchDeleteTasks(taskIDs []string) error {
	if len(taskIDs) == 0 {
		return nil
	}

	if err := repository.BatchDeleteTasks(database.DB, taskIDs); err != nil {
		return err
	}
	return nil
}

// GetTaskDetail retrieves detailed information about a specific task
func GetTaskDetail(taskID string) (*dto.TaskDetailResp, error) {
	task, err := repository.GetTaskByID(database.DB, taskID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: task id: %s", consts.ErrNotFound, taskID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// TODO logs retrieval can be added later
	resp := dto.NewTaskDetailResp(task, []string{})
	return resp, nil
}

// ListTasks lists tasks based on filter options and pagination
func ListTasks(req *dto.ListTaskReq) (*dto.ListResp[dto.TaskResp], error) {
	if req == nil {
		return nil, fmt.Errorf("list tasks request is nil")
	}

	limit, offset := req.ToGormParams()
	fitlerOptions := req.ToFilterOptions()

	tasks, total, err := repository.ListTasks(database.DB, limit, offset, fitlerOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	taskResps := make([]dto.TaskResp, 0, len(tasks))
	for _, task := range tasks {
		taskResps = append(taskResps, *dto.NewTaskResp(&task))
	}

	resp := dto.ListResp[dto.TaskResp]{
		Items:      taskResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// StreamLogs sets up the WebSocket lifecycle, queries Loki for historical logs, subscribes to Redis Pub/Sub for real-time logs,
// and polls for task completion. It blocks until the context is cancelled or the task completes.
func (s *TaskLogStreamer) StreamLogs(ctx context.Context, task *database.Task) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Setup WebSocket connection parameters
	s.conn.SetReadLimit(maxMsgSize)
	_ = s.conn.SetReadDeadline(time.Now().Add(pongWait))
	s.conn.SetPongHandler(func(string) error {
		_ = s.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Read pump — handles client messages and detects disconnection
	go func() {
		defer cancel()
		for {
			_, _, err := s.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					s.log.Warnf("WebSocket unexpected close: %v", err)
				}
				return
			}
		}
	}()

	// Ping ticker for keepalive
	go s.runPingLoop(ctx, cancel)

	// Step 1: Subscribe to Redis Pub/Sub first (before querying Loki to avoid gaps)
	pubsubChannel := "joblogs:" + s.taskID
	pubsub := client.GetRedisClient().Subscribe(ctx, pubsubChannel)
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		s.log.Errorf("Failed to subscribe to Redis Pub/Sub channel %s: %v", pubsubChannel, err)
		s.WriteMessage(dto.WSLogMessage{
			Type:    consts.WSLogTypeError,
			Message: "failed to subscribe to log stream",
		})
		return
	}
	s.log.Info("Subscribed to Redis Pub/Sub for real-time logs")

	// Step 2: Query Loki for historical logs
	lastHistoricalTime := s.sendHistoricalLogs(ctx, task)

	// Step 3: Check if task is already completed
	if isTaskTerminal(task.State) {
		s.WriteMessage(dto.WSLogMessage{
			Type:    consts.WSLogTypeEnd,
			Message: "task already completed",
		})
		s.closeNormal("task completed")
		return
	}

	// Step 4: Forward real-time logs from Redis Pub/Sub
	s.streamRealtime(ctx, pubsub.Channel(), lastHistoricalTime)
}

// WriteMessage sends a WSLogMessage to the WebSocket connection with thread-safe locking.
func (s *TaskLogStreamer) WriteMessage(msg dto.WSLogMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = s.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := s.conn.WriteJSON(msg); err != nil {
		s.log.Warnf("WebSocket write error: %v", err)
	}
}

// ForwardRedisLog parses a Redis Pub/Sub payload and forwards it as a realtime log entry.
// It deduplicates against lastHistoricalTime to avoid sending overlapping logs.
func (s *TaskLogStreamer) ForwardRedisLog(payload string, lastHistoricalTime time.Time) {
	var entry dto.LogEntry
	if err := json.Unmarshal([]byte(payload), &entry); err != nil {
		s.log.Warnf("Failed to unmarshal Redis log message: %v", err)
		return
	}

	// Deduplicate: skip entries that are before or equal to the last historical log
	if !lastHistoricalTime.IsZero() && !entry.Timestamp.After(lastHistoricalTime) {
		return
	}

	s.WriteMessage(dto.WSLogMessage{
		Type: consts.WSLogTypeRealtime,
		Logs: []dto.LogEntry{entry},
	})
}

// runPingLoop sends periodic WebSocket ping messages to keep the connection alive.
func (s *TaskLogStreamer) runPingLoop(ctx context.Context, cancel context.CancelFunc) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			_ = s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := s.conn.WriteMessage(websocket.PingMessage, nil)
			s.mu.Unlock()
			if err != nil {
				cancel()
				return
			}
		}
	}
}

// sendHistoricalLogs queries Loki for historical logs and sends them to the client.
// Returns the timestamp of the last historical entry for deduplication.
func (s *TaskLogStreamer) sendHistoricalLogs(ctx context.Context, task *database.Task) time.Time {
	lokiClient := client.NewLokiClient()
	queryOpts := client.QueryOpts{
		Start:     task.CreatedAt,
		Direction: "forward",
	}

	historicalLogs, err := lokiClient.QueryJobLogs(ctx, s.taskID, queryOpts)
	if err != nil {
		s.log.Warnf("Failed to query Loki for historical logs: %v", err)
		return time.Time{}
	}

	if len(historicalLogs) > 0 {
		s.WriteMessage(dto.WSLogMessage{
			Type:  consts.WSLogTypeHistory,
			Logs:  historicalLogs,
			Total: len(historicalLogs),
		})
		s.log.Infof("Sent %d historical log entries", len(historicalLogs))
		return historicalLogs[len(historicalLogs)-1].Timestamp
	}

	return time.Time{}
}

// streamRealtime forwards real-time logs from Redis Pub/Sub and polls for task completion.
func (s *TaskLogStreamer) streamRealtime(ctx context.Context, redisCh <-chan *redis.Message, lastHistoricalTime time.Time) {
	// Task completion polling
	taskDoneCh := make(chan struct{})
	go func() {
		ticker := time.NewTicker(taskPollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				t, err := repository.GetTaskByID(database.DB, s.taskID)
				if err != nil {
					s.log.Warnf("Failed to poll task state: %v", err)
					continue
				}
				if isTaskTerminal(t.State) {
					s.log.Info("Task detected as terminal, initiating close")
					close(taskDoneCh)
					return
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			s.log.Info("Context cancelled, closing WebSocket")
			return

		case <-taskDoneCh:
			s.flushAndClose(redisCh, lastHistoricalTime)
			return

		case msg, ok := <-redisCh:
			if !ok {
				s.log.Warn("Redis Pub/Sub channel closed")
				s.WriteMessage(dto.WSLogMessage{
					Type:    consts.WSLogTypeError,
					Message: "log stream interrupted",
				})
				return
			}
			s.ForwardRedisLog(msg.Payload, lastHistoricalTime)
		}
	}
}

// flushAndClose drains remaining Redis messages after task completion, then closes.
func (s *TaskLogStreamer) flushAndClose(redisCh <-chan *redis.Message, lastHistoricalTime time.Time) {
	s.log.Info("Task completed, flushing remaining logs...")
	flushTimer := time.NewTimer(completionFlushDelay)

flushLoop:
	for {
		select {
		case msg, ok := <-redisCh:
			if !ok {
				break flushLoop
			}
			s.ForwardRedisLog(msg.Payload, lastHistoricalTime)
		case <-flushTimer.C:
			break flushLoop
		}
	}
	flushTimer.Stop()

	s.WriteMessage(dto.WSLogMessage{
		Type:    consts.WSLogTypeEnd,
		Message: "task completed",
	})
	s.closeNormal("task completed")
}

// closeNormal sends a WebSocket close frame with NormalClosure status.
func (s *TaskLogStreamer) closeNormal(reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = s.conn.SetWriteDeadline(time.Now().Add(writeWait))
	_ = s.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason))
}

// isTaskTerminal checks if a task state represents a terminal (completed/error/cancelled) state.
func isTaskTerminal(state consts.TaskState) bool {
	return state == consts.TaskCompleted || state == consts.TaskError || state == consts.TaskCancelled
}
