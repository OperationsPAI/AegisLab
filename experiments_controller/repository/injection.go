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

func FindExistingDisplayConfigs(configs []string) ([]string, error) {
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

func GetDisplayConfigByTraceIDs(traceIDs []string) (map[string]any, error) {
	result := make(map[string]any)
	for _, traceID := range traceIDs {
		result[traceID] = nil
	}

	var records []struct {
		TraceID       string `gorm:"column:trace_id"`
		DisplayConfig string `gorm:"column:display_config"`
	}

	if err := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Select("tasks.trace_id, fault_injection_schedules.display_config").
		Joins("JOIN tasks ON tasks.id = fault_injection_schedules.task_id").
		Where("tasks.trace_id IN (?)", traceIDs).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to query display configs: %v", err)
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

func ListEngineConfigByNames(names []string) ([]string, error) {
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

func GetInjection(column, param string) (*dto.InjectionItem, error) {
	query := database.DB.Where(fmt.Sprintf("%s = ?", column), param)

	var record database.FaultInjectionSchedule
	if err := query.First(&record).Error; err != nil {
		return nil, err
	}

	var item dto.InjectionItem
	if err := item.Convert(record); err != nil {
		return nil, err
	}

	return &item, nil
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
	return paginateQuery[database.FaultInjectionSchedule](
		"status = ?",
		[]any{consts.DatasetBuildSuccess},
		"created_at desc",
		pageNum,
		pageSize,
		[]string{"id", "fault_type", "task_id", "injection_name", "display_config", "pre_duration", "start_time", "end_time"},
	)
}

func ListInjectionWithPagination(pageNum, pageSize int) (int64, []database.FaultInjectionSchedule, error) {
	return paginateQuery[database.FaultInjectionSchedule](
		"status != ?",
		[]any{consts.DatasetDeleted},
		"created_at desc",
		pageNum,
		pageSize,
		[]string{"id", "task_id", "fault_type", "display_config", "status", "start_time", "end_time"},
	)
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

// GetFaultInjectionNoIssues 查询没有问题的故障注入记录
func GetFaultInjectionNoIssues(pageNum, pageSize int) (int64, []database.FaultInjectionNoIssues, error) {
	return paginateQuery[database.FaultInjectionNoIssues](
		"1 = 1", // 视图本身已经包含过滤条件，这里使用始终为真的条件
		[]any{},
		"DatasetID desc", // 按数据集ID降序排序
		pageNum,
		pageSize,
		[]string{}, // 查询所有字段
	)
}

// GetFaultInjectionWithIssues 查询有问题的故障注入记录
func GetFaultInjectionWithIssues(pageNum, pageSize int) (int64, []database.FaultInjectionWithIssues, error) {
	return paginateQuery[database.FaultInjectionWithIssues](
		"1 = 1", // 视图本身已经包含过滤条件，这里使用始终为真的条件
		[]any{},
		"DatasetID desc", // 按数据集ID降序排序
		pageNum,
		pageSize,
		[]string{}, // 查询所有字段
	)
}

// GetAllFaultInjectionNoIssues 查询所有没有问题的故障注入记录（不分页）
func GetAllFaultInjectionNoIssues() (int64, []database.FaultInjectionNoIssues, error) {
	return queryAll[database.FaultInjectionNoIssues](
		"1 = 1", // 视图本身已经包含过滤条件，这里使用始终为真的条件
		[]any{},
		"DatasetID desc", // 按数据集ID降序排序
		[]string{},       // 查询所有字段
	)
}

// GetAllFaultInjectionWithIssues 查询所有有问题的故障注入记录（不分页）
func GetAllFaultInjectionWithIssues() (int64, []database.FaultInjectionWithIssues, error) {
	return queryAll[database.FaultInjectionWithIssues](
		"1 = 1", // 视图本身已经包含过滤条件，这里使用始终为真的条件
		[]any{},
		"DatasetID desc", // 按数据集ID降序排序
		[]string{},       // 查询所有字段
	)
}

// GetFaultInjectionNoIssuesByDatasetID 根据数据集ID查询没有问题的故障注入记录
func GetFaultInjectionNoIssuesByDatasetID(datasetID int) (*database.FaultInjectionNoIssues, error) {
	var record database.FaultInjectionNoIssues
	if err := database.DB.Where("DatasetID = ?", datasetID).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// GetFaultInjectionWithIssuesByDatasetID 根据数据集ID查询有问题的故障注入记录
func GetFaultInjectionWithIssuesByDatasetID(datasetID int) (*database.FaultInjectionWithIssues, error) {
	var record database.FaultInjectionWithIssues
	if err := database.DB.Where("DatasetID = ?", datasetID).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// GetFaultInjectionStatistics 获取故障注入统计信息
func GetFaultInjectionStatistics() (map[string]int64, error) {
	var noIssuesCount, withIssuesCount int64

	// 统计没有问题的记录数
	if err := database.DB.Model(&database.FaultInjectionNoIssues{}).Count(&noIssuesCount).Error; err != nil {
		return nil, err
	}

	// 统计有问题的记录数
	if err := database.DB.Model(&database.FaultInjectionWithIssues{}).Count(&withIssuesCount).Error; err != nil {
		return nil, err
	}

	return map[string]int64{
		"no_issues":   noIssuesCount,
		"with_issues": withIssuesCount,
		"total":       noIssuesCount + withIssuesCount,
	}, nil
}
