package dto

import (
	"fmt"
	"time"

	"github.com/LGU-SE-Internal/rcabench/consts"
)

const (
	LookbackFiveMinutes = 5 * time.Minute
	LookbackFifteenMin  = 15 * time.Minute
	LookbackThirtyMin   = 30 * time.Minute
	LookbackOneHour     = 1 * time.Hour
	LookbackTwoHours    = 2 * time.Hour
	LookbackThreeHours  = 3 * time.Hour
	LookbackSixHours    = 6 * time.Hour
	LookbackTwelveHours = 12 * time.Hour
	LookbackOneDay      = 24 * time.Hour
	LookbackTweDay      = 48 * time.Hour
)

var ValidAnaylzeEventNames = map[consts.EventType]struct{}{
	consts.EventAlgoRunSucceed: {},

	consts.EventDatasetResultCollection: {},
	consts.EventDatasetNoAnomaly:        {},
	consts.EventDatasetNoConclusionFile: {},
	consts.EventDatasetBuildSucceed:     {},

	consts.EventTaskRetryStatus: {},
	consts.EventTaskStarted:     {},

	consts.EventNoNamespaceAvailable:    {},
	consts.EventRestartServiceStarted:   {},
	consts.EventRestartServiceCompleted: {},
	consts.EventRestartServiceFailed:    {},

	consts.EventFaultInjectionStarted:   {},
	consts.EventFaultInjectionCompleted: {},
	consts.EventFaultInjectionFailed:    {},
}

var ValidLookbackValues = map[string]time.Duration{
	"5m":     LookbackFiveMinutes,
	"15m":    LookbackFifteenMin,
	"30m":    LookbackThirtyMin,
	"1h":     LookbackOneHour,
	"2h":     LookbackTwoHours,
	"3h":     LookbackThreeHours,
	"6h":     LookbackSixHours,
	"12h":    LookbackTwelveHours,
	"1d":     LookbackOneDay,
	"2d":     LookbackTweDay,
	"custom": 0,
}

type TraceReq struct {
	TraceID string `uri:"trace_id" binding:"required"`
}

type TraceStreamReq struct {
	LastID string `bind:"last_event_id"`
}

type TraceAnalyzeReq struct {
	EventName      string `form:"event_name" binding:"omitempty"`
	Lookback       string `form:"lookback" binding:"omitempty"`
	CustomStartStr string `form:"custom_start_time" binding:"omitempty"`
	CustomEndStr   string `form:"custom_end_time" binding:"omitempty"`
}

func (req *TraceAnalyzeReq) Validate() error {
	if req.EventName != "" {
		if _, exists := ValidAnaylzeEventNames[consts.EventType(req.EventName)]; !exists {
			return fmt.Errorf("Invalid event name: %s", req.EventName)
		}
	}

	if req.Lookback != "" {
		if _, exists := ValidLookbackValues[req.Lookback]; !exists {
			return fmt.Errorf("Invalid lookback value: %s", req.Lookback)
		}

		if req.Lookback == "custom" {
			if req.CustomStartStr == "" || req.CustomEndStr == "" {
				return fmt.Errorf("Custom start and end times are required for custom lookback")
			}

			if _, err := time.Parse(time.RFC3339, req.CustomStartStr); err != nil {
				return fmt.Errorf("Invalid custom start time: %v", err)
			}

			if _, err := time.Parse(time.RFC3339, req.CustomEndStr); err != nil {
				return fmt.Errorf("Invalid custom end time: %v", err)
			}
		}
	}

	return nil
}

type TraceAnalyzeFilterOptions struct {
	EventName       consts.EventType
	Lookback        time.Duration
	UseCustomRange  bool
	CustomStartTime time.Time
	CustomEndTime   time.Time
}

func (opts *TraceAnalyzeFilterOptions) Convert(req TraceAnalyzeReq) error {
	opts.EventName = consts.EventType("")
	opts.Lookback = 0
	opts.UseCustomRange = false
	opts.CustomStartTime = time.Time{}
	opts.CustomEndTime = time.Time{}

	opts.EventName = consts.EventType(req.EventName)

	if req.Lookback != "" {
		duration := ValidLookbackValues[req.Lookback]
		if req.Lookback == "custom" {
			customStart, err := time.Parse(time.RFC3339, req.CustomStartStr)
			if err != nil {
				return fmt.Errorf("Invalid custom start time: %v", err)
			}

			customEnd, err := time.Parse(time.RFC3339, req.CustomEndStr)
			if err != nil {
				return fmt.Errorf("Invalid custom end time: %v", err)
			}

			opts.UseCustomRange = true
			opts.CustomStartTime = customStart
			opts.CustomEndTime = customEnd
		} else {
			opts.Lookback = duration
		}
	}

	return nil
}
