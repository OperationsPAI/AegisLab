package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"aegis/cmd/aegisctl/client"
)

// injectWaitEvent is a normalized view of a TraceStreamEvent, used by the
// wait loop and tests. Tests feed pre-canned sequences of these.
type injectWaitEvent struct {
	// SSEEvent is the raw SSE event type (e.g. "update", "end").
	// A blank SSEEvent with non-empty EventName represents an "update".
	SSEEvent string
	// EventName is the trace-stream `event_name` (e.g. "fault.injection.started").
	EventName string
	// Payload is the decoded payload. The shape depends on EventName:
	//   - fault.injection.started → string (CRD name)
	//   - datapack.build.succeed  → string
	//   - datapack.result.collection → map (with "datapack" / "job_name" keys)
	Payload any
}

// injectWaitResult is the structured summary emitted on exit.
type injectWaitResult struct {
	InjectionName   string  `json:"injection_name"`
	InjectionID     int     `json:"injection_id,omitempty"`
	TraceID         string  `json:"trace_id"`
	TraceState      string  `json:"trace_state"`
	DatapackID      int     `json:"datapack_id,omitempty"`
	DurationSeconds float64 `json:"duration_seconds"`

	// populated on failure/timeout for stderr diagnostics
	currentStage  string
	failureReason string
}

// waitUntil marker enum.
const (
	waitUntilInjectionCreated     = "injection_created"
	waitUntilFaultInjectionStart  = "fault_injection_started"
	waitUntilDatapackReady        = "datapack_ready"
	waitUntilFinished             = "finished"
)

func validWaitUntil(v string) bool {
	switch v {
	case "", waitUntilInjectionCreated, waitUntilFaultInjectionStart,
		waitUntilDatapackReady, waitUntilFinished:
		return true
	}
	return false
}

// terminalEventNames — consumer-emitted event_name values that mean the
// pipeline reached a terminal state. Sources: src/service/producer/trace.go
// (StreamProcessor.ProcessMessageForSSE) + src/consts/consts.go.
var terminalSuccessEvents = map[string]struct{}{
	"datapack.build.succeed":        {},
	"datapack.result.collection":    {},
	"datapack.no_anomaly":           {},
	"datapack.no_detector_data":     {},
	"algorithm.run.succeed":         {},
	"algorithm.result.collection":   {},
	"image.build.succeed":           {},
}

var terminalFailureEvents = map[string]struct{}{
	"restart.pedestal.failed":   {},
	"fault.injection.failed":    {},
	"datapack.build.failed":     {},
	"algorithm.run.failed":      {},
	"image.build.failed":        {},
}

// datapackReadyEvents — events that mark the datapack as built/ready.
var datapackReadyEvents = map[string]struct{}{
	"datapack.build.succeed":     {},
	"datapack.result.collection": {},
}

