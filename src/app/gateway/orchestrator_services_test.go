package gatewayapp

import (
	"context"
	"testing"
	"time"

	"aegis/dto"
	groupmodule "aegis/module/group"
	taskmodule "aegis/module/task"
	tracemodule "aegis/module/trace"

	"github.com/redis/go-redis/v9"
)

type orchestratorTaskClientStub struct {
	enabled bool
}

func (s *orchestratorTaskClientStub) Enabled() bool { return s.enabled }

func (s *orchestratorTaskClientStub) GetTask(context.Context, string) (*taskmodule.TaskDetailResp, error) {
	return &taskmodule.TaskDetailResp{TaskResp: taskmodule.TaskResp{ID: "task-1"}}, nil
}

func (s *orchestratorTaskClientStub) PollTaskLogs(context.Context, string, time.Time) (*taskmodule.TaskLogPollResp, error) {
	return &taskmodule.TaskLogPollResp{
		Logs:      []dto.LogEntry{{TaskID: "task-1", Line: "hello"}},
		Terminal:  true,
		State:     "completed",
		CreatedAt: time.Unix(1710000000, 0),
	}, nil
}

func (s *orchestratorTaskClientStub) ListTasks(context.Context, *taskmodule.ListTaskReq) (*dto.ListResp[taskmodule.TaskResp], error) {
	return &dto.ListResp[taskmodule.TaskResp]{Items: []taskmodule.TaskResp{{ID: "task-1"}}}, nil
}

type orchestratorTraceClientStub struct {
	enabled bool
}

func (s *orchestratorTraceClientStub) Enabled() bool { return s.enabled }

func (s *orchestratorTraceClientStub) GetTrace(context.Context, string) (*tracemodule.TraceDetailResp, error) {
	return &tracemodule.TraceDetailResp{TraceResp: tracemodule.TraceResp{ID: "trace-1"}}, nil
}

func (s *orchestratorTraceClientStub) ListTraces(context.Context, *tracemodule.ListTraceReq) (*dto.ListResp[tracemodule.TraceResp], error) {
	return &dto.ListResp[tracemodule.TraceResp]{Items: []tracemodule.TraceResp{{ID: "trace-1"}}}, nil
}

func (s *orchestratorTraceClientStub) GetTraceStreamAlgorithms(context.Context, string) ([]dto.ContainerVersionItem, error) {
	return []dto.ContainerVersionItem{{ContainerName: "algo-a"}}, nil
}

func (s *orchestratorTraceClientStub) ReadTraceStreamMessages(context.Context, string, string, int64, time.Duration) ([]redis.XStream, error) {
	return []redis.XStream{{Stream: "trace:trace-1:log"}}, nil
}

type orchestratorGroupClientStub struct {
	enabled bool
}

func (s *orchestratorGroupClientStub) Enabled() bool { return s.enabled }

func (s *orchestratorGroupClientStub) GetGroupStats(context.Context, string) (*groupmodule.GroupStats, error) {
	return &groupmodule.GroupStats{TotalTraces: 2}, nil
}

func (s *orchestratorGroupClientStub) GetGroupTraceCount(context.Context, string) (int, error) {
	return 2, nil
}

func (s *orchestratorGroupClientStub) ReadGroupStreamMessages(context.Context, string, string, int64, time.Duration) ([]redis.XStream, error) {
	return []redis.XStream{{Stream: "group:group-1:log"}}, nil
}

type orchestratorNotificationClientStub struct {
	enabled bool
}

func (s *orchestratorNotificationClientStub) Enabled() bool { return s.enabled }

func (s *orchestratorNotificationClientStub) ReadNotificationStreamMessages(context.Context, string, int64, time.Duration) ([]redis.XStream, error) {
	return []redis.XStream{{Stream: "notifications:global"}}, nil
}

func TestRemoteAwareTaskServiceRequiresOrchestrator(t *testing.T) {
	service := remoteAwareTaskService{}
	if _, err := service.List(context.Background(), &taskmodule.ListTaskReq{}); err == nil {
		t.Fatal("List() error = nil, want missing dependency")
	}
}

