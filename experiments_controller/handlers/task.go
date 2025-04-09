package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/repository"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func GetTaskDetail(c *gin.Context) {
	var req dto.TaskReq
	if err := c.BindUri(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid URI")
		return
	}

	logEntry := logrus.WithField("task_id", req.TaskID)

	taskItem, err := repository.FindTaskItemByID(req.TaskID)
	if err != nil {
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

	logKey := fmt.Sprintf("task:%s:logs", taskItem.ID)
	logs, err := client.GetRedisClient().LRange(c.Request.Context(), logKey, 0, -1).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			message := "Failed to retrieve logs"
			logrus.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		logs = []string{}
	}

	dto.SuccessResponse(c, dto.TaskDetailResp{Task: *taskItem, Logs: logs})
}

// GetStream
//
//	@Summary      获取 trace 任务状态事件流
//	@Description  通过Server-Sent Events (SSE) 实时获取任务的执行状态更新，直到任务完成或连接关闭
//	@Tags         injection
//	@Produce      text/event-stream
//	@Consumes	  application/json
//	@Param        task_id  path      string  				true  "需要监控的任务ID"
//	@Success      200      {object}  nil     				"成功建立SSE连接，持续推送事件流"
//	@Failure      400      {object}  dto.GenericResponse[any]	"无效的任务ID格式"
//	@Failure      404      {object}  dto.GenericResponse[any]  	"指定ID的任务不存在"
//	@Failure      500      {object}  dto.GenericResponse[any]  	"服务器内部错误"
func GetTaskStream(c *gin.Context) {
	var req dto.TaskReq
	if err := c.BindUri(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid param")
		return
	}

	logEntry := logrus.WithField("task_id", req.TaskID)

	item, err := repository.FindTaskItemByID(req.TaskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			message := "Task not found"
			logEntry.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusNotFound, message)
			return
		}

		message := "failed to retrieve task of injection"
		logEntry.Errorf("%s: %v", message, err)
		dto.ErrorResponse(c, http.StatusInternalServerError, message)

		return
	}

	pubsub := client.GetRedisClient().Subscribe(c, fmt.Sprintf(consts.SubChannel, item.TraceID))
	defer pubsub.Close()

	expectedTaskType := consts.TaskType(item.Type)

	switch consts.TaskType(item.Type) {
	case consts.TaskTypeRunAlgorithm:
		expectedTaskType = consts.TaskTypeCollectResult
	case consts.TaskTypeFaultInjection:
		benchmark, ok := item.Payload[consts.InjectBenchmark]
		if !ok {

		}
		if benchmark != "" {
			expectedTaskType = consts.TaskTypeBuildDataset
		}
	}

	sendStreamMessge(c, item.TraceID, expectedTaskType)
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
