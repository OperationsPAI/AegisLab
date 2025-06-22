package dto

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type TimeRangeQuery struct {
	Lookback       string `form:"lookback" binding:"omitempty"`
	CustomStartStr string `form:"custom_start_time" binding:"omitempty"`
	CustomEndStr   string `form:"custom_end_time" binding:"omitempty"`
}

type TimeFilterOption struct {
	Lookback        time.Duration
	UseCustomRange  bool
	CustomStartTime time.Time
	CustomEndTime   time.Time
}

func (req *TimeRangeQuery) Convert() (*TimeFilterOption, error) {
	opts := &TimeFilterOption{
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
			duration, err := parseLookbackDuration(req.Lookback)
			if err != nil {
				return nil, fmt.Errorf("Invalid lookback value: %v", err)
			}

			opts.Lookback = duration
		}
	}

	return opts, nil
}

func (req *TimeRangeQuery) Validate() error {
	if req.Lookback != "" {
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

func (opts *TimeFilterOption) GetTimeRange() (time.Time, time.Time) {
	now := time.Now()
	var startTime, endTime time.Time
	if opts.UseCustomRange {
		startTime = opts.CustomStartTime
		endTime = opts.CustomEndTime
	} else if opts.Lookback != 0 {
		endTime = now
		startTime = now.Add(-opts.Lookback)
	} else {
		endTime = now
		startTime = time.Time{}
	}

	return startTime, endTime
}

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
