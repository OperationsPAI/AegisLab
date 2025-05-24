package dto

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/LGU-SE-Internal/rcabench/consts"
)

const ErrorStructList = "list"
const ErrorStructMap = "map"

// parseLookbackDuration parses a duration string with format like "5m", "2h", "1d"
// Supports: m (minutes), h (hours), d (days)
func parseLookbackDuration(lookback string) (time.Duration, error) {
	if lookback == "custom" {
		return 0, nil
	}

	// Use regex to match patterns like "5m", "2h", "1d"
	re := regexp.MustCompile(`^(\d+)([mhd])$`)
	matches := re.FindStringSubmatch(lookback)

	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid duration format: %s (expected format: 5m, 2h, 1d)", lookback)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid duration value: %s", matches[1])
	}

	unit := matches[2]
	switch unit {
	case "m":
		return time.Duration(value) * time.Minute, nil
	case "h":
		return time.Duration(value) * time.Hour, nil
	case "d":
		return time.Duration(value) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid duration unit: %s (supported: m, h, d)", unit)
	}
}

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

var ValidTaskEventMap = map[consts.TaskType][]consts.EventType{
	consts.TaskTypeBuildDataset: {
		consts.EventDatasetBuildSucceed,
	},
	consts.TaskTypeCollectResult: {
		consts.EventDatasetResultCollection,
		consts.EventDatasetNoAnomaly,
		consts.EventDatasetNoConclusionFile,
	},
	consts.TaskTypeFaultInjection: {
		consts.EventFaultInjectionStarted,
		consts.EventFaultInjectionCompleted,
		consts.EventFaultInjectionFailed,
	},
	consts.TaskTypeRunAlgorithm: {
		consts.EventAlgoRunSucceed,
	},
	consts.TaskTypeRestartService: {
		consts.EventNoNamespaceAvailable,
		consts.EventRestartServiceStarted,
		consts.EventRestartServiceCompleted,
		consts.EventRestartServiceFailed,
	},
}

var ValidTaskTypes = map[consts.TaskType]struct{}{
	consts.TaskTypeBuildDataset:   {},
	consts.TaskTypeRestartService: {},
	consts.TaskTypeRunAlgorithm:   {},
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

type TraceAnalyzeFilterOptions struct {
	FirstTaskType   consts.TaskType
	Lookback        time.Duration
	UseCustomRange  bool
	CustomStartTime time.Time
	CustomEndTime   time.Time
	ErrorStruct     string
}

type GetCompletedMapReq struct {
	Lookback       string `form:"lookback" binding:"omitempty"`
	CustomStartStr string `form:"custom_start_time" binding:"omitempty"`
	CustomEndStr   string `form:"custom_end_time" binding:"omitempty"`
}

func (req *GetCompletedMapReq) Validate() error {
	if req.Lookback != "" {
		// Try to parse the lookback duration using the flexible parser
		_, err := parseLookbackDuration(req.Lookback)
		if err != nil {
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

func (req *GetCompletedMapReq) Convert() (*TraceAnalyzeFilterOptions, error) {
	opts := &TraceAnalyzeFilterOptions{
		Lookback:        0,
		UseCustomRange:  false,
		CustomStartTime: time.Time{},
		CustomEndTime:   time.Time{},
	}

	if req.Lookback != "" {
		if req.Lookback == "custom" {
			customStart, err := time.Parse(time.RFC3339, req.CustomStartStr)
			if err != nil {
				return nil, fmt.Errorf("Invalid custom start time: %v", err)
			}

			customEnd, err := time.Parse(time.RFC3339, req.CustomEndStr)
			if err != nil {
				return nil, fmt.Errorf("Invalid custom end time: %v", err)
			}

			opts.UseCustomRange = true
			opts.CustomStartTime = customStart
			opts.CustomEndTime = customEnd
		} else {
			// Use the flexible duration parser
			duration, err := parseLookbackDuration(req.Lookback)
			if err != nil {
				return nil, fmt.Errorf("Invalid lookback value: %v", err)
			}
			opts.Lookback = duration
		}
	}

	return opts, nil
}

type TraceAnalyzeReq struct {
	FirstTaskType  string `form:"first_task_type" binding:"omitempty"`
	Lookback       string `form:"lookback" binding:"omitempty"`
	CustomStartStr string `form:"custom_start_time" binding:"omitempty"`
	CustomEndStr   string `form:"custom_end_time" binding:"omitempty"`
	ErrorStruct    string `form:"error_struct" binding:"omitempty"`
}

func (req *TraceAnalyzeReq) Validate() error {
	if req.FirstTaskType != "" {
		if _, exists := ValidTaskTypes[consts.TaskType(req.FirstTaskType)]; !exists {
			return fmt.Errorf("Invalid event name: %s", req.FirstTaskType)
		}
	}

	if req.Lookback != "" {
		// Try to parse the lookback duration using the flexible parser
		_, err := parseLookbackDuration(req.Lookback)
		if err != nil {
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

	if req.ErrorStruct != "" {
		if req.ErrorStruct != ErrorStructList && req.ErrorStruct != ErrorStructMap {
			return fmt.Errorf("Invalid error structure: %s", req.ErrorStruct)
		}
	}

	return nil
}

func (req *TraceAnalyzeReq) Convert() (*TraceAnalyzeFilterOptions, error) {
	opts := &TraceAnalyzeFilterOptions{
		FirstTaskType:   consts.TaskType(req.FirstTaskType),
		Lookback:        0,
		UseCustomRange:  false,
		CustomStartTime: time.Time{},
		CustomEndTime:   time.Time{},
		ErrorStruct:     ErrorStructMap,
	}

	if req.Lookback != "" {
		if req.Lookback == "custom" {
			customStart, err := time.Parse(time.RFC3339, req.CustomStartStr)
			if err != nil {
				return nil, fmt.Errorf("Invalid custom start time: %v", err)
			}

			customEnd, err := time.Parse(time.RFC3339, req.CustomEndStr)
			if err != nil {
				return nil, fmt.Errorf("Invalid custom end time: %v", err)
			}

			opts.UseCustomRange = true
			opts.CustomStartTime = customStart
			opts.CustomEndTime = customEnd
		} else {
			// Use the flexible duration parser
			duration, err := parseLookbackDuration(req.Lookback)
			if err != nil {
				return nil, fmt.Errorf("Invalid lookback value: %v", err)
			}
			opts.Lookback = duration
		}
	}

	if req.ErrorStruct != "" {
		opts.ErrorStruct = req.ErrorStruct
	}

	return opts, nil
}
