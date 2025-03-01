package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	cli "github.com/CUHK-SE-Group/chaos-experiment/client"
	chaos "github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type InjectCancelResp struct {
}

type InjectDetailResp struct {
	Task InjectTask `json:"task"`
	Logs []string   `json:"logs"`
}

type InjectListResult struct {
	TaskID     string                         `json:"task_id"`
	Name       string                         `json:"name"`
	Status     string                         `json:"status"`
	InjectTime time.Time                      `json:"inject_time"`
	Duration   int                            `json:"duration"` // minutes
	FaultType  string                         `json:"fault_type"`
	Para       executor.FaultInjectionPayload `json:"para"`
}

type InjectListResp struct {
	Results []InjectListResult `json:"result"`
}

type InjectNamespacePodResp struct {
	NamespaceInfo map[string][]string `json:"namespace_info"`
}

type InjectParaResp struct {
	Specification map[string][]chaos.ActionSpace `json:"specification"`
	KeyMap        map[chaos.ChaosType]string     `json:"keymap"`
}

type InjectTask struct {
	ID        string                         `json:"id"`
	Type      string                         `json:"type"`
	Payload   executor.FaultInjectionPayload `json:"payload"`
	Status    string                         `json:"status"`
	CreatedAt time.Time                      `json:"created_at"`
}

type SubmitResp struct {
	GroupID string   `json:"group_id"`
	TaskIDs []string `json:"task_ids"`
}

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
	var taskReq TaskReq
	if err := c.BindUri(&taskReq); err != nil {
		JSONResponse[any](c, http.StatusBadRequest, "Invalid URI", nil)
		return
	}

	logEntry := logrus.WithField("task_id", taskReq.TaskID)

	var task database.Task
	if err := database.DB.Where("tasks.id = ?", taskReq.TaskID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			message := "Task not found"
			logEntry.WithError(err).Error(message)
			JSONResponse[any](c, http.StatusNotFound, message, nil)
		} else {
			message := "Failed to retrieve task of injection"
			logEntry.WithError(err).Error(message)
			JSONResponse[any](c, http.StatusInternalServerError, message, nil)
		}

		return
	}

	var payload executor.FaultInjectionPayload
	if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
		logEntry.WithError(err).Error("Failed to unmarshal payload of injection record")
		JSONResponse[any](c, http.StatusInternalServerError, "Failed to unmarshal payload", nil)
		return
	}

	injectTask := InjectTask{
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
		JSONResponse[any](c, http.StatusInternalServerError, "Failed to retrieve logs", nil)
		return
	}

	SuccessResponse(c, InjectDetailResp{Task: injectTask, Logs: logs})
}

// StreamInjection
//
//	@Summary      获取任务状态事件流
//	@Description  通过Server-Sent Events (SSE) 实时注入任务的执行状态更新，直到任务完成或连接关闭
//	@Tags         injection
//	@Produce      text/event-stream
//	@Consumes	  application/json
//	@Param        task_id  path      string  				true  "需要监控的任务ID"
//	@Success      200      {object}  nil     				"成功建立SSE连接，持续推送事件流"
//	@Failure      400      {object}  GenericResponse[any]	"无效的任务ID格式"
//	@Failure      404      {object}  GenericResponse[any]  	"指定ID的任务不存在"
//	@Failure      500      {object}  GenericResponse[any]  	"服务器内部错误"
//	@Router       /api/v1/injection/{task_id}/stream [get]
func StreamInjection(c *gin.Context) {
	var taskReq TaskReq
	if err := c.BindUri(&taskReq); err != nil {
		JSONResponse[any](c, http.StatusBadRequest, "Invalid URI", nil)
		return
	}

	logEntry := logrus.WithField("task_id", taskReq.TaskID)

	var task database.Task
	if err := database.DB.Where("tasks.id = ?", taskReq.TaskID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			message := "Task not found"
			logEntry.WithError(err).Error(message)
			JSONResponse[any](c, http.StatusNotFound, message, nil)
		} else {
			message := "Failed to retrieve task of injection"
			logEntry.WithError(err).Error(message)
			JSONResponse[any](c, http.StatusInternalServerError, message, nil)
		}

		return
	}

	var payload executor.FaultInjectionPayload
	if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
		logEntry.WithError(err).Error("Failed to unmarshal payload of injection record")
		JSONResponse[any](c, http.StatusInternalServerError, "Failed to unmarshal payload", nil)
		return
	}

	pubsub := client.GetRedisClient().Subscribe(c, fmt.Sprintf(executor.SubChannel, task.TraceID))
	defer pubsub.Close()

	for {
		select {
		case message := <-pubsub.Channel():
			c.SSEvent("update", message.Payload)
			c.Writer.Flush()

			var rdbMsg executor.RdbMsg
			if err := json.Unmarshal([]byte(message.Payload), &rdbMsg); err != nil {
				msg := "Failed to unmarshal payload of redis message"
				logEntry.WithError(err).Error(msg)

				c.SSEvent("error", map[string]string{
					"error":   msg,
					"details": err.Error(),
				})
				c.Writer.Flush()

				return
			}

			// 主动退出函数，关闭连接
			expectedTaskType := executor.TaskTypeFaultInjection
			if payload.Benchmark != nil {
				expectedTaskType = executor.TaskTypeBuildDataset
			}

			switch rdbMsg.Status {
			case executor.TaskStatusCompleted:
				if rdbMsg.Type == expectedTaskType {
					return
				}
			case executor.TaskStatusError:
				return
			}

		case <-c.Writer.CloseNotify():
			return

		case <-c.Done():
			return
		}
	}
}