func TestRemoteAwareTaskServiceUsesOrchestratorClient(t *testing.T) {
	service := remoteAwareTaskService{orchestrator: &orchestratorTaskClientStub{enabled: true}}
	resp, err := service.GetDetail(context.Background(), "task-1")
	if err != nil {
		t.Fatalf("GetDetail() error = %v", err)
	}
	if resp.ID != "task-1" {
		t.Fatalf("GetDetail() unexpected response: %+v", resp)
	}

	task, err := service.GetForLogStream(context.Background(), "task-1")
	if err != nil {
		t.Fatalf("GetForLogStream() error = %v", err)
	}
	if task.ID != "task-1" {
		t.Fatalf("GetForLogStream() unexpected response: %+v", task)
	}
}

func TestRemoteAwareTraceServiceRequiresOrchestrator(t *testing.T) {
	service := remoteAwareTraceService{}
	if _, err := service.ListTraces(context.Background(), &tracemodule.ListTraceReq{}); err == nil {
		t.Fatal("ListTraces() error = nil, want missing dependency")
	}
}

func TestRemoteAwareTraceServiceUsesOrchestratorClient(t *testing.T) {
	service := remoteAwareTraceService{orchestrator: &orchestratorTraceClientStub{enabled: true}}
	resp, err := service.GetTrace(context.Background(), "trace-1")
	if err != nil {
		t.Fatalf("GetTrace() error = %v", err)
	}
	if resp.ID != "trace-1" {
		t.Fatalf("GetTrace() unexpected response: %+v", resp)
	}

	processor, err := service.GetTraceStreamProcessor(context.Background(), "trace-1")
	if err != nil {
		t.Fatalf("GetTraceStreamProcessor() error = %v", err)
	}
	if processor == nil {
		t.Fatal("GetTraceStreamProcessor() = nil")
	}
}

func TestRemoteAwareGroupServiceRequiresOrchestrator(t *testing.T) {
	service := remoteAwareGroupService{}
	if _, err := service.GetGroupStats(context.Background(), &groupmodule.GetGroupStatsReq{
		GroupID: "d7a4ed4b-1c91-4cdb-8af8-5520fa8d0ce0",
	}); err == nil {
		t.Fatal("GetGroupStats() error = nil, want missing dependency")
	}
}

func TestRemoteAwareGroupServiceUsesOrchestratorClient(t *testing.T) {
	service := remoteAwareGroupService{orchestrator: &orchestratorGroupClientStub{enabled: true}}
	resp, err := service.GetGroupStats(context.Background(), &groupmodule.GetGroupStatsReq{
		GroupID: "d7a4ed4b-1c91-4cdb-8af8-5520fa8d0ce0",
	})
	if err != nil {
		t.Fatalf("GetGroupStats() error = %v", err)
	}
	if resp.TotalTraces != 2 {
		t.Fatalf("GetGroupStats() unexpected response: %+v", resp)
	}

	processor, err := service.NewGroupStreamProcessor(context.Background(), "group-1")
	if err != nil {
		t.Fatalf("NewGroupStreamProcessor() error = %v", err)
	}
	if processor == nil {
		t.Fatal("NewGroupStreamProcessor() = nil")
	}
}

func TestRemoteAwareNotificationServiceRequiresOrchestrator(t *testing.T) {
	service := remoteAwareNotificationService{}
	if _, err := service.ReadStreamMessages(context.Background(), "notifications:global", "0", 10, time.Second); err == nil {
		t.Fatal("ReadStreamMessages() error = nil, want missing dependency")
	}
}

func TestRemoteAwareNotificationServiceUsesOrchestratorClient(t *testing.T) {
	service := remoteAwareNotificationService{orchestrator: &orchestratorNotificationClientStub{enabled: true}}
	resp, err := service.ReadStreamMessages(context.Background(), "notifications:global", "0", 10, time.Second)
	if err != nil {
		t.Fatalf("ReadStreamMessages() error = %v", err)
	}
	if len(resp) != 1 || resp[0].Stream != "notifications:global" {
		t.Fatalf("ReadStreamMessages() unexpected response: %+v", resp)
	}
}
