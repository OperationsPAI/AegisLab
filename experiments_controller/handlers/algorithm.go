package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/CUHK-SE-Group/rcabench/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetAlgorithmResp
type GetAlgorithmResp struct {
	Name string `json:"name"`
}

type AlgorithmExecutionResp struct {
	TaskID string `json:"task_id"`
}

// GetAlgorithmList
//
//	@Summary		获取算法列表
//	@Description	获取算法列表
//	@Tags			algorithm
//	@Produce		application/json
//	@Success		200		{object}	GenericResponse[[]GetAlgorithmResp]
//	@Failure		400		{object}	GenericResponse[any]
//	@Failure		500		{object}	GenericResponse[any]
//	@Router			/api/v1/algo/getlist [get]
func GetAlgorithmList(c *gin.Context) {
	pwd, err := os.Getwd()
	if err != nil {
		JSONResponse[interface{}](c, http.StatusInternalServerError, "Get work directory failed", nil)
		return
	}

	parentDir := filepath.Dir(pwd)
	algoPath := filepath.Join(parentDir, "algorithms")

	algoFiles, err := utils.GetSubFiles(algoPath)
	if err != nil {
		JSONResponse[interface{}](c, http.StatusInternalServerError, fmt.Sprintf("Failed to list files in %s: %v", algoPath, err), nil)
		return
	}

	var algoResps []GetAlgorithmResp
	for _, algoFile := range algoFiles {
		tomlPath := filepath.Join(algoPath, algoFile, "info.toml")

		var algoResp GetAlgorithmResp
		if _, err := toml.DecodeFile(tomlPath, &algoResp); err != nil {
			logrus.Error(fmt.Sprintf("Failed to get info.toml in %s: %v", algoPath, err))
			continue
		}
		algoResps = append(algoResps, algoResp)
	}

	JSONResponse(c, http.StatusOK, "OK", algoResps)
}

// GetAlgorithmList
//
//	@Summary		执行算法
//	@Description	执行算法
//	@Tags			algorithm
//	@Produce		application/json
//	@Consumes		application/json
//	@Param			type	query		string								true	"任务类型"
//	@Param			body	body		executor.AlgorithmExecutionPayload	true	"请求体"
//	@Success		200		{object}	GenericResponse[AlgorithmExecutionResp]
//	@Failure		400		{object}	GenericResponse[any]
//	@Failure		500		{object}	GenericResponse[any]
//	@Router			/api/v1/algorithm [post]
func SubmitAlgorithmExecution(c *gin.Context) {
	var payload executor.AlgorithmExecutionPayload
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
	content, ok := executor.Task.SubmitTask(ctx, "RunAlgorithm", jsonPayload)
	if !ok {
		JSONResponse[interface{}](c, http.StatusInternalServerError, content, nil)
		return
	}

	var resp AlgorithmExecutionResp
	if err := json.Unmarshal([]byte(content), &resp); err != nil {
		JSONResponse[interface{}](c, http.StatusInternalServerError, "Failed to unmarshal content to response", nil)
		return
	}
	JSONResponse(c, http.StatusAccepted, "Algorithm Execution submitted successfully", resp)
}
