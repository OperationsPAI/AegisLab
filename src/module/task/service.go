package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"aegis/consts"
	"aegis/dto"
	redisinfra "aegis/infra/redis"
	"aegis/model"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service struct {
	repository *Repository
	logService *TaskLogService
	loki       *LokiGateway
	redis      *redisinfra.Gateway
}

func NewService(repository *Repository, logService *TaskLogService, loki *LokiGateway, redis *redisinfra.Gateway) *Service {
	return &Service{
		repository: repository,
		logService: logService,
		loki:       loki,
		redis:      redis,
	}
}

func (s *Service) BatchDelete(ctx context.Context, taskIDs []string) error {
	if len(taskIDs) == 0 {
		return nil
	}

	return s.repository.BatchDelete(taskIDs)
}

func (s *Service) GetDetail(ctx context.Context, taskID string) (*TaskDetailResp, error) {
	task, err := s.repository.GetByID(taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: task id: %s", consts.ErrNotFound, taskID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	logs := s.queryHistoricalLogs(ctx, task)
	return NewTaskDetailResp(task, logs), nil
}

func (s *Service) List(ctx context.Context, req *ListTaskReq) (*dto.ListResp[TaskResp], error) {
	if req == nil {
		return nil, fmt.Errorf("list tasks request is nil")
	}

	limit, offset := req.ToGormParams()
	filterOptions := req.ToFilterOptions()

	tasks, total, err := s.repository.List(limit, offset, filterOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	taskResps := make([]TaskResp, 0, len(tasks))
	for _, task := range tasks {
		taskResps = append(taskResps, *NewTaskResp(&task))
	}

	return &dto.ListResp[TaskResp]{
		Items:      taskResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

// Expedite moves a Pending task's execute_time to now.
// Contract:
//   - task not found → wrapped consts.ErrNotFound
//   - task not Pending → wrapped consts.ErrBadRequest
//   - already due → no-op, returns task resp (idempotent)
//
// DB update is authoritative; Redis rescore is best-effort — if the entry
// is already promoted by the scheduler, the call still succeeds.
func (s *Service) Expedite(ctx context.Context, taskID string) (*TaskResp, error) {
	task, err := s.repository.GetByID(taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: task id: %s", consts.ErrNotFound, taskID)
		}
		return nil, fmt.Errorf("failed to load task: %w", err)
	}

	if task.State != consts.TaskPending {
		return nil, fmt.Errorf("%w: state=%s, cannot expedite",
			consts.ErrBadRequest, consts.GetTaskStateName(task.State))
	}

	now := time.Now().Unix()
	if task.ExecuteTime <= now {
		return NewTaskResp(task), nil
	}

	if err := s.repository.UpdateExecuteTime(taskID, now); err != nil {
		return nil, fmt.Errorf("failed to update execute_time: %w", err)
	}

	if _, err := s.redis.ExpediteDelayedTask(ctx, taskID, now); err != nil {
		logrus.WithField("task_id", taskID).
			Warnf("DB updated but Redis rescore failed: %v", err)
	}

	s.emitExpediteScheduledEvent(ctx, task, now)

	task.ExecuteTime = now
	return NewTaskResp(task), nil
}

// emitExpediteScheduledEvent publishes a task.scheduled event for a manually
// expedited task. Best-effort — failures are logged only.
func (s *Service) emitExpediteScheduledEvent(ctx context.Context, task *model.Task, executeTime int64) {
	if task == nil || task.TraceID == "" || s.redis == nil {
		return
	}
	event := dto.TraceStreamEvent{
		TaskID:    task.ID,
		TaskType:  task.Type,
		EventName: consts.EventTaskScheduled,
		Payload: dto.TaskScheduledPayload{
			ExecuteTime: executeTime,
			Reason:      dto.TaskScheduledReasonExpedite,
		},
	}
	stream := fmt.Sprintf(consts.StreamTraceLogKey, task.TraceID)
	if err := s.redis.XAdd(ctx, stream, event.ToRedisStream()); err != nil {
		logrus.WithField("task_id", task.ID).
			Warnf("failed to emit expedite task.scheduled event: %v", err)
	}
}

func (s *Service) GetForLogStream(ctx context.Context, taskID string) (*model.Task, error) {
	task, err := s.repository.GetByID(taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: task id: %s", consts.ErrNotFound, taskID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

func (s *Service) StreamLogs(ctx context.Context, conn *websocket.Conn, task *model.Task) {
	s.logService.StreamLogs(ctx, conn, task)
}

func (s *Service) PollLogs(ctx context.Context, taskID string, after time.Time) (*TaskLogPollResp, error) {
	task, err := s.repository.GetByID(taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: task id: %s", consts.ErrNotFound, taskID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	start := task.CreatedAt
	if !after.IsZero() && after.After(start) {
		start = after.Add(time.Nanosecond)
	}

	lokiCtx, lokiCancel := context.WithTimeout(ctx, 10*time.Second)
	defer lokiCancel()

	logEntries, err := s.loki.QueryJobLogs(lokiCtx, task.ID, start)
	if err != nil {
		return nil, fmt.Errorf("failed to query task logs: %w", err)
	}

	return &TaskLogPollResp{
		Logs:      logEntries,
		Terminal:  isTaskTerminal(task.State),
		State:     consts.GetTaskStateName(task.State),
		CreatedAt: task.CreatedAt,
	}, nil
}

func (s *Service) queryHistoricalLogs(ctx context.Context, task *model.Task) []string {
	lokiCtx, lokiCancel := context.WithTimeout(ctx, 10*time.Second)
	defer lokiCancel()

	logEntries, err := s.loki.QueryJobLogs(lokiCtx, task.ID, task.CreatedAt)
	if err != nil {
		logrus.Warnf("Failed to query Loki for task %s logs: %v", task.ID, err)
		return []string{}
	}

	logs := make([]string, 0, len(logEntries))
	for _, entry := range logEntries {
		logs = append(logs, entry.Line)
	}
	return logs
}
