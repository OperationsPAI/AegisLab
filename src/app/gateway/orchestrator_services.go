package gatewayapp

import (
	"context"
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/internalclient/orchestratorclient"
	"aegis/model"
	executionmodule "aegis/module/execution"
	groupmodule "aegis/module/group"
	injectionmodule "aegis/module/injection"
	notificationmodule "aegis/module/notification"
	taskmodule "aegis/module/task"
	tracemodule "aegis/module/trace"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type remoteAwareExecutionService struct {
	executionmodule.HandlerService
	orchestrator *orchestratorclient.Client
}

func (s remoteAwareExecutionService) SubmitAlgorithmExecution(ctx context.Context, req *executionmodule.SubmitExecutionReq, groupID string, userID int) (*executionmodule.SubmitExecutionResp, error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		return s.orchestrator.SubmitExecution(ctx, req, groupID, userID)
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

type remoteAwareInjectionService struct {
	injectionmodule.HandlerService
	orchestrator *orchestratorclient.Client
}

func (s remoteAwareInjectionService) SubmitFaultInjection(ctx context.Context, req *injectionmodule.SubmitInjectionReq, groupID string, userID int, projectID *int) (*injectionmodule.SubmitInjectionResp, error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		return s.orchestrator.SubmitFaultInjection(ctx, req, groupID, userID, projectID)
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

func (s remoteAwareInjectionService) SubmitDatapackBuilding(ctx context.Context, req *injectionmodule.SubmitDatapackBuildingReq, groupID string, userID int, projectID *int) (*injectionmodule.SubmitDatapackBuildingResp, error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		return s.orchestrator.SubmitDatapackBuilding(ctx, req, groupID, userID, projectID)
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

type taskOrchestratorClient interface {
	Enabled() bool
	GetTask(context.Context, string) (*taskmodule.TaskDetailResp, error)
	PollTaskLogs(context.Context, string, time.Time) (*taskmodule.TaskLogPollResp, error)
	ListTasks(context.Context, *taskmodule.ListTaskReq) (*dto.ListResp[taskmodule.TaskResp], error)
}

type traceOrchestratorClient interface {
	Enabled() bool
	GetTrace(context.Context, string) (*tracemodule.TraceDetailResp, error)
	ListTraces(context.Context, *tracemodule.ListTraceReq) (*dto.ListResp[tracemodule.TraceResp], error)
	GetTraceStreamAlgorithms(context.Context, string) ([]dto.ContainerVersionItem, error)
	ReadTraceStreamMessages(context.Context, string, string, int64, time.Duration) ([]redis.XStream, error)
}

type remoteAwareTaskService struct {
	taskmodule.HandlerService
	orchestrator taskOrchestratorClient
}

func (s remoteAwareTaskService) GetDetail(ctx context.Context, taskID string) (*taskmodule.TaskDetailResp, error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		return s.orchestrator.GetTask(ctx, taskID)
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

func (s remoteAwareTaskService) List(ctx context.Context, req *taskmodule.ListTaskReq) (*dto.ListResp[taskmodule.TaskResp], error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		return s.orchestrator.ListTasks(ctx, req)
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

func (s remoteAwareTaskService) GetForLogStream(ctx context.Context, taskID string) (*model.Task, error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		if _, err := s.orchestrator.GetTask(ctx, taskID); err != nil {
			return nil, err
		}
		return &model.Task{ID: taskID}, nil
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

func (s remoteAwareTaskService) StreamLogs(ctx context.Context, conn *websocket.Conn, task *model.Task) {
	if s.orchestrator == nil || !s.orchestrator.Enabled() {
		writeTaskWSMessage(conn, taskmodule.WSLogMessage{
			Type:    consts.WSLogTypeError,
			Message: missingRemoteDependency("orchestrator-service").Error(),
		})
		_ = conn.Close()
		return
	}

	streamer := remoteTaskLogStreamer{
		conn:         conn,
		orchestrator: s.orchestrator,
		taskID:       task.ID,
	}
	streamer.stream(ctx)
}

const (
	remoteTaskLogWriteWait  = 10 * time.Second
	remoteTaskLogPongWait   = 60 * time.Second
	remoteTaskLogPingPeriod = 54 * time.Second
	remoteTaskLogMaxMsgSize = 512
	remoteTaskPollInterval  = time.Second
	remoteTaskFlushWindow   = 5 * time.Second
)

type remoteTaskLogStreamer struct {
	conn         *websocket.Conn
	orchestrator taskOrchestratorClient
	taskID       string
}

func (s remoteTaskLogStreamer) stream(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.conn.SetReadLimit(remoteTaskLogMaxMsgSize)
	_ = s.conn.SetReadDeadline(time.Now().Add(remoteTaskLogPongWait))
	s.conn.SetPongHandler(func(string) error {
		_ = s.conn.SetReadDeadline(time.Now().Add(remoteTaskLogPongWait))
		return nil
	})

	go s.readLoop(cancel)
	go s.pingLoop(ctx, cancel)

	initial, err := s.orchestrator.PollTaskLogs(ctx, s.taskID, time.Time{})
	if err != nil {
		writeTaskWSMessage(s.conn, taskmodule.WSLogMessage{
			Type:    consts.WSLogTypeError,
			Message: err.Error(),
		})
		_ = s.conn.Close()
		return
	}
	lastTimestamp := initial.CreatedAt
	if len(initial.Logs) > 0 {
		writeTaskWSMessage(s.conn, taskmodule.WSLogMessage{
			Type:  consts.WSLogTypeHistory,
			Logs:  initial.Logs,
			Total: len(initial.Logs),
		})
		lastTimestamp = initial.Logs[len(initial.Logs)-1].Timestamp
	}
	if initial.Terminal {
		s.flushTerminalLogs(ctx, lastTimestamp)
		return
	}

	ticker := time.NewTicker(remoteTaskPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			resp, err := s.orchestrator.PollTaskLogs(ctx, s.taskID, lastTimestamp)
			if err != nil {
				writeTaskWSMessage(s.conn, taskmodule.WSLogMessage{
					Type:    consts.WSLogTypeError,
					Message: err.Error(),
				})
				return
			}
			if len(resp.Logs) > 0 {
				writeTaskWSMessage(s.conn, taskmodule.WSLogMessage{
					Type: consts.WSLogTypeRealtime,
					Logs: resp.Logs,
				})
				lastTimestamp = resp.Logs[len(resp.Logs)-1].Timestamp
			}
			if resp.Terminal {
				s.flushTerminalLogs(ctx, lastTimestamp)
				return
			}
		}
	}
}

func (s remoteTaskLogStreamer) flushTerminalLogs(ctx context.Context, lastTimestamp time.Time) {
	deadline := time.Now().Add(remoteTaskFlushWindow)
	for time.Now().Before(deadline) {
		resp, err := s.orchestrator.PollTaskLogs(ctx, s.taskID, lastTimestamp)
		if err == nil && len(resp.Logs) > 0 {
			writeTaskWSMessage(s.conn, taskmodule.WSLogMessage{
				Type: consts.WSLogTypeRealtime,
				Logs: resp.Logs,
			})
			lastTimestamp = resp.Logs[len(resp.Logs)-1].Timestamp
		}
		time.Sleep(remoteTaskPollInterval)
	}
	writeTaskWSMessage(s.conn, taskmodule.WSLogMessage{
		Type:    consts.WSLogTypeEnd,
		Message: "task completed",
	})
	_ = s.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "task completed"), time.Now().Add(remoteTaskLogWriteWait))
}

func (s remoteTaskLogStreamer) readLoop(cancel context.CancelFunc) {
	defer cancel()
	for {
		if _, _, err := s.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (s remoteTaskLogStreamer) pingLoop(ctx context.Context, cancel context.CancelFunc) {
	ticker := time.NewTicker(remoteTaskLogPingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(remoteTaskLogWriteWait)); err != nil {
				cancel()
				return
			}
		}
	}
}

func writeTaskWSMessage(conn *websocket.Conn, msg taskmodule.WSLogMessage) {
	_ = conn.SetWriteDeadline(time.Now().Add(remoteTaskLogWriteWait))
	_ = conn.WriteJSON(msg)
}

type remoteAwareTraceService struct {
	tracemodule.HandlerService
	orchestrator traceOrchestratorClient
}

func (s remoteAwareTraceService) GetTrace(ctx context.Context, traceID string) (*tracemodule.TraceDetailResp, error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		return s.orchestrator.GetTrace(ctx, traceID)
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

func (s remoteAwareTraceService) ListTraces(ctx context.Context, req *tracemodule.ListTraceReq) (*dto.ListResp[tracemodule.TraceResp], error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		return s.orchestrator.ListTraces(ctx, req)
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

func (s remoteAwareTraceService) GetTraceStreamProcessor(ctx context.Context, traceID string) (*tracemodule.StreamProcessor, error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		algorithms, err := s.orchestrator.GetTraceStreamAlgorithms(ctx, traceID)
		if err != nil {
			return nil, err
		}
		return tracemodule.NewStreamProcessor(algorithms), nil
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

func (s remoteAwareTraceService) ReadTraceStreamMessages(ctx context.Context, streamKey, lastID string, count int64, block time.Duration) ([]redis.XStream, error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		return s.orchestrator.ReadTraceStreamMessages(ctx, streamKey, lastID, count, block)
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

type groupOrchestratorClient interface {
	Enabled() bool
	GetGroupStats(context.Context, string) (*groupmodule.GroupStats, error)
	GetGroupTraceCount(context.Context, string) (int, error)
	ReadGroupStreamMessages(context.Context, string, string, int64, time.Duration) ([]redis.XStream, error)
}

type remoteAwareGroupService struct {
	groupmodule.HandlerService
	orchestrator groupOrchestratorClient
}

func (s remoteAwareGroupService) GetGroupStats(ctx context.Context, req *groupmodule.GetGroupStatsReq) (*groupmodule.GroupStats, error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		return s.orchestrator.GetGroupStats(ctx, req.GroupID)
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

func (s remoteAwareGroupService) NewGroupStreamProcessor(ctx context.Context, groupID string) (*groupmodule.GroupStreamProcessor, error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		totalTraces, err := s.orchestrator.GetGroupTraceCount(ctx, groupID)
		if err != nil {
			return nil, err
		}
		return groupmodule.NewGroupStreamProcessor(totalTraces), nil
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

func (s remoteAwareGroupService) ReadGroupStreamMessages(ctx context.Context, streamKey, lastID string, count int64, block time.Duration) ([]redis.XStream, error) {
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		return s.orchestrator.ReadGroupStreamMessages(ctx, streamKey, lastID, count, block)
	}
	return nil, missingRemoteDependency("orchestrator-service")
}

type notificationOrchestratorClient interface {
	Enabled() bool
	ReadNotificationStreamMessages(context.Context, string, int64, time.Duration) ([]redis.XStream, error)
}

type remoteAwareNotificationService struct {
	notificationmodule.HandlerService
	orchestrator notificationOrchestratorClient
}

func (s remoteAwareNotificationService) ReadStreamMessages(ctx context.Context, streamKey, lastID string, count int64, block time.Duration) ([]redis.XStream, error) {
	_ = streamKey
	if s.orchestrator != nil && s.orchestrator.Enabled() {
		return s.orchestrator.ReadNotificationStreamMessages(ctx, lastID, count, block)
	}
	return nil, missingRemoteDependency("orchestrator-service")
}
