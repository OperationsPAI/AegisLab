package grpcruntimeinterface

import (
	"context"
	"testing"
	"time"

	"aegis/consts"
	runtimev1 "aegis/proto/runtime/v1"
	"aegis/service/consumer"
)

func TestRuntimeServerStatusEndpoints(t *testing.T) {
	originalStart := consts.InitialTime
	originalAppID := consts.AppID
	startedAt := time.Unix(1_700_000_000, 0)
	consts.InitialTime = &startedAt
	consts.AppID = "app-test"
	t.Cleanup(func() {
		consts.InitialTime = originalStart
		consts.AppID = originalAppID
	})

	server := &runtimeServer{
		snapshots: consumer.NewRuntimeSnapshotService(nil, nil, nil, nil, nil, nil, nil, nil),
	}

	pingResp, err := server.Ping(context.Background(), &runtimev1.PingRequest{})
	if err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if pingResp.Service != consumer.RuntimeServiceName {
		t.Fatalf("Ping() service = %q, want %q", pingResp.Service, consumer.RuntimeServiceName)
	}
	if pingResp.AppId != "app-test" {
		t.Fatalf("Ping() app id = %q, want %q", pingResp.AppId, "app-test")
	}

	statusResp, err := server.GetRuntimeStatus(context.Background(), &runtimev1.RuntimeStatusRequest{})
	if err != nil {
		t.Fatalf("GetRuntimeStatus() error = %v", err)
	}
	if statusResp.Service != consumer.RuntimeServiceName {
		t.Fatalf("GetRuntimeStatus() service = %q, want %q", statusResp.Service, consumer.RuntimeServiceName)
	}
	if statusResp.Mode != "runtime-worker" {
		t.Fatalf("GetRuntimeStatus() mode = %q, want %q", statusResp.Mode, "runtime-worker")
	}
	if statusResp.DbAvailable || statusResp.RedisAvailable || statusResp.K8SAvailable || statusResp.BuildkitAvailable || statusResp.HelmAvailable {
		t.Fatalf("GetRuntimeStatus() unexpected dependency availability: %+v", statusResp)
	}
}
