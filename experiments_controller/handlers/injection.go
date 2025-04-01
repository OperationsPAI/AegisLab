package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/CUHK-SE-Group/chaos-experiment/handler"
	chaos "github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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

type NodeItem struct {
	Description string `json:"description"`
	Range       []int  `json:"range"`
}

func GetInjectionConf(c *gin.Context) {
	root, err := chaos.StructToNode[handler.InjectionConf]()
	if err != nil {
		logrus.Errorf("struct InjectionConf to node failed: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to read injection conf")
		return
	}

	type result struct {
		key   string
		value map[string]NodeItem
	}

	chaosMap := make(map[string]map[string]NodeItem, len(root.Children))
	resultChan := make(chan result, len(root.Children))
	var wg sync.WaitGroup

	// 并行处理每个节点
	for _, node := range root.Children {
		wg.Add(1)
		go func(n *chaos.Node) {
			defer wg.Done()
			m := make(map[string]NodeItem, len(n.Children))
			for _, child := range n.Children {
				m[child.Name] = NodeItem{
					Description: child.Description,
					Range:       child.Range,
				}
			}
			resultChan <- result{key: n.Name, value: m}
		}(node)
	}

	// 等待所有处理完成并关闭channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集处理结果
	for res := range resultChan {
		chaosMap[res.key] = res.value
	}

	dto.SuccessResponse(c, root)
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
	var taskReq dto.TaskReq
	if err := c.BindUri(&taskReq); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid URI")
		return
	}

	logEntry := logrus.WithField("task_id", taskReq.TaskID)

	var task database.Task
	if err := database.DB.Where("tasks.id = ?", taskReq.TaskID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			message := "task not found"
			logEntry.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusNotFound, message)
		} else {
			message := "failed to retrieve task of injection"
			logEntry.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
		}

		return
	}

	var payload dto.InjectionPayload
	if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
		message := "failed to unmarshal payload of injection record"
		logEntry.Error(message)
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
		logrus.Errorf("%s: %v", message, err)
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
		dto.ErrorResponse(c, http.StatusBadRequest, formatErrorMessage(err, map[string]string{}))
		return
	}

	db := database.DB.Model(&database.FaultInjectionSchedule{}).Where("status != ?", consts.DatasetDeleted)
	db.Scopes(
		database.Sort("proposed_end_time desc"),
		database.Paginate(req.PageNum, req.PageSize),
	).Select("SQL_CALC_FOUND_ROWS *")

	// 查询总记录数
	var total int64
	if err := db.Raw("SELECT FOUND_ROWS()").Scan(&total).Error; err != nil {
		message := "failed to count injection schedules"
		logrus.Errorf("%s: %v", message, err)
		dto.ErrorResponse(c, http.StatusInternalServerError, message)
		return
	}

	// 查询分页数据
	var records []database.FaultInjectionSchedule
	if err := db.Find(&records).Error; err != nil {
		message := "failed to retrieve injections"
		logrus.Errorf("%s: %v", message, err)
		dto.ErrorResponse(c, http.StatusInternalServerError, message)
		return
	}

	var injections []dto.InjectionItem
	for _, record := range records {
		var payload map[string]any
		if err := json.Unmarshal([]byte(record.Config), &payload); err != nil {
			logrus.WithField("id", record.ID).Errorf("failed to parse injection config: %v", err)
		}

		injections = append(injections, dto.InjectionItem{
			ID:         record.ID,
			TaskID:     record.TaskID,
			FaultType:  chaos.ChaosTypeMap[chaos.ChaosType(record.FaultType)],
			Name:       record.InjectionName,
			Status:     dto.DatasetStatusMap[record.Status],
			InjectTime: record.StartTime,
			Duration:   record.Duration,
			Payload:    payload,
		})
	}

	dto.SuccessResponse(c, &dto.PaginationResp[dto.InjectionItem]{
		Total: total,
		Data:  injections,
	})
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

	executionTimes, err := req.GetExecutionTimes()
	if err != nil {
		logrus.Error(err)
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var traces []dto.Trace
	for i, spec := range req.Specs {
		payload := map[string]any{
			consts.InjectBenchmark:   req.Benchmark,
			consts.InjectPreDuration: req.PreDuration,
			consts.InjectSpec:        spec,
		}

		taskID, traceID, err := executor.SubmitTask(context.Background(), &executor.UnifiedTask{
			Type:        consts.TaskTypeFaultInjection,
			Payload:     payload,
			Immediate:   false,
			ExecuteTime: executionTimes[i].Unix(),
			GroupID:     groupID,
		})

		if err != nil {
			message := "failed to submit task"
			logrus.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}
		traces = append(traces, dto.Trace{TraceID: traceID, HeadTaskID: taskID})
	}

	dto.JSONResponse(c, http.StatusAccepted, "Fault injections submitted successfully", dto.SubmitResp{GroupID: groupID, Traces: traces})
}
