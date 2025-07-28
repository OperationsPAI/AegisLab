package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/executor/analyzer"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func GetTraceStream(c *gin.Context) {
	var traceReq dto.TraceReq
	if err := c.BindUri(&traceReq); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid URI")
		return
	}

	var req dto.TraceStreamReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid Parameters")
		return
	}

	lastID := req.LastID
	if lastID == "" {
		lastID = "0"
	}

	streamKey := fmt.Sprintf(consts.StreamLogKey, traceReq.TraceID)
	logEntry := logrus.WithFields(logrus.Fields{
		"trace_id":   traceReq.TraceID,
		"stream_key": streamKey,
	})

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	if c.IsAborted() {
		return
	}

	listReq := &dto.ListTasksReq{
		TraceID: traceReq.TraceID,
	}

	if err := listReq.Validate(); err != nil {
		logEntry.Errorf("Invalid request: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	_, tasks, err := repository.ListTasks(listReq)
	if err != nil {
		logrus.Errorf("failed to fetch tasks: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tasks")
		return
	}

	headTask := tasks[0]
	var algorithms []dto.AlgorithmItem
	if consts.TaskType(headTask.Type) == consts.TaskTypeRestartService {
		if repository.CheckCachedField(ctx, consts.InjectionAlgorithmsKey, headTask.GroupID) {
			algorithms, err = repository.GetCachedAlgorithmItemsFromRedis(ctx, consts.InjectionAlgorithmsKey, headTask.GroupID)
			if err != nil {
				logEntry.Errorf("failed to get algorithms from Redis: %v", err)
				dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get algorithms")
				return
			}
		}
	}

	processor := repository.NewStreamProcessor(algorithms)

	logEntry.Infof("Reading historical events from Stream")
	historicalMessages, err := repository.ReadStreamEvents(ctx, streamKey, lastID, 100, 0)
	if err != nil && err != redis.Nil {
		logEntry.Errorf("failed to read historical events from redis: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to read event history")
		return
	}

	if len(historicalMessages) > 0 {
		lastID, err = sendSSEMessages(c, processor, historicalMessages)
		if err != nil {
			logEntry.Errorf("failed to read stream events: %v", err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "failed to read stream events")
			return
		}
	}

	logEntry.Infof("Switching to real-time event monitoring from ID: %s", lastID)
	for {
		select {
		case <-c.Done():
			logEntry.Info("Request context done")
			return

		default:
			newMessages, err := repository.ReadStreamEvents(ctx, streamKey, lastID, 10, time.Second)
			if err != nil && err != redis.Nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					logEntry.Infof("Context done while reading stream: %v", err)
					return
				}

				logEntry.Errorf("Error reading stream: %v", err)
				dto.ErrorResponse(c, http.StatusInternalServerError, "failed to read stream events")
				return
			}

			if err == redis.Nil {
				continue
			}

			lastID, err = sendSSEMessages(c, processor, newMessages)
			if err != nil {
				logEntry.Error(err)
				return
			}

			logrus.Info("Sent SSE messages, lastID:", lastID)
		}
	}
}

func sendSSEMessages(c *gin.Context, processor *repository.StreamProcessor, streams []redis.XStream) (string, error) {
	lastID, sseMessage, err := processor.ProcessMessageForSSE(streams[0].Messages[0])
	if err != nil {
		return "", err
	}

	c.Render(-1, sse.Event{
		Id:    lastID,
		Event: consts.EventUpdate,
		Data:  sseMessage,
	})
	c.Writer.Flush()

	if processor.IsCompleted() {
		c.SSEvent(consts.EventEnd, nil)
		c.Writer.Flush()
	}

	return lastID, nil
}

func GetTaskEventMap(c *gin.Context) {
	dto.SuccessResponse(c, dto.ValidTaskEventMap)
}

func GetValidTaskTypes(c *gin.Context) {
	dto.SuccessResponse(c, dto.ValidFirstTaskTypes)
}

// GetCompletedMap 获取完成状态的链路
// @Summary     获取完成状态的链路
// @Description 根据指定的时间范围获取完成状态的链路
// @Tags        trace
// @Produce     json
// @Param		lookback			query		string	false	"相对时间查询，如 1h, 24h, 7d或者是custom"
// @Param       custom_start_time 	query   string   false "当lookback=custom时必需，自定义开始时间(RFC3339格式)"
// @Param       custom_end_time  	query   string   false "当lookback=custom时必需，自定义结束时间(RFC3339格式)"
// @Success     200 {object}     	dto.GenericResponse[dto.GetCompletedMapResp]
// @Failure     400 {object}     	dto.GenericResponse[any]
// @Failure     500 {object}     	dto.GenericResponse[any]
// @Router      /api/v1/traces/completed [get]
func GetCompletedMap(c *gin.Context) {
	var req dto.GetCompletedMapReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid Parameters")
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	result, err := analyzer.GetCompletedMap(c.Request.Context(), &req)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to analyze trace")
		return
	}

	dto.SuccessResponse(c, result)
}
