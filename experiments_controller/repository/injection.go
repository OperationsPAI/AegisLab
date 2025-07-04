package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func DeleteDatasetByName(names []string) (int64, []string, error) {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingNames []string
	if err := tx.Model(&database.FaultInjectionSchedule{}).
		Select("injection_name").
		Where("injection_name IN (?) AND status != ?", names, consts.DatasetDeleted).
		Pluck("injection_name", &existingNames).
		Error; err != nil {
		tx.Rollback()
		logrus.Errorf("failed to query existing records: %v", err)
		return 0, nil, fmt.Errorf("database operation query failed")
	}

	nonExisting := getMissingNames(names, existingNames)
	if len(nonExisting) > 0 {
		logrus.Warnf("Non-existing names: %v", nonExisting)
	}

	result := tx.Model(&database.FaultInjectionSchedule{}).
		Where("injection_name IN (?)", existingNames).
		Update("status", consts.DatasetDeleted)

	if result.Error != nil {
		tx.Rollback()
		logrus.Errorf("update failed: %v", result.Error)
		return 0, nil, fmt.Errorf("database operation update failed")
	}

	var failedUpdates []string
	if err := tx.Model(&database.FaultInjectionSchedule{}).
		Select("injection_name").
		Where("injection_name IN (?) AND status != ?", existingNames, consts.DatasetDeleted).
		Pluck("injection_name", &failedUpdates).
		Error; err != nil {
		tx.Rollback()
		logrus.Errorf("verification failed: %v", err)
		return 0, nil, fmt.Errorf("database operation query failed")
	}

	allFailed := utils.Union(nonExisting, failedUpdates)

	if err := tx.Commit().Error; err != nil {
		logrus.Errorf("commit failed: %v", err)
		return 0, nil, fmt.Errorf("database operation failed")
	}

	return result.RowsAffected, allFailed, nil
}

// 计算差集
func getMissingNames(requested []string, existing []string) []string {
	existingSet := make(map[string]struct{})
	for _, name := range existing {
		existingSet[name] = struct{}{}
	}

	var missing []string
	for _, name := range requested {
		if _, ok := existingSet[name]; !ok {
			missing = append(missing, name)
		}
	}

	return missing
}

func GetDatasetBuildPayloads(payloads []dto.DatasetBuildPayload) ([]dto.DatasetBuildPayload, error) {
	if len(payloads) == 0 {
		return nil, fmt.Errorf("empty payloads")
	}

	names := make([]string, 0, len(payloads))
	payloadNameMap := make(map[string]*dto.DatasetBuildPayload, len(payloads))
	for i := range payloads {
		names = append(names, payloads[i].Name)
		payloadNameMap[payloads[i].Name] = &payloads[i]
	}

	var records []struct {
		InjectionName string `gorm:"column:injection_name"`
		PreDuration   int    `gorm:"column:pre_duration"`
		Benchmark     string `gorm:"column:benchmark"`
	}

	if err := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Select("injection_name, pre_duration, benchmark").
		Where("injection_name IN ?", names).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to query injection schedules: %v", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no records found for the given names")
	}

	result := make([]dto.DatasetBuildPayload, 0, len(records))
	for _, record := range records {
		if payload, ok := payloadNameMap[record.InjectionName]; ok {
			payload.PreDuration = record.PreDuration
			payload.Benchmark = record.Benchmark
			result = append(result, *payload)
		}
	}

	return result, nil
}

func GetDatasetWithGroupIDs(groupIDs []string) ([]dto.DatasetJoinedResult, error) {
	var results []struct {
		GroupID string `gorm:"column:group_id"`
		Name    string `gorm:"column:injection_name"`
	}

	if err := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Joins("JOIN tasks ON tasks.id	 = fault_injection_schedules.task_id").
		Where("tasks.group_id IN ? AND fault_injection_schedules.status = ?", groupIDs, consts.DatasetBuildSuccess).
		Select("tasks.group_id, fault_injection_schedules.injection_name").
		Scan(&results).
		Error; err != nil {
		return nil, err
	}

	joinedResults := make([]dto.DatasetJoinedResult, 0, len(results))
	for _, r := range results {
		var joinedResult dto.DatasetJoinedResult
		joinedResult.Convert(r.GroupID, r.Name)
		joinedResults = append(joinedResults, joinedResult)
	}

	return joinedResults, nil
}

