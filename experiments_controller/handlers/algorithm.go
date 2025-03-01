package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/CUHK-SE-Group/rcabench/utils"
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp/v3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// GetAlgorithmResp
type AlgorithmResp struct {
	Algorithms []string `json:"algorithms"`
}

type AlgorithmToml struct {
	Name string `json:"name"`
}

// GetAlgorithmList
//
//	@Summary		获取算法列表
//	@Description	获取算法列表
//	@Tags			algorithm
//	@Produce		application/json
//	@Success		200		{object}	GenericResponse[AlgorithmResp]
//	@Failure		400		{object}	GenericResponse[any]
//	@Failure		500		{object}	GenericResponse[any]
//	@Router			/api/v1/algo/getlist [get]
func GetAlgorithmList(c *gin.Context) {
	parentDir := config.GetString("workspace")
	algoPath := filepath.Join(parentDir, "algorithms")

	algoDirs, err := utils.GetAllSubDirectories(algoPath)
	if err != nil {
		JSONResponse[any](c, http.StatusInternalServerError, fmt.Sprintf("Failed to list files in %s: %v", algoPath, err), nil)
		return
	}

	tomlName := "info.toml"
	var algorithms []string
	for _, algoDir := range algoDirs {
		tomlPath := filepath.Join(algoDir, tomlName)

		var algoToml AlgorithmToml
		if _, err := toml.DecodeFile(tomlPath, &algoToml); err != nil {
			logrus.Error(fmt.Sprintf("Failed to get %s in %s: %v", tomlName, algoDir, err))
			continue
		}
		algorithms = append(algorithms, algoToml.Name)
	}

	SuccessResponse(c, AlgorithmResp{Algorithms: algorithms})
}

// StreamAlgorithm
//
//	@Summary      获取任务状态事件流
//	@Description  通过Server-Sent Events (SSE) 实时推送算法任务的执行状态更新，直到任务完成或连接关闭
//	@Tags         algorithm
//	@Produce      text/event-stream
//	@Consumes	  application/json
//	@Param        task_id  path      string  				true  "需要监控的任务ID"
//	@Success      200      {object}  nil     				"成功建立SSE连接，持续推送事件流"
//	@Failure      400      {object}  GenericResponse[any]	"无效的任务ID格式"
//	@Failure      404      {object}  GenericResponse[any]  	"指定ID的任务不存在"
//	@Failure      500      {object}  GenericResponse[any]  	"服务器内部错误"
//	@Router       /api/v1/algorithm/{task_id}/stream [get]
func StreamAlgorithmExecution(c *gin.Context) {
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
			message := "Failed to retrieve task of algorithm execution"
			logEntry.WithError(err).Error(message)
			JSONResponse[any](c, http.StatusInternalServerError, message, nil)
		}

		return
	}

	pubsub := client.GetRedisClient().Subscribe(c, fmt.Sprintf(executor.SubChannel, task.TraceID))
	defer pubsub.Close()

	for {
		select {
		case message := <-pubsub.Channel():
			c.SSEvent("update", message.Payload)
			pp.Println(message.Payload)
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
			if rdbMsg.Status == executor.TaskStatusCompleted && rdbMsg.Type == executor.TaskTypeCollectResult {
				return
			}

			if rdbMsg.Status == executor.TaskStatusError {
				return
			}

		case <-c.Writer.CloseNotify():
			return

		case <-c.Done():
			return
		}
	}
}

// SubmitAlgorithmExecution
//
//	@Summary		执行算法
//	@Description	执行算法
//	@Tags			algorithm
//	@Produce		application/json
//	@Consumes		application/json
//	@Param			body	body		[]executor.AlgorithmExecutionPayload	true	"请求体"
//	@Success		200		{object}	GenericResponse[SubmitResp]
//	@Failure		400		{object}	GenericResponse[any]
//	@Failure		500		{object}	GenericResponse[any]
//	@Router			/api/v1/algorithm [post]
func SubmitAlgorithmExecution(c *gin.Context) {
	groupID := c.GetString("groupID")
	logrus.Infof("SubmitAlgorithmExecution called, groupID: %s", groupID)

	var payloads []executor.AlgorithmExecutionPayload
	if err := c.BindJSON(&payloads); err != nil {
		JSONResponse[any](c, http.StatusBadRequest, "Invalid JSON payload", nil)
		return
	}
	logrus.Infof("Received executing algorithm payloads: %+v", payloads)

	var ids []string
	for _, payload := range payloads {
		id, err := executor.SubmitTask(c.Request.Context(), &executor.UnifiedTask{
			Type:      executor.TaskTypeRunAlgorithm,
			Payload:   StructToMap(payload),
			Immediate: true,
			GroupID:   groupID,
		})
		if err != nil {
			JSONResponse[any](c, http.StatusInternalServerError, id, nil)
			return
		}

		ids = append(ids, id)
	}

	JSONResponse(c, http.StatusAccepted, "Algorithm Execution submitted successfully", SubmitResp{GroupID: groupID, TaskIDs: ids})
}
