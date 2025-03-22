package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	cli "github.com/CUHK-SE-Group/chaos-experiment/client"
	chaos "github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/CUHK-SE-Group/rcabench/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	reLog     *regexp.Regexp
	reLogOnce sync.Once
)

// CancelInjection
//
//	@Summary		取消注入
//	@Description	取消注入
//	@Tags			injection
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		InjectCancelReq	true	"请求体"
//	@Success		200		{object}	GenericResponse[InjectCancelResp]
//	@Failure		400		{object}	GenericResponse[InjectCancelResp]
//	@Failure		500		{object}	GenericResponse[InjectCancelResp]
//	@Router			/api/v1/injection/cancel [post]
func CancelInjection(c *gin.Context) {
}

// GetInjectionDetail
//
//	@Summary		获取单个注入的详细信息
//	@Description	获取单个注入的详细信息
//	@Tags			injection
//	@Produce		json
//	@Consumes		application/json
//	@Param 			taskID 	path		string		true		"任务 ID"
//	@Success		200		{objec}		GenericResponse[InjectStatusResp]
//	@Failure		404		{object}	GenericResponse[any]
//	@Failure		500		{object}	GenericResponse[any]
//	@Router			/api/v1/injections [get]
func GetInjectionDetail(c *gin.Context) {
	// 获取查询参数并校验是否合法
	var req dto.InjectionListReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, dto.FormatErrorMessage(err, map[string]string{}))
		return
	}

	var taskReq dto.TaskReq
	if err := c.BindUri(&taskReq); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid URI")
		return
	}

	logEntry := logrus.WithField("task_id", taskReq.TaskID)

	var task database.Task
	if err := database.DB.Where("tasks.id = ?", taskReq.TaskID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			message := "Task not found"
			logEntry.WithError(err).Error(message)
			dto.ErrorResponse(c, http.StatusNotFound, message)
		} else {
			message := "Failed to retrieve task of injection"
			logEntry.WithError(err).Error(message)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
		}

		return
	}

	var payload executor.FaultInjectionPayload
	if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
		message := "Failed to unmarshal payload of injection record"
		logEntry.WithError(err).Error("Failed to unmarshal payload of injection record")
		dto.ErrorResponse(c, http.StatusInternalServerError, message)
		return
	}

	injectTask := dto.InjectionTask{
		ID:        task.ID,
		Type:      task.Type,
		Payload:   payload,
		Status:    task.Status,
		CreatedAt: task.CreatedAt,
	}

	logKey := fmt.Sprintf("task:%s:logs", task.ID)
	ctx := c.Request.Context()
	logs, err := client.GetRedisClient().LRange(ctx, logKey, 0, -1).Result()
	if errors.Is(err, redis.Nil) {
		logs = []string{}
	} else if err != nil {
		message := "Failed to retrieve logs"
		logrus.WithError(err).Error(message)
		dto.ErrorResponse(c, http.StatusInternalServerError, message)
		return
	}

	dto.SuccessResponse(c, dto.InjectionDetailResp{Task: injectTask, Logs: logs})
}

// GetInjectionList
//
//	@Summary		分页查询注入记录列表
//	@Description	获取注入记录列表（支持分页参数）
//	@Tags			injection
//	@Produce		json
//	@Consumes		application/json
//	@Success		200	{object}		GenericResponse[[]InjectListResp]
//	@Failure		400	{object}		GenericResponse[[]InjectListResp]
//	@Failure		500	{object}		GenericResponse[[]InjectListResp]
//	@Router			/api/v1/injections/getlist [post]
func GetInjectionList(c *gin.Context) {
	var req dto.InjectionListReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, dto.FormatErrorMessage(err, map[string]string{}))
		return
	}

	pageNum := *req.PageNum
	pageSize := *req.PageSize

	db := database.DB.Model(&database.FaultInjectionSchedule{}).Where("status != ?", executor.DatesetDeleted)
	db.Scopes(
		database.Sort("proposed_end_time desc"),
		database.Paginate(pageNum, pageSize),
	).Select("SQL_CALC_FOUND_ROWS *")

	// 查询总记录数
	var total int64
	if err := db.Raw("SELECT FOUND_ROWS()").Scan(&total).Error; err != nil {
		message := "Failed to count injection schedules"
		logrus.WithError(err).Error(message)
		dto.ErrorResponse(c, http.StatusInternalServerError, message)
		return
	}

	// 查询分页数据
	var records []database.FaultInjectionSchedule
	if err := db.Find(&records).Error; err != nil {
		message := "Failed to retrieve injections"
		logrus.WithError(err).Error(message)
		dto.ErrorResponse(c, http.StatusInternalServerError, message)
		return
	}

	var injections []dto.InjectionItem
	for _, record := range records {
		var payload executor.FaultInjectionPayload
		if err := json.Unmarshal([]byte(record.Config), &payload); err != nil {
			logrus.WithField("id", record.ID).WithError(err).Error("Failed t unmarshal payload of injection")
			continue
		}

		injections = append(injections, dto.InjectionItem{
			ID:              record.ID,
			TaskID:          record.TaskID,
			FaultType:       chaos.ChaosTypeMap[chaos.ChaosType(record.FaultType)],
			Name:            record.InjectionName,
			Status:          dto.DatasetStatusMap[record.Status],
			InjectTime:      record.StartTime,
			ProposedEndTime: record.ProposedEndTime,
			Duration:        record.Duration,
			Payload:         payload,
		})
	}

	dto.SuccessResponse(c, &dto.PaginationResp[dto.InjectionItem]{
		Total: total,
		Data:  injections,
	})
}