// runInjectWait consumes a channel of injectWaitEvents until a terminal
// state is observed, the wait-until event is seen, the context is cancelled
// (timeout), or the channel closes.
//
// Returns (result, exitCode, err).
//   exitCode 0 → Succeeded
//   exitCode 2 → Failed   (result.failureReason is set)
//   exitCode 3 → Timeout  (result.currentStage is set)
//   exitCode 1 → other (e.g. stream error)
func runInjectWait(
	ctx context.Context,
	events <-chan injectWaitEvent,
	errs <-chan error,
	traceID string,
	waitUntil string,
	start time.Time,
) (injectWaitResult, int, error) {
	res := injectWaitResult{TraceID: traceID}
	if waitUntil == "" {
		waitUntil = waitUntilFinished
	}

	finalize := func(state string) (injectWaitResult, int) {
		res.TraceState = state
		res.DurationSeconds = time.Since(start).Seconds()
		switch state {
		case "Succeeded":
			return res, 0
		case "Failed":
			return res, 2
		case "Timeout":
			return res, 3
		}
		return res, 1
	}

	for {
		select {
		case <-ctx.Done():
			if res.currentStage == "" {
				res.currentStage = "waiting"
			}
			r, code := finalize("Timeout")
			return r, code, nil

		case err := <-errs:
			if err == nil {
				continue
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				if res.currentStage == "" {
					res.currentStage = "waiting"
				}
				r, code := finalize("Timeout")
				return r, code, nil
			}
			return res, 1, err

		case evt, ok := <-events:
			if !ok {
				// Stream closed without a terminal event — treat as unknown; exit 1.
				if res.TraceState == "" {
					return res, 1, fmt.Errorf("event stream closed before terminal state")
				}
				r, code := finalize(res.TraceState)
				return r, code, nil
			}

			// Terminal SSE event: "end".
			if evt.SSEEvent == "end" {
				state := "Succeeded"
				if res.failureReason != "" {
					state = "Failed"
				}
				r, code := finalize(state)
				return r, code, nil
			}

			// Track injection_name on fault.injection.started.
			switch evt.EventName {
			case "fault.injection.started":
				if s, ok := evt.Payload.(string); ok && s != "" && res.InjectionName == "" {
					res.InjectionName = s
				}
				res.currentStage = "fault_injection"
				if waitUntil == waitUntilInjectionCreated ||
					waitUntil == waitUntilFaultInjectionStart {
					r, code := finalize("Succeeded")
					return r, code, nil
				}

			case "restart.pedestal.started":
				res.currentStage = "restart_pedestal"

			case "datapack.build.started":
				res.currentStage = "datapack_build"

			case "datapack.build.succeed", "datapack.result.collection":
				if res.currentStage != "algorithm" {
					res.currentStage = "datapack_ready"
				}
				// Try to extract datapack_id from payload, if present.
				if m, ok := evt.Payload.(map[string]any); ok {
					if v, ok := m["datapack_id"]; ok {
						switch n := v.(type) {
						case float64:
							res.DatapackID = int(n)
						case int:
							res.DatapackID = n
						}
					}
				}
				if waitUntil == waitUntilDatapackReady {
					r, code := finalize("Succeeded")
					return r, code, nil
				}

			case "algorithm.run.started":
				res.currentStage = "algorithm"

			case "algorithm.run.failed", "fault.injection.failed",
				"restart.pedestal.failed", "datapack.build.failed",
				"image.build.failed":
				reason := evt.EventName
				if s, ok := evt.Payload.(string); ok && s != "" {
					reason = fmt.Sprintf("%s: %s", evt.EventName, s)
				}
				res.failureReason = reason
				// Terminal: mark failure and return immediately (the "end"
				// SSE frame follows, but failure events are themselves terminal
				// per the backend StreamProcessor).
				r, code := finalize("Failed")
				return r, code, nil

			case "datapack.no_anomaly", "datapack.no_detector_data":
				// Terminal per StreamProcessor; treat as success for the
				// trace (the injection ran fine), but note no datapack.
				res.currentStage = "datapack_no_anomaly"
				r, code := finalize("Succeeded")
				return r, code, nil

			case "algorithm.result.collection", "algorithm.run.succeed":
				// Full pipeline — keep going; final "end" event terminates.
			}
		}
	}
}

// parseTraceSSEEvent converts a raw client.SSEEvent into an injectWaitEvent.
func parseTraceSSEEvent(e client.SSEEvent) injectWaitEvent {
	out := injectWaitEvent{SSEEvent: e.Event}
	if e.Event == "end" {
		return out
	}
	// "update" events carry a JSON-encoded TraceStreamEvent.
	var data struct {
		EventName string          `json:"event_name"`
		Payload   json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal([]byte(e.Data), &data); err != nil {
		return out
	}
	out.EventName = data.EventName
	if len(data.Payload) > 0 {
		// Try string first, then generic.
		var s string
		if err := json.Unmarshal(data.Payload, &s); err == nil {
			out.Payload = s
			return out
		}
		var generic any
		if err := json.Unmarshal(data.Payload, &generic); err == nil {
			out.Payload = generic
		}
	}
	return out
}
