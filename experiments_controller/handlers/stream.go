package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// GetStream
//
//	@Summary      获取 trace 任务状态事件流
//	@Description  通过Server-Sent Events (SSE) 实时获取任务的执行状态更新，直到任务完成或连接关闭
//	@Tags         injection
//	@Produce      text/event-stream
//	@Consumes	  application/json
//	@Param        task_id  path      string  				true  "需要监控的任务ID"
//	@Success      200      {object}  nil     				"成功建立SSE连接，持续推送事件流"
//	@Failure      400      {object}  GenericResponse[any]	"无效的任务ID格式"
//	@Failure      404      {object}  GenericResponse[any]  	"指定ID的任务不存在"
//	@Failure      500      {object}  GenericResponse[any]  	"服务器内部错误"
func GetStream(c *gin.Context) {
	var req dto.StreamReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid param")
		return
	}

	var logEntry *logrus.Entry
	var task database.Task
	var err error

	if req.TaskID != "" {
		logEntry = logrus.WithField("task_id", req.TaskID)
		err = database.DB.Where("tasks.id = ?", req.TaskID).First(&task).Error
	} else if req.TraceID != "" {
		logEntry = logrus.WithField("trace_id", req.TraceID)
		err = database.DB.Where("tasks.trace_id = ?", req.TraceID).
			Order("created_at ASC").
			First(&task).Error
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			message := "Task not found"
			logEntry.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusNotFound, message)
			return
		}

		message := "failed to retrieve task of injection"
		logEntry.Error("%s: %v", message, err)
		dto.ErrorResponse(c, http.StatusInternalServerError, message)

		return
	}

	pubsub := client.GetRedisClient().Subscribe(c, fmt.Sprintf(consts.SubChannel, task.TraceID))
	defer pubsub.Close()

	// 主动退出函数，关闭连接
	expectedTaskType := consts.TaskType(task.Type)

	switch consts.TaskType(task.Type) {
	case consts.TaskTypeRunAlgorithm:
		expectedTaskType = consts.TaskTypeCollectResult
	case consts.TaskTypeFaultInjection:
		var payload dto.InjectionPayload
		if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
			message := "Failed to unmarshal payload of injection record"
			logEntry.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		if payload.Benchmark != "" {
			expectedTaskType = consts.TaskTypeBuildDataset
		}
	}

	sendStreamMessge(c, task.TraceID, expectedTaskType)
}

func sendStreamMessge(c *gin.Context, traceID string, expectedTaskType consts.TaskType) {
	logEntry := logrus.WithField("trace_id", traceID)

	pubsub := client.GetRedisClient().Subscribe(c, fmt.Sprintf(consts.SubChannel, traceID))
	defer pubsub.Close()

	for {
		select {
		case message := <-pubsub.Channel():
			c.SSEvent(consts.EventUpdate, message.Payload)
			c.Writer.Flush()

			var rdbMsg dto.RdbMsg
			if err := json.Unmarshal([]byte(message.Payload), &rdbMsg); err != nil {
				msg := "unmarshal payload of redis message failed"
				logEntry.Errorf("%s: %v", msg, err)

				c.SSEvent(consts.EventError, map[string]string{
					"error":   msg,
					"details": err.Error(),
				})
				c.Writer.Flush()

				return
			}

			switch rdbMsg.Status {
			case consts.TaskStatusCompleted:
				if rdbMsg.Type == expectedTaskType {
					c.SSEvent(consts.EventEnd, nil)
					c.Writer.Flush()

					return
				}
			case consts.TaskStatusError:
				c.SSEvent(consts.EventError, map[string]string{
					"error":   fmt.Sprintf("execute task %s failed", rdbMsg.TaskID),
					"details": rdbMsg.Error,
				})
				c.SSEvent(consts.EventEnd, nil)
				c.Writer.Flush()

				return
			}

		case <-c.Writer.CloseNotify():
			return

		case <-c.Done():
			return
		}
	}
}