// GetInjectionParameters
//
//	@Summary		获取故障注入参数
//	@Description	获取可用的故障注入参数和类型映射
//	@Tags			injection
//	@Produce		json
//	@Success		200	{object}	GenericResponse[InjectParaResp]
//	@Failure		500	{object}	GenericResponse[any]
//	@Router			/api/v1/injections/getpara [get]
func GetInjectionParameters(c *gin.Context) {
	choice := make(map[string][]chaos.ActionSpace, 0)
	for tp, spec := range chaos.SpecMap {
		actionSpace, err := chaos.GenerateActionSpace(spec)
		if err != nil {
			message := "Failed to generate action space"
			logrus.WithError(err).Error(message)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		name := chaos.GetChaosTypeName(tp)
		choice[name] = actionSpace
	}

	dto.SuccessResponse(c, dto.InjectionParaResp{Specification: choice, KeyMap: chaos.ChaosTypeMap})
}

// SubmitFaultInjection
//
//	@Summary		注入故障
//	@Description	注入故障
//	@Tags			injection
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		[]executor.FaultInjectionPayload	true	"请求体"
//	@Success		200		{object}	GenericResponse[InjectResp]
//	@Failure		400		{object}	GenericResponse[any]
//	@Failure		500		{object}	GenericResponse[any]
//	@Router			/api/v1/injections [post]
func SubmitFaultInjection(c *gin.Context) {
	groupID := c.GetString("groupID")
	logrus.Infof("SubmitFaultInjection called, groupID: %s", groupID)

	var req dto.InjectionSubmitReq
	if err := c.BindJSON(&req); err != nil {
		logrus.Error(err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// TODO 多种类任务组
	if req.IsCroned {
		handleCronTask(&req)
	} else {
		if !validateOneTimeTask(c, &req, time.Now()) {
			return
		}
	}

	if req.CheckConflicts() {
		dto.ErrorResponse(c, http.StatusBadRequest,
			"Conflicts between the execution_time of the payloads exist",
		)
		return
	}

	var traces []dto.Trace
	for _, payload := range req.Payloads {
		taskID, traceID, err := executor.SubmitTask(context.Background(), &executor.UnifiedTask{
			Type:        executor.TaskTypeFaultInjection,
			Payload:     utils.StructToMap(payload.FaultInjectionPayload),
			Immediate:   false,
			ExecuteTime: payload.ExecutionTime.Unix(),
			GroupID:     groupID,
		})
		if err != nil {
			message := "Failed to submit task"
			logrus.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}
		traces = append(traces, dto.Trace{TraceID: traceID, HeadTaskID: taskID})
	}

	dto.JSONResponse(c, http.StatusAccepted, "Fault injections submitted successfully", dto.SubmitResp{GroupID: groupID, Traces: traces})
}

// GetNamespacePod 获取命名空间中的 Pod 标签
//
//	@Summary 获取命名空间中的 Pod 标签
//	@Description 返回指定命名空间中符合条件的 Pod 标签列表
//	@Tags	injection
//	@Produce json
//	@Success 200 {object} InjectNamespacePodResp "返回命名空间和对应的 Pod 标签信息"
//	@Failure 500 {object} any "服务器内部错误，无法获取 Pod 标签"
//	@Router			/api/v1/injections/namespace_pods [get]
func GetNamespacePods(c *gin.Context) {
	namespaceInfo := make(map[string][]string)
	for _, ns := range config.GetStringSlice("injection.namespace") {
		labels, err := cli.GetLabels(ns, config.GetString("injection.label"))
		if err != nil {
			message := "Failed to get labels"
			logrus.WithField("namespace", ns).WithError(err).Error(message)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
		}
		namespaceInfo[ns] = labels
	}

	dto.SuccessResponse(c, dto.InjectionNamespaceInfoResp{NamespaceInfo: namespaceInfo})
}

// 处理周期性任务的时间生成
func handleCronTask(req *dto.InjectionSubmitReq) {
	currentTime := time.Now()
	interval := time.Duration(req.Interval) * time.Minute

	for i := range req.Payloads {
		offset := interval * time.Duration(i)
		newTime := currentTime.Add(offset)
		req.Payloads[i].ExecutionTime = &newTime
	}
}

// 检查一次性任务的必填字段和时间有效性
func validateOneTimeTask(c *gin.Context, req *dto.InjectionSubmitReq, currentTime time.Time) bool {
	// 检查空值
	var missingIndices []int
	for i, payload := range req.Payloads {
		if payload.ExecutionTime == nil {
			missingIndices = append(missingIndices, i)
		}
	}

	if len(missingIndices) > 0 {
		errorMsg := fmt.Sprintf("以下Payload的execution_time不能为空: %v",
			strings.Trim(strings.Join(strings.Fields(fmt.Sprint(missingIndices)), ", "), "[]"))
		dto.ErrorResponse(c, http.StatusBadRequest, errorMsg)
		return false
	}

	// 按时间排序
	sort.Slice(req.Payloads, func(i, j int) bool {
		return req.Payloads[i].ExecutionTime.Before(*req.Payloads[j].ExecutionTime)
	})

	// 检查时间有效性
	if len(req.Payloads) > 0 {
		for i, payload := range req.Payloads {
			if payload.ExecutionTime.Before(currentTime) {
				dto.ErrorResponse(c, http.StatusBadRequest,
					fmt.Sprintf("The earliest execution_time has expired: payloads[0] (%s)",
						req.Payloads[i].ExecutionTime.Format(time.DateTime),
					))

				return false
			}
		}
	}

	return true
}
