package taskmodule

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	writeWait            = 10 * time.Second
	pongWait             = 60 * time.Second
	pingPeriod           = 54 * time.Second
	maxMsgSize           = 512
	taskPollInterval     = 5 * time.Second
	completionFlushDelay = 5 * time.Second
)

type TaskLogService struct {
	repository *Repository
	queueStore *TaskQueueStore
	loki       *LokiGateway
}

func NewTaskLogService(repository *Repository, queueStore *TaskQueueStore, loki *LokiGateway) *TaskLogService {
	return &TaskLogService{
		repository: repository,
		queueStore: queueStore,
		loki:       loki,
	}
}

func (s *TaskLogService) StreamLogs(ctx context.Context, conn *websocket.Conn, task *model.Task) {
	streamer := &taskLogStreamer{
		ctx:     ctx,
		conn:    conn,
		task:    task,
		taskID:  task.ID,
		service: s,
		log:     logrus.WithField("task_id", task.ID),
	}
	streamer.StreamLogs(ctx)
}

type taskLogStreamer struct {
	ctx     context.Context
	conn    *websocket.Conn
	mu      sync.Mutex
	log     *logrus.Entry
	taskID  string
	task    *model.Task
	service *TaskLogService
}

func (s *taskLogStreamer) StreamLogs(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.conn.SetReadLimit(maxMsgSize)
	_ = s.conn.SetReadDeadline(time.Now().Add(pongWait))
	s.conn.SetPongHandler(func(string) error {
		_ = s.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	go s.runReadLoop(cancel)
	go s.runPingLoop(ctx, cancel)

	pubsub, err := s.service.queueStore.SubscribeJobLogs(ctx, s.taskID)
	if err != nil {
		s.log.Errorf("Failed to subscribe to Redis Pub/Sub for task logs: %v", err)
		s.WriteMessage(WSLogMessage{
			Type:    consts.WSLogTypeError,
			Message: "failed to subscribe to log stream",
		})
		return
	}
	defer func() { _ = pubsub.Close() }()
	s.log.Info("Subscribed to Redis Pub/Sub for real-time logs")

	lastHistoricalTime := s.sendHistoricalLogs()

	if isTaskTerminal(s.task.State) {
		s.WriteMessage(WSLogMessage{
			Type:    consts.WSLogTypeEnd,
			Message: "task already completed",
		})
		s.closeNormal("task completed")
		return
	}

	s.streamRealtime(ctx, pubsub.Channel(), lastHistoricalTime)
}

func (s *taskLogStreamer) runReadLoop(cancel context.CancelFunc) {
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
}

func (s *taskLogStreamer) WriteMessage(msg WSLogMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = s.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := s.conn.WriteJSON(msg); err != nil {
		s.log.Warnf("WebSocket write error: %v", err)
	}
}

func (s *taskLogStreamer) ForwardRedisLog(payload string, lastHistoricalTime time.Time) {
	var entry dto.LogEntry
	if err := json.Unmarshal([]byte(payload), &entry); err != nil {
		s.log.Warnf("Failed to unmarshal Redis log message: %v", err)
		return
	}

	if !lastHistoricalTime.IsZero() && !entry.Timestamp.After(lastHistoricalTime) {
		return
	}

	s.WriteMessage(WSLogMessage{
		Type: consts.WSLogTypeRealtime,
		Logs: []dto.LogEntry{entry},
	})
}

func (s *taskLogStreamer) runPingLoop(ctx context.Context, cancel context.CancelFunc) {
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

func (s *taskLogStreamer) sendHistoricalLogs() time.Time {
	lokiCtx, lokiCancel := context.WithTimeout(s.ctx, 15*time.Second)
	defer lokiCancel()

	historicalLogs, err := s.service.loki.QueryJobLogs(lokiCtx, s.taskID, s.task.CreatedAt)
	if err != nil {
		s.log.Warnf("Failed to query Loki for historical logs: %v", err)
		return time.Time{}
	}

	if len(historicalLogs) > 0 {
		s.WriteMessage(WSLogMessage{
			Type:  consts.WSLogTypeHistory,
			Logs:  historicalLogs,
			Total: len(historicalLogs),
		})
		s.log.Infof("Sent %d historical log entries", len(historicalLogs))
		return historicalLogs[len(historicalLogs)-1].Timestamp
	}

	return time.Time{}
}

func (s *taskLogStreamer) streamRealtime(ctx context.Context, redisCh <-chan *redis.Message, lastHistoricalTime time.Time) {
	taskDoneCh := make(chan struct{})
	go s.pollTaskCompletion(ctx, taskDoneCh)

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
				s.WriteMessage(WSLogMessage{
					Type:    consts.WSLogTypeError,
					Message: "log stream interrupted",
				})
				return
			}
			s.ForwardRedisLog(msg.Payload, lastHistoricalTime)
		}
	}
}

func (s *taskLogStreamer) pollTaskCompletion(ctx context.Context, taskDoneCh chan<- struct{}) {
	ticker := time.NewTicker(taskPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			task, err := s.service.repository.GetByID(s.taskID)
			if err != nil {
				s.log.Warnf("Failed to poll task state: %v", err)
				continue
			}
			if isTaskTerminal(task.State) {
				s.log.Info("Task detected as terminal, initiating close")
				close(taskDoneCh)
				return
			}
		}
	}
}

func (s *taskLogStreamer) flushAndClose(redisCh <-chan *redis.Message, lastHistoricalTime time.Time) {
	s.log.Info("Task completed, flushing remaining logs...")
	flushTimer := time.NewTimer(completionFlushDelay)
	defer flushTimer.Stop()

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

	s.WriteMessage(WSLogMessage{
		Type:    consts.WSLogTypeEnd,
		Message: "task completed",
	})
	s.closeNormal("task completed")
}

func (s *taskLogStreamer) closeNormal(reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = s.conn.SetWriteDeadline(time.Now().Add(writeWait))
	_ = s.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason))
}

func isTaskTerminal(state consts.TaskState) bool {
	return state == consts.TaskCompleted || state == consts.TaskError || state == consts.TaskCancelled
}
