package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/CUHK-SE-Group/rcabench/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
