package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/consts"
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
//	@Success		200		{object}	dto.GenericResponse[dto.AlgorithmListResp]
//	@Failure		400		{object}	dto.GenericResponse[any]
//	@Failure		500		{object}	dto.GenericResponse[any]
//	@Router			/api/v1/algorithms [get]
func GetAlgorithmList(c *gin.Context) {
	algoPath := config.GetString("algo.storage_path")
	algoDirs, err := utils.GetAllSubDirectories(algoPath)
	if err != nil {
		message := "failed to list files"
		logrus.WithField("algo_path", algoPath).Errorf("%s: %v", message, err)
		dto.ErrorResponse(c, http.StatusInternalServerError, message)
		return
	}

	tomlName := "info.toml"
	var algorithms []string
	for _, algoDir := range algoDirs {
		tomlPath := filepath.Join(algoDir, tomlName)

		data, err := os.ReadFile(tomlPath)
		if err != nil {
			message := "failed to read toml file"
			logrus.WithField("toml_path", tomlPath).Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		var config map[string]any
		if err := toml.Unmarshal(data, &config); err != nil {
			message := "failed to parse toml file"
			logrus.WithField("toml_path", tomlPath).Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		field := "name"
		name, ok := utils.GetMapField(config, field)
		if !ok {
			message := fmt.Sprintf("missing field in %s", tomlPath)
			logrus.WithField("field", field).Errorf("%s: %v", message, err)
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
//	@Param			body	body		[]dto.AlgorithmExecutionPayload	true	"请求体"
//	@Success		200		{object}	dto.GenericResponse[dto.SubmitResp]
//	@Failure		400		{object}	dto.GenericResponse[any]
//	@Failure		500		{object}	dto.GenericResponse[any]
//	@Router			/api/v1/algorithms [post]
func SubmitAlgorithmExecution(c *gin.Context) {
	groupID := c.GetString("groupID")
	logrus.Infof("SubmitAlgorithmExecution called, groupID: %s", groupID)

	var payloads []dto.AlgorithmExecutionPayload
	if err := c.BindJSON(&payloads); err != nil {
		logrus.Error(err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	parts := strings.Split(config.GetString("harbor.repository"), "/")
	harborConfig := utils.HarborConfig{
		Host:     config.GetString("harbor.host"),
		Project:  parts[len(parts)-1],
		Username: config.GetString("harbor.username"),
		Password: config.GetString("harbor.password"),
	}

	for i := range payloads {
		if payloads[i].Tag == "" {
			harborConfig.Repo = payloads[i].Image
			tag, err := utils.GetLatestTag(harborConfig)
			if err != nil {
				logrus.Errorf("failed to get latest tag: %v", err)
				dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get latest tag")
				return
			}

			payloads[i].Tag = tag
		}

		for key := range payloads[i].EnvVars {
			if _, exists := dto.ExecuteEnvVarNameMap[key]; !exists {
				message := fmt.Sprintf("the key %s is invalid in env_vars", key)
				logrus.Errorf(message)
				dto.ErrorResponse(c, http.StatusInternalServerError, message)
				return
			}
		}
	}

	traces := make([]dto.Trace, 0, len(payloads))
	for idx, payload := range payloads {
		taskID, traceID, err := executor.SubmitTask(c.Request.Context(), &executor.UnifiedTask{
			Type:      consts.TaskTypeRunAlgorithm,
			Payload:   utils.StructToMap(payload),
			Immediate: true,
			GroupID:   groupID,
		})
		if err != nil {
			message := "failed to submit task"
			logrus.Error(message)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		traces = append(traces, dto.Trace{TraceID: traceID, HeadTaskID: taskID, Index: idx})
	}

	dto.JSONResponse(c, http.StatusAccepted, "Algorithm executions submitted successfully", dto.SubmitResp{GroupID: groupID, Traces: traces})
}

// BuildAlgorithm handles algorithm file upload, extraction and build submission
//
//	@Summary		构建算法镜像
//	@Description	通过上传文件或指定算法名称来构建算法镜像
//	@Tags			algorithm
//	@Accept			multipart/form-data
//	@Produce		application/json
//	@Param			file		formData	file	false	"算法文件 (zip/tar.gz)"
//	@Param			algo	formData	string	false	"算法名称"
//	@Success		202			{object}	dto.GenericResponse[dto.SubmitResp]
//	@Failure		400			{object}	dto.GenericResponse[any]
//	@Failure		500			{object}	dto.GenericResponse[any]
//	@Router			/api/v1/algorithms/build [post]
func SubmitAlgorithmBuilding(c *gin.Context) {
	var extractDir string
	algoName := c.PostForm("algo")

	if algoName == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "Either file upload or algorithm name is required")
		return
	}
	payload := make(map[string]any)
	payload[consts.BuildAlgorithm] = algoName

	file, header, err := c.Request.FormFile("file")
	if err == nil {
		defer file.Close()

		const maxSize = 5 * 1024 * 1024
		if header.Size > maxSize {
			dto.ErrorResponse(c, http.StatusBadRequest, "File size exceeds 5MB limit")
			return
		}

		// Check file type
		fileName := header.Filename
		fileExt := strings.ToLower(filepath.Ext(fileName))

		isZip := fileExt == ".zip"
		isTarGz := (fileExt == ".gz" && strings.HasSuffix(strings.ToLower(fileName), ".tar.gz")) ||
			(fileExt == ".tgz")

		if !isZip && !isTarGz {
			dto.ErrorResponse(c, http.StatusBadRequest, "Only zip and tar.gz files are allowed")
			return
		}

		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "algorithm-upload-*")
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create temporary directory")
			return
		}

		// Save the file to the temporary directory
		filePath := filepath.Join(tempDir, header.Filename)
		out, err := os.Create(filePath)
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to save file")
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to save file")
			return
		}

		// Create extraction directory
		extractDir = path.Join(config.GetString("algo.storage_path"), algoName)
		err = os.MkdirAll(extractDir, 0755)
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create extraction directory")
			return
		}

		// Extract the file based on its format
		var extractErr error
		if isZip {
			extractErr = utils.ExtractZip(filePath, extractDir)
		} else if isTarGz {
			extractErr = utils.ExtractTarGz(filePath, extractDir)
		}

		if extractErr != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to extract file: "+extractErr.Error())
			return
		}

		payload[consts.BuildAlgorithmPath] = extractDir
	}

	taskID, traceID, err := executor.SubmitTask(context.Background(), &executor.UnifiedTask{
		Type:      consts.TaskTypeBuildImages,
		Payload:   payload,
		Immediate: true,
	})

	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to submit build task")
		return
	}

	dto.JSONResponse(c, http.StatusAccepted, "Algorithm build task submitted successfully",
		dto.SubmitResp{Traces: []dto.Trace{{TraceID: traceID, HeadTaskID: taskID, Index: 0}}})
}
