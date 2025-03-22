package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// StreamTask
//
//	@Summary      获取任务状态事件流
//	@Description  通过Server-Sent Events (SSE) 实时获取任务的执行状态更新，直到任务完成或连接关闭
//	@Tags         injection
//	@Produce      text/event-stream
//	@Consumes	  application/json
//	@Param        task_id  path      string  				true  "需要监控的任务ID"
//	@Success      200      {object}  nil     				"成功建立SSE连接，持续推送事件流"
//	@Failure      400      {object}  GenericResponse[any]	"无效的任务ID格式"
//	@Failure      404      {object}  GenericResponse[any]  	"指定ID的任务不存在"
//	@Failure      500      {object}  GenericResponse[any]  	"服务器内部错误"
func StreamTask(c *gin.Context) {
	var req dto.TaskReq
	if err := c.BindUri(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid URI")
		return
	}

	logEntry := logrus.WithField("task_id", req.TaskID)

	var task database.Task
	if err := database.DB.Where("tasks.id = ?", req.TaskID).First(&task).Error; err != nil {
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

	pubsub := client.GetRedisClient().Subscribe(c, fmt.Sprintf(executor.SubChannel, task.TraceID))
	defer pubsub.Close()

	// 主动退出函数，关闭连接
	expectedTaskType := executor.TaskType(task.Type)

	switch executor.TaskType(task.Type) {
	case executor.TaskTypeRunAlgorithm:
		expectedTaskType = executor.TaskTypeCollectResult
	case executor.TaskTypeFaultInjection:
		var payload dto.InjectionPayload
		if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
			message := "Failed to unmarshal payload of injection record"
			logEntry.WithError(err).Error(message)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		if payload.Benchmark != "" {
			expectedTaskType = executor.TaskTypeBuildDataset
		}
	}

	for {
		select {
		case message := <-pubsub.Channel():
			c.SSEvent(executor.EventUpdate, message.Payload)
			c.Writer.Flush()

			var rdbMsg executor.RdbMsg
			if err := json.Unmarshal([]byte(message.Payload), &rdbMsg); err != nil {
				msg := "Failed to unmarshal payload of redis message"
				logEntry.WithError(err).Error(msg)

				c.SSEvent(executor.EventError, map[string]string{
					"error":   msg,
					"details": err.Error(),
				})
				c.Writer.Flush()

				return
			}

			switch rdbMsg.Status {
			case executor.TaskStatusCompleted:
				if rdbMsg.Type == expectedTaskType {
					c.SSEvent(executor.EventEnd, nil)
					c.Writer.Flush()

					return
				}
			case executor.TaskStatusError:
				c.SSEvent(executor.EventError, map[string]string{
					"error":   fmt.Sprintf("Failed to execute task %s", task.ID),
					"details": *rdbMsg.Error,
				})
				c.SSEvent(executor.EventEnd, nil)
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