func GetDatasetByName(name string, status ...int) (*dto.DatasetItemWithID, error) {
	query := database.DB.Where("injection_name = ?", name)

	if len(status) == 0 {
		query = query.Where("status != ?", consts.DatasetDeleted)
	} else if len(status) == 1 {
		query = query.Where("status = ?", status[0])
	} else {
		query = query.Where("status IN ?", status)
	}

	var record database.FaultInjectionSchedule
	if err := query.First(&record).Error; err != nil {
		return nil, err
	}

	var item dto.DatasetItemWithID
	if err := item.Convert(record); err != nil {
		return nil, err
	}

	return &item, nil
}

func GetInjection(column, param string) (*database.FaultInjectionSchedule, error) {
	query := database.DB.Where(fmt.Sprintf("%s = ?", column), param)

	var record database.FaultInjectionSchedule
	if err := query.First(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

func ListDisplayConfigsByTraceIDs(traceIDs []string) (map[string]any, error) {
	query := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Select("tasks.trace_id, fault_injection_schedules.display_config").
		Joins("JOIN tasks ON tasks.id = fault_injection_schedules.task_id")

	if len(traceIDs) > 0 {
		query = query.Where("tasks.trace_id IN (?)", traceIDs)
	}

	var records []struct {
		TraceID       string `gorm:"column:trace_id"`
		DisplayConfig string `gorm:"column:display_config"`
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to query display configs: %v", err)
	}

	result := make(map[string]any, len(records))
	if len(traceIDs) > 0 {
		for _, traceID := range traceIDs {
			result[traceID] = nil
		}
	}

	for _, record := range records {
		var config map[string]any
		if err := json.Unmarshal([]byte(record.DisplayConfig), &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal display config for trace_id %s: %v", record.TraceID, err)
		}

		result[record.TraceID] = config
	}

	return result, nil
}

func ListExistingDisplayConfigs(configs []string) ([]string, error) {
	query := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Select("display_config").
		Where("display_config in (?) AND status = ?", configs, consts.DatasetBuildSuccess)

	var existingEngineConfigs []string
	if err := query.Pluck("engine_config", &existingEngineConfigs).Error; err != nil {
		return nil, err
	}

	return existingEngineConfigs, nil
}

func ListEngineConfigsByNames(names []string) ([]string, error) {
	query := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Select("engine_config").
		Where("injection_name IN (?)", names)

	var configs []string
	if err := query.Pluck("engine_config", &configs).Error; err != nil {
		return nil, fmt.Errorf("failed to query engine configs: %v", err)
	}

	return configs, nil
}

func ListDatasetByExecutionIDs(executionIDs []int) ([]dto.DatasetItemWithID, error) {
	query := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Joins("JOIN execution_results ON execution_results.dataset = fault_injection_schedules.id")

	if len(executionIDs) > 0 {
		query = query.Where("execution_results.id IN (?)", executionIDs)
	}

	var injections []database.FaultInjectionSchedule
	if err := query.Find(&injections).Error; err != nil {
		return nil, err
	}

	items := make([]dto.DatasetItemWithID, 0, len(injections))
	for _, injection := range injections {
		var item dto.DatasetItemWithID
		if err := item.Convert(injection); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func ListDatasetWithPagination(pageNum, pageSize int) (int64, []database.FaultInjectionSchedule, error) {
	genericQueryParams := &genericQueryParams{
		builder: func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", consts.DatasetBuildSuccess)
		},
		sortField:     "created_at desc",
		pageNum:       pageNum,
		pageSize:      pageSize,
		selectColumns: []string{"id", "fault_type", "task_id", "injection_name", "display_config", "pre_duration", "start_time", "end_time"},
	}
	return genericQueryWithBuilder[database.FaultInjectionSchedule](genericQueryParams)
}

func ListInjections(params *dto.ListInjectionsReq) (int64, []database.FaultInjectionSchedule, error) {
	opts, err := params.TimeRangeQuery.Convert()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to convert time range query: %v", err)
	}

	builder := func(db *gorm.DB) *gorm.DB {
		query := db

		if params.Env != "" {
			query = query.Where("labels ->> 'env' = ?", params.Env)
		}

		if params.Batch != "" {
			query = query.Where("labels ->> 'batch' = ?", params.Batch)
		}

		if params.Benchmark != "" {
			query = query.Where("benchmark = ?", params.Benchmark)
		}

		if params.Status != nil {
			query = query.Where("status = ?", *params.Status)
		}

		if params.FaultType != nil {
			query = query.Where("fault_type = ?", *params.FaultType)
		}

		if opts != nil {
			startTime, endTime := opts.GetTimeRange()
			if !startTime.IsZero() && !endTime.IsZero() {
				query = query.Where("created_at >= ? AND created_at <= ?", startTime, endTime)
			}
		}

		return query
	}

	genericQueryParams := &genericQueryParams{
		builder:   builder,
		sortField: fmt.Sprintf("created_at %s", params.Sort),
		limit:     params.Limit,
	}
	return genericQueryWithBuilder[database.FaultInjectionSchedule](genericQueryParams)
}

func UpdateStatusByDataset(name string, status int) error {
	return updateRecord(name, map[string]any{
		"status": status,
	})
}

func UpdateTimeByDataset(name string, startTime, endTime time.Time) error {
	return updateRecord(name, map[string]any{
		"start_time": startTime,
		"end_time":   endTime,
		"status":     consts.DatasetInjectSuccess,
	})
}

func updateRecord(name string, updates map[string]any) error {
	if len(updates) == 0 {
		return fmt.Errorf("empty update fields")
	}

	var record database.FaultInjectionSchedule
	err := database.DB.
		Where("injection_name = ?", name).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("record with name %q not found", name)
		}
		return fmt.Errorf("failed to query record: %v", err)
	}

	result := database.DB.
		Model(&record).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update record: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("record found but no fields were updated, possibly because values are unchanged")
	}

	return nil
}