// GetInjectionList
//
//	@Summary		获取注入列表和必要的简略信息
//	@Description	获取注入列表和必要的简略信息
//	@Tags			injection
//	@Produce		json
//	@Consumes		application/json
//	@Success		200	{object}		GenericResponse[[]InjectListResp]
//	@Failure		400	{object}		GenericResponse[[]InjectListResp]
//	@Failure		500	{object}		GenericResponse[[]InjectListResp]
//	@Router			/api/v1/injections/getlist [post]
func GetInjectionList(c *gin.Context) {
	var faultRecords []database.FaultInjectionSchedule
	if err := database.DB.Find(&faultRecords).Error; err != nil {
		JSONResponse[any](c, http.StatusInternalServerError, "Failed to retrieve tasks", nil)
		return
	}

	var results []InjectListResult
	for _, record := range faultRecords {
		var payload executor.FaultInjectionPayload
		if err := json.Unmarshal([]byte(record.Config), &payload); err != nil {
			logrus.Error(fmt.Sprintf("Payload of inject record %d unmarshaling failed: %s", record.ID, err))
			continue
		}

		results = append(results, InjectListResult{
			TaskID:     record.TaskID,
			Name:       record.InjectionName,
			Status:     DatasetStatusMap[record.Status],
			InjectTime: record.StartTime,
			Duration:   record.Duration,
			FaultType:  chaos.ChaosTypeMap[chaos.ChaosType(record.FaultType)],
			Para:       payload,
		})
	}

	SuccessResponse(c, InjectListResp{Results: results})
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
			JSONResponse[any](c, http.StatusInternalServerError, "Failed to generate action space", nil)
			return
		}

		name := chaos.GetChaosTypeName(tp)
		choice[name] = actionSpace
	}

	SuccessResponse(c, InjectParaResp{Specification: choice, KeyMap: chaos.ChaosTypeMap})
}

// InjectFault
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

	var payloads []executor.FaultInjectionPayload // 改为接收数组
	if err := c.BindJSON(&payloads); err != nil {
		JSONResponse[any](c, http.StatusBadRequest, "Invalid JSON payload", nil)
		return
	}
	logrus.Infof("Received fault injection payloads: %+v", payloads)

	var ids []string
	t := time.Now()
	for _, payload := range payloads {
		id, err := executor.SubmitTask(context.Background(), &executor.UnifiedTask{
			Type:        executor.TaskTypeFaultInjection,
			Payload:     StructToMap(payload),
			Immediate:   false,
			ExecuteTime: t.Unix(),
			GroupID:     groupID,
		})
		if err != nil {
			JSONResponse[any](c, http.StatusInternalServerError, id, nil)
			return
		}
		t = t.Add(time.Duration(payload.Duration)*time.Minute + time.Duration(config.GetInt("injection.interval"))*time.Minute)
		ids = append(ids, id)
	}

	JSONResponse(c, http.StatusAccepted, "Fault injections submitted successfully", SubmitResp{GroupID: groupID, TaskIDs: ids})
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
			JSONResponse[any](c, http.StatusInternalServerError, fmt.Sprintf("Failed to get labels from namespace %s", ns), nil)
		}
		namespaceInfo[ns] = labels
	}

	JSONResponse(c, http.StatusOK, "OK", InjectNamespacePodResp{NamespaceInfo: namespaceInfo})
}
