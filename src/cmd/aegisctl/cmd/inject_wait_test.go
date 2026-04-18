package cmd

import (
	"context"
	"testing"
	"time"

	"aegis/cmd/aegisctl/client"
)

func struct_SSE(event, data string) client.SSEEvent {
	return client.SSEEvent{Event: event, Data: data}
}

// feed returns closed event and error channels pre-loaded with evts.
func feed(evts []injectWaitEvent) (<-chan injectWaitEvent, <-chan error) {
	ec := make(chan injectWaitEvent, len(evts)+1)
	errC := make(chan error, 1)
	for _, e := range evts {
		ec <- e
	}
	close(ec)
	return ec, errC
}

func TestRunInjectWait_SucceededFullPipeline(t *testing.T) {
	evts := []injectWaitEvent{
		{EventName: "fault.injection.started", Payload: "otel-demo0-checkout-delay-p7qd5c"},
		{EventName: "datapack.build.started"},
		{EventName: "datapack.build.succeed", Payload: "foo"},
		{EventName: "algorithm.run.started"},
		{EventName: "algorithm.run.succeed", Payload: "baro"},
		{SSEEvent: "end"},
	}
	events, errs := feed(evts)
	res, code, err := runInjectWait(context.Background(), events, errs, "trace-1", "", time.Now())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if code != 0 {
		t.Fatalf("want exit 0, got %d (state=%s)", code, res.TraceState)
	}
	if res.InjectionName != "otel-demo0-checkout-delay-p7qd5c" {
		t.Errorf("injection_name not resolved: %+v", res)
	}
	if res.TraceState != "Succeeded" {
		t.Errorf("want Succeeded, got %s", res.TraceState)
	}
}

func TestRunInjectWait_FailedFaultInjection(t *testing.T) {
	evts := []injectWaitEvent{
		{EventName: "fault.injection.started", Payload: "name-x"},
		{EventName: "fault.injection.failed", Payload: "CRD timeout"},
	}
	events, errs := feed(evts)
	res, code, _ := runInjectWait(context.Background(), events, errs, "trace-2", "", time.Now())
	if code != 2 {
		t.Fatalf("want exit 2, got %d", code)
	}
	if res.TraceState != "Failed" {
		t.Errorf("want Failed, got %s", res.TraceState)
	}
	if res.failureReason == "" {
		t.Errorf("expected failure reason to be set")
	}
}

func TestRunInjectWait_Timeout(t *testing.T) {
	// Empty stream, never sends; ctx times out.
	ec := make(chan injectWaitEvent)
	errC := make(chan error)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	res, code, _ := runInjectWait(ctx, ec, errC, "trace-3", "", time.Now())
	if code != 3 {
		t.Fatalf("want exit 3, got %d", code)
	}
	if res.TraceState != "Timeout" {
		t.Errorf("want Timeout, got %s", res.TraceState)
	}
}

func TestRunInjectWait_WaitUntilInjectionCreated(t *testing.T) {
	evts := []injectWaitEvent{
		{EventName: "restart.pedestal.started"},
		{EventName: "fault.injection.started", Payload: "inj-abc"},
		// More events follow; should not be consumed.
		{EventName: "datapack.build.started"},
	}
	events, errs := feed(evts)
	res, code, _ := runInjectWait(context.Background(), events, errs, "trace-4", "injection_created", time.Now())
	if code != 0 {
		t.Fatalf("want exit 0, got %d", code)
	}
	if res.InjectionName != "inj-abc" {
		t.Errorf("injection_name=%q", res.InjectionName)
	}
}

func TestRunInjectWait_WaitUntilDatapackReady(t *testing.T) {
	evts := []injectWaitEvent{
		{EventName: "fault.injection.started", Payload: "inj-y"},
		{EventName: "datapack.build.succeed", Payload: map[string]any{"datapack_id": float64(17)}},
	}
	events, errs := feed(evts)
	res, code, _ := runInjectWait(context.Background(), events, errs, "trace-5", "datapack_ready", time.Now())
	if code != 0 {
		t.Fatalf("want exit 0, got %d", code)
	}
	if res.DatapackID != 17 {
		t.Errorf("datapack_id=%d, want 17", res.DatapackID)
	}
}

func TestRunInjectWait_NoAnomalyIsTerminalSuccess(t *testing.T) {
	evts := []injectWaitEvent{
		{EventName: "fault.injection.started", Payload: "inj-z"},
		{EventName: "datapack.no_anomaly"},
	}
	events, errs := feed(evts)
	_, code, _ := runInjectWait(context.Background(), events, errs, "trace-6", "", time.Now())
	if code != 0 {
		t.Fatalf("want exit 0, got %d", code)
	}
}

func TestValidWaitUntil(t *testing.T) {
	for _, ok := range []string{"", "injection_created", "fault_injection_started", "datapack_ready", "finished"} {
		if !validWaitUntil(ok) {
			t.Errorf("%q should be valid", ok)
		}
	}
	if validWaitUntil("bogus") {
		t.Errorf("bogus should be invalid")
	}
}

func TestParseTraceSSEEvent(t *testing.T) {
	raw := `{"timestamp":0,"task_id":"t","task_type":1,"event_name":"fault.injection.started","payload":"crd-name-xyz"}`
	evt := parseTraceSSEEvent(struct_SSE("update", raw))
	if evt.EventName != "fault.injection.started" {
		t.Errorf("event_name=%q", evt.EventName)
	}
	if s, _ := evt.Payload.(string); s != "crd-name-xyz" {
		t.Errorf("payload=%v", evt.Payload)
	}

	end := parseTraceSSEEvent(struct_SSE("end", ""))
	if end.SSEEvent != "end" {
		t.Errorf("expected end, got %+v", end)
	}
}