func GetAllFaultInjectionNoIssues(params *dto.FaultInjectionNoIssuesReq) (int64, []database.FaultInjectionNoIssues, error) {
	opts, err := params.TimeRangeQuery.Convert()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to convert time range query: %v", err)
	}

	builder := func(db *gorm.DB) *gorm.DB {
		query := db

		if params.Env != "" {
			query = query.Where("labels ->> 'env' = ?", params.Env)
		}

		if params.Batch != "" {
			query = query.Where("labels ->> 'batch' = ?", params.Batch)
		}

		if opts != nil {
			startTime, endTime := opts.GetTimeRange()
			if !startTime.IsZero() && !endTime.IsZero() {
				query = query.Where("created_at >= ? AND created_at <= ?", startTime, endTime)
			}
		}

		return query
	}

	genericQueryParams := &genericQueryParams{
		builder:       builder,
		sortField:     "dataset_id desc",
		selectColumns: []string{},
	}
	return genericQueryWithBuilder[database.FaultInjectionNoIssues](genericQueryParams)
}

func GetAllFaultInjectionWithIssues(params *dto.FaultInjectionWithIssuesReq) ([]database.FaultInjectionWithIssues, error) {
	opts, err := params.TimeRangeQuery.Convert()
	if err != nil {
		return nil, fmt.Errorf("failed to convert time range query: %v", err)
	}

	subQuery := database.DB.
		Model(&database.FaultInjectionWithIssues{}).
		Select("dataset_id, MAX(created_at) as max_created_at")

	if params.Env != "" {
		subQuery = subQuery.Where("labels ->> 'env' = ?", params.Env)
	}

	if params.Batch != "" {
		subQuery = subQuery.Where("labels ->> 'batch' = ?", params.Batch)
	}

	if opts != nil {
		startTime, endTime := opts.GetTimeRange()
		if !startTime.IsZero() && !endTime.IsZero() {
			subQuery = subQuery.Where("created_at >= ? AND created_at <= ?", startTime, endTime)
		}
	}

	subQuery = subQuery.Group("dataset_id")

	var results []database.FaultInjectionWithIssues
	if err := database.DB.
		Model(&database.FaultInjectionWithIssues{}).
		Joins("JOIN (?) AS latest ON fault_injection_with_issues.dataset_id = latest.dataset_id AND fault_injection_with_issues.created_at = latest.max_created_at", subQuery).Find(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

func GetFLByDatasetName(datasetName string) (*database.FaultInjectionSchedule, error) {
	var record database.FaultInjectionSchedule
	if err := database.DB.Where("injection_name = ?", datasetName).First(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

func GetFaultInjectionStatistics(opts dto.TimeFilterOptions) (map[string]int64, error) {
	startTime, endTime := opts.GetTimeRange()
	var noIssuesCount, withIssuesCount int64

	if err := database.DB.Model(&database.FaultInjectionNoIssues{}).
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Count(&noIssuesCount).Error; err != nil {
		return nil, err
	}

	if err := database.DB.Model(&database.FaultInjectionWithIssues{}).
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Count(&withIssuesCount).Error; err != nil {
		return nil, err
	}

	return map[string]int64{
		"no_issues":   noIssuesCount,
		"with_issues": withIssuesCount,
		"total":       noIssuesCount + withIssuesCount,
	}, nil
}
