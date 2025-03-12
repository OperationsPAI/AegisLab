package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/CUHK-SE-Group/rcabench/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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
		message := "Failed to list files"
		logrus.WithField("algo_path", algoPath).WithError(err).Error(message)
		dto.ErrorResponse(c, http.StatusInternalServerError, message)
		return
	}

	tomlName := "info.toml"
	var algorithms []string
	for _, algoDir := range algoDirs {
		tomlPath := filepath.Join(algoDir, tomlName)

		data, err := os.ReadFile(tomlPath)
		if err != nil {
			message := "Failed to read toml file"
			logrus.WithField("toml_path", tomlPath).WithError(err).Error(message)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		var config map[string]any
		if err := toml.Unmarshal(data, &config); err != nil {
			message := "Failed to parse toml file"
			logrus.WithField("toml_path", tomlPath).WithError(err).Error(message)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		field := "name"
		name, ok := utils.GetTomlString(config, field)
		if !ok {
			message := fmt.Sprintf("Missing field in %s", tomlPath)
			logrus.WithField("field", field).WithError(err).Error(message)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		algorithms = append(algorithms, name)
	}

	dto.SuccessResponse(c, dto.AlgorithmListResp{Algorithms: algorithms})
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
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	logrus.Infof("Received executing algorithm payloads: %+v", payloads)

	var taskIDs []string
	for _, payload := range payloads {
		taskID, err := executor.SubmitTask(c.Request.Context(), &executor.UnifiedTask{
			Type:      executor.TaskTypeRunAlgorithm,
			Payload:   utils.StructToMap(payload),
			Immediate: true,
			GroupID:   groupID,
		})
		if err != nil {
			message := "Failed to submit task"
			logrus.WithField("task_id", taskID).WithError(err).Error(message)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		taskIDs = append(taskIDs, taskID)
	}

	dto.JSONResponse(c, http.StatusAccepted, "Algorithm Execution submitted successfully", dto.SubmitResp{GroupID: groupID, TaskIDs: taskIDs})
}
