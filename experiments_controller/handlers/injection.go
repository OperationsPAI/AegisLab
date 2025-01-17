package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	chaos "github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type InjectCancelReq struct {
	TaskID string `json:"task_id"`
}

type InjectCancelResp struct {
}

type InjectListResp struct {
	TaskID     string                         `json:"task_id"`
	Name       string                         `json:"name"`
	Status     string                         `json:"status"`
	InjectTime time.Time                      `json:"inject_time"`
	Duration   int                            `json:"duration"` // minutes
	FaultType  string                         `json:"fault_type"`
	Para       executor.FaultInjectionPayload `json:"para"`
}

type InjectParaResp struct {
	Specification map[string][]chaos.ActionSpace `json:"specification"`
	KeyMap        map[chaos.ChaosType]string     `json:"keymap"`
}

type InjectResp struct {
	TaskID string `json:"task_id"`
}

type InjectStatusResp struct {
	Task InjectTask `json:"task"`
	Logs []string   `json:"logs"`
}

type InjectTask struct {
	ID        string                         `json:"id"`
	Type      string                         `json:"type"`
	Payload   executor.FaultInjectionPayload `json:"payload"`
	Status    string                         `json:"status"`
	CreatedAt time.Time                      `json:"created_at"`
}

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

// GetInjectionList
//
//	@Summary		获取注入列表和必要的简略信息
//	@Description	获取注入列表和必要的简略信息
//	@Tags			injection
//	@Produce		json
//	@Consumes		application/json
//	@Success		200	{object}		GenericResponse[[]InjectListResp]
//	@Failure		400	{object}	GenericResponse[[]InjectListResp]
//	@Failure		500	{object}	GenericResponse[[]InjectListResp]
//	@Router			/api/v1/injection/getlist [post]
func GetInjectionList(c *gin.Context) {
	var faultRecords []database.FaultInjectionSchedule
	if err := database.DB.Find(&faultRecords).Error; err != nil {
		JSONResponse[interface{}](c, http.StatusInternalServerError, "Failed to retrieve tasks", nil)
		return
	}

	var resps []InjectListResp
	for _, record := range faultRecords {
		var payload executor.FaultInjectionPayload
		if err := json.Unmarshal([]byte(record.Config), &payload); err != nil {
			logrus.Error(fmt.Sprintf("Payload of inject record %d unmarshaling failed: %s", record.ID, err))
			continue
		}

		resps = append(resps, InjectListResp{
			TaskID:     record.TaskID,
			Name:       record.InjectionName,
			Status:     DatasetStatusMap[record.Status],
			InjectTime: record.StartTime,
			Duration:   record.Duration,
			FaultType:  chaos.ChaosTypeMap[chaos.ChaosType(record.FaultType)],
			Para:       payload,
		})
	}

	JSONResponse(c, http.StatusOK, "", resps)
}

// GetInjectionStatus
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
//	@Router			/api/v1/injection/getstatus [get]
func GetInjectionStatus(c *gin.Context) {
	taskID := c.Param("taskID")

	var task database.Task
	if err := database.DB.First(&task, "id = ?", taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			JSONResponse[interface{}](c, http.StatusNotFound, "Task not found", nil)
		} else {
			JSONResponse[interface{}](c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve task %s", taskID), nil)
		}
		return
	}

	var payload executor.FaultInjectionPayload
	if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
		logrus.Error(fmt.Sprintf("Payload of inject record %s unmarshaling failed: %s", task.ID, err))
		JSONResponse[interface{}](c, http.StatusInternalServerError, fmt.Sprintf("Failed to unmarshal payload %s", taskID), nil)
		return
	}

	injectTask := InjectTask{
		ID:        task.ID,
		Type:      task.Type,
		Payload:   payload,
		Status:    task.Status,
		CreatedAt: task.CreatedAt,
	}

	logKey := fmt.Sprintf("task:%s:logs", taskID)
	ctx := c.Request.Context()
	logs, err := client.GetRedisClient().LRange(ctx, logKey, 0, -1).Result()
	if errors.Is(err, redis.Nil) {
		logs = []string{}
	} else if err != nil {
		JSONResponse[interface{}](c, http.StatusInternalServerError, "Failed to retrieve logs", nil)
		return
	}

	JSONResponse(c, http.StatusOK, "", InjectStatusResp{Task: injectTask, Logs: logs})
}

// GetInjectionPara 获取注入参数
//
//	@Summary		获取故障注入参数
//	@Description	获取可用的故障注入参数和类型映射
//	@Tags			injection
//	@Produce		application/json
//	@Success		200	{object}	GenericResponse[InjectParaResp]	"返回故障注入参数和类型映射"
//	@Failure		500	{object}	GenericResponse[ant]
//	@Router			/api/v1/injection/getpara [get]
func GetInjectionPara(c *gin.Context) {
	choice := make(map[string][]chaos.ActionSpace, 0)
	for tp, spec := range chaos.SpecMap {
		actionSpace, err := chaos.GenerateActionSpace(spec)
		if err != nil {
			JSONResponse[interface{}](c, http.StatusInternalServerError, "Failed to generate action space", nil)
			return
		}

		name := chaos.GetChaosTypeName(tp)
		choice[name] = actionSpace
	}

	resp := InjectParaResp{Specification: choice, KeyMap: chaos.ChaosTypeMap}
	JSONResponse[interface{}](c, http.StatusOK, "", resp)
}

// InjectFault
// TODO 批量注入故障
//
//	@Summary		注入故障
//	@Description	注入故障
//	@Tags			injection
//	@Produce		application/json
//	@Consumes		application/json
//	@Param			body	body		executor.FaultInjectionPayload	true	"请求体"
//	@Success		200		{object}	GenericResponse[InjectResp]
//	@Failure		400		{object}	GenericResponse[any]
//	@Failure		500		{object}	GenericResponse[any]
//	@Router			/api/v1/injection/submit [post]
func SubmitFaultInjection(c *gin.Context) {
	var payload executor.FaultInjectionPayload
	if err := c.BindJSON(&payload); err != nil {
		JSONResponse[interface{}](c, http.StatusBadRequest, "Invalid JSON payload", nil)
		return
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		JSONResponse[interface{}](c, http.StatusInternalServerError, "Failed to marshal payload", nil)
		return
	}

	ctx := c.Request.Context()
	content, ok := executor.Task.SubmitTask(ctx, "FaultInjection", jsonPayload)
	if !ok {
		JSONResponse[interface{}](c, http.StatusInternalServerError, content, nil)
		return
	}

	var resp InjectResp
	if err := json.Unmarshal([]byte(content), &resp); err != nil {
		JSONResponse[interface{}](c, http.StatusInternalServerError, "Failed to unmarshal content to response", nil)
		return
	}
	JSONResponse(c, http.StatusAccepted, "Fault injection submitted successfully", resp)
}
