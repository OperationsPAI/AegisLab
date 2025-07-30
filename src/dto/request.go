package dto

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ListOptionsQuery struct {
	SortField string `form:"sort_field" bindging:"omitempty"`
	SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
	Limit     int    `form:"limit" binding:"omitempty"`
}

func (req *ListOptionsQuery) setDefaults() {
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	if req.SortField == "" {
		req.SortField = "created_at"
	}
}

func (req *ListOptionsQuery) Validate() error {
	req.setDefaults()

	if req.Limit < 0 {
		return fmt.Errorf("Limit must be a non-negative integer")
	}

	return nil
}

var ValidPageSizeMap = map[int]struct{}{
	10: {},
	20: {},
	50: {},
}

func getValidPageSizes() string {
	sizes := make([]string, 0, len(ValidPageSizeMap))
	for size := range ValidPageSizeMap {
		sizes = append(sizes, fmt.Sprintf("%d", size))
	}

	return strings.Join(sizes, ", ")
}

type PaginationQuery struct {
	PageNum  int `form:"page_num" binding:"omitempty"`
	PageSize int `form:"page_size" binding:"omitempty"`
}

func (req *PaginationQuery) Validate() error {
	if req.PageNum < 0 {
		return fmt.Errorf("Page number must be a non-negative integer")
	}

	if req.PageSize < 0 {
		return fmt.Errorf("Page size must be a non-negative integer")
	}

	if (req.PageNum == 0) != (req.PageSize == 0) {
		return fmt.Errorf("Both page_num and page_size must be provided together or both be 0")
	}

	if req.PageSize > 0 {
		if _, exists := ValidPageSizeMap[req.PageSize]; !exists {
			return fmt.Errorf("Invalid page size: %d (supported: %s)", req.PageSize, getValidPageSizes())
		}
	}

	return nil
}

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
	opts := &TimeFilterOptions{
		Lookback:        0,
		UseCustomRange:  false,
		CustomStartTime: time.Time{},
		CustomEndTime:   time.Time{},
	}

	if req.Lookback != "custom" {
		duration, err := parseLookbackDuration(req.Lookback)
		if err != nil {
			return nil, fmt.Errorf("Invalid lookback value: %v", err)
		}

		opts.Lookback = duration
	} else {
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
	}

	return opts, nil
}

func (req *TimeRangeQuery) Validate() error {
	if req.Lookback != "custom" {
		if _, err := parseLookbackDuration(req.Lookback); err != nil {
			return fmt.Errorf("Invalid lookback value: %s", req.Lookback)
		}
	} else {
		if req.CustomStartStr == "" || req.CustomEndStr == "" {
			return fmt.Errorf("Custom start and end times are required for custom lookback")
		}

		startTime, err := time.Parse(time.RFC3339, req.CustomStartStr)
		if err != nil {
			return fmt.Errorf("Invalid custom start time: %v", err)
		}

		endTime, err := time.Parse(time.RFC3339, req.CustomEndStr)
		if err != nil {
			return fmt.Errorf("Invalid custom end time: %v", err)
		}

		if startTime.After(endTime) {
			return fmt.Errorf("Custom start time cannot be after custom end time")
		}
	}

	return nil
}

func (opts *TimeFilterOptions) GetTimeRange() (time.Time, time.Time) {
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

func (opts *TimeFilterOptions) AddTimeFilter(query *gorm.DB, column string) *gorm.DB {
	startTime, endTime := opts.GetTimeRange()
	return query.Where(fmt.Sprintf("%s >= ? AND %s <= ?", column, column), startTime, endTime)
}

// parseLookbackDuration parses a duration string with format like "5m", "2h", "1d"
// Supports: m (minutes), h (hours), d (days)
func parseLookbackDuration(lookback string) (time.Duration, error) {
	if lookback == "" {
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

	if value <= 0 {
		return 0, fmt.Errorf("duration value must be a positive integer: %s", matches[1])
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
