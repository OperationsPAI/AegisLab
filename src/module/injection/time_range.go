package injection

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type TimeRangeQuery struct {
	Lookback       string `form:"lookback" binding:"omitempty"`
	CustomStartStr string `form:"custom_start_time" binding:"omitempty"`
	CustomEndStr   string `form:"custom_end_time" binding:"omitempty"`
}

type TimeFilterOptions struct {
	Lookback        time.Duration
	UseCustomRange  bool
	CustomStartTime time.Time
	CustomEndTime   time.Time
}

func (req *TimeRangeQuery) Convert() (*TimeFilterOptions, error) {
	opts := &TimeFilterOptions{}
	if req.Lookback != "custom" {
		duration, err := parseLookbackDuration(req.Lookback)
		if err != nil {
			return nil, fmt.Errorf("invalid lookback value: %v", err)
		}
		opts.Lookback = duration
		return opts, nil
	}

	customStart, err := time.Parse(time.RFC3339, req.CustomStartStr)
	if err != nil {
		return nil, fmt.Errorf("invalid custom start time: %v", err)
	}
	customEnd, err := time.Parse(time.RFC3339, req.CustomEndStr)
	if err != nil {
		return nil, fmt.Errorf("invalid custom end time: %v", err)
	}

	opts.UseCustomRange = true
	opts.CustomStartTime = customStart
	opts.CustomEndTime = customEnd
	return opts, nil
}

func (req *TimeRangeQuery) Validate() error {
	if req.Lookback != "custom" {
		if _, err := parseLookbackDuration(req.Lookback); err != nil {
			return fmt.Errorf("invalid lookback value: %s", req.Lookback)
		}
		return nil
	}

	if req.CustomStartStr == "" || req.CustomEndStr == "" {
		return fmt.Errorf("custom start and end times are required for custom lookback")
	}

	startTime, err := time.Parse(time.RFC3339, req.CustomStartStr)
	if err != nil {
		return fmt.Errorf("invalid custom start time: %v", err)
	}
	endTime, err := time.Parse(time.RFC3339, req.CustomEndStr)
	if err != nil {
		return fmt.Errorf("invalid custom end time: %v", err)
	}
	if startTime.After(endTime) {
		return fmt.Errorf("custom start time cannot be after custom end time")
	}
	return nil
}

func (opts *TimeFilterOptions) GetTimeRange() (time.Time, time.Time) {
	now := time.Now()
	if opts.UseCustomRange {
		return opts.CustomStartTime, opts.CustomEndTime
	}
	if opts.Lookback != 0 {
		return now.Add(-opts.Lookback), now
	}
	return time.Time{}, now
}

func (opts *TimeFilterOptions) AddTimeFilter(query *gorm.DB, column string) *gorm.DB {
	startTime, endTime := opts.GetTimeRange()
	return query.Where(fmt.Sprintf("%s >= ? AND %s <= ?", column, column), startTime, endTime)
}

func parseLookbackDuration(lookback string) (time.Duration, error) {
	if lookback == "" {
		return 0, nil
	}

	re := regexp.MustCompile(`^(\d+)([mhd])$`)
	matches := re.FindStringSubmatch(lookback)
	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid duration format: %s (expected format: 5m, 2h, 1d)", lookback)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid duration value: %s", matches[1])
	}
	if value <= 0 {
		return 0, fmt.Errorf("duration value must be a positive integer: %s", matches[1])
	}

	switch matches[2] {
	case "m":
		return time.Duration(value) * time.Minute, nil
	case "h":
		return time.Duration(value) * time.Hour, nil
	case "d":
		return time.Duration(value) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid duration unit: %s (supported: m, h, d)", matches[2])
	}
}
