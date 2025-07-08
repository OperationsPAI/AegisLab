package handlers

import (
	"archive/zip"
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/executor"
	"github.com/LGU-SE-Internal/rcabench/middleware"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/LGU-SE-Internal/rcabench/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// DeleteDataset
//
//	@Summary		删除数据集数据
//	@Description	删除数据集数据
//	@Tags			dataset
//	@Produce		application/json
//	@Param			names	query		[]string	true	"数据集名称列表"
//	@Success		200		{object}	dto.GenericResponse[dto.DatasetDeleteResp]
//	@Failure		400		{object}	dto.GenericResponse[any]
//	@Failure		500		{object}	dto.GenericResponse[any]
//	@Router			/api/v1/datasets [delete]
func DeleteDataset(c *gin.Context) {
	var req dto.DatasetDeleteReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid Parameters")
		return
	}

	successCount, failedNames, err := repository.DeleteDatasetByName(req.Names)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	dto.SuccessResponse(c, dto.DatasetDeleteResp{SuccessCount: successCount, FailedNames: failedNames})
}

// DownloadDataset 处理数据集下载请求
//
//	@Summary		下载数据集打包文件
//	@Description	将指定的多个数据集打包为 ZIP 文件下载，自动排除 result.csv 和检测器结论文件。支持按组ID或数据集名称进行下载，两种方式二选一。下载文件结构：按组ID下载时为 datasets/{groupId}/{datasetName}/...，按名称下载时为 datasets/{datasetName}/...
//	@Tags			dataset
//	@Produce		application/zip
//	@Param			group_ids	query		[]string	false	"任务组ID列表，格式：group1,group2,group3。与names参数二选一，优先使用group_ids")
//	@Param			names		query		[]string	false	"数据集名称列表，格式：dataset1,dataset2,dataset3。与group_ids参数二选一"
//	@Success		200			{string}	binary		"ZIP 文件流，Content-Disposition 头中包含文件名 datasets.zip"
//	@Failure		400			{object}	dto.GenericResponse[any]	"请求参数错误：1) 参数绑定失败 2) 两个参数都为空 3) 同时提供两种参数"
//	@Failure		403			{object}	dto.GenericResponse[any]	"权限错误：请求访问的数据集路径不在系统允许的范围内"
//	@Failure		500			{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/datasets/download [get]
func DownloadDataset(c *gin.Context) {
	var req dto.DatasetDownloadReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid Parameters")
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", consts.DownloadFilename))

	zipWriter := zip.NewWriter(c.Writer)
	defer zipWriter.Close()

	excludeRules := []utils.ExculdeRule{
		{Pattern: consts.DetectorConclusionFile, IsGlob: false},
		{Pattern: consts.ExecutionResultFile, IsGlob: false},
	}

	// 定义处理函数
	handleError := func(statusCode int, err error) {
		delete(c.Writer.Header(), "Content-Disposition")
		c.Header("Content-Type", "application/json; charset=utf-8")
		dto.ErrorResponse(c, statusCode, err.Error())
	}

	// 根据输入选择下载方式
	var (
		downloadFunc func(*zip.Writer, []string, []utils.ExculdeRule) (int, error)
		input        []string
	)

	switch {
	case len(req.GroupIDs) > 0:
		downloadFunc = downloadByGroupIds
		input = req.GroupIDs
	case len(req.Names) > 0:
		downloadFunc = downloadByNames
		input = req.Names
	}

	if statusCode, err := downloadFunc(zipWriter, input, excludeRules); err != nil {
		handleError(statusCode, err)
		return
	}
}

func downloadByGroupIds(zipWriter *zip.Writer, groupIDs []string, excludeRules []utils.ExculdeRule) (int, error) {
	joinedResults, err := repository.GetDatasetWithGroupIDs(groupIDs)
	if err != nil {
		message := "failed to query datasets"
		logrus.Errorf("%s: %v", message, err)
		return http.StatusInternalServerError, fmt.Errorf("failed to query datasets")
	}

	groupedResults := make(map[string][]string)
	for _, joinedResult := range joinedResults {
		if _, exists := groupedResults[joinedResult.GroupID]; !exists {
			groupedResults[joinedResult.GroupID] = []string{}
		}

		groupedResults[joinedResult.GroupID] = append(groupedResults[joinedResult.GroupID], joinedResult.Name)
	}

	// 预先检查所有数据集路径合法性
	for _, datasets := range groupedResults {
		for _, dataset := range datasets {
			workDir := filepath.Join(config.GetString("nfs.path"), dataset)
			if !utils.IsAllowedPath(workDir) {
				logrus.WithField("path", workDir).Errorf("invalid path access")
				return http.StatusInternalServerError, fmt.Errorf("Invalid path access")
			}
		}
	}

	for groupID, datasets := range groupedResults {
		folderName := filepath.Join(consts.DownloadFilename, groupID)
		for _, dataset := range datasets {
			workDir := filepath.Join(config.GetString("nfs.path"), dataset)

			err := filepath.WalkDir(workDir, func(path string, dir fs.DirEntry, err error) error {
				if err != nil || dir.IsDir() {
					return err
				}

				relPath, _ := filepath.Rel(workDir, path)
				fullRelPath := filepath.Join(folderName, filepath.Base(workDir), relPath)
				fileName := filepath.Base(path)

				// 应用排除规则
				for _, rule := range excludeRules {
					if utils.MatchFile(fileName, rule) {
						return nil
					}
				}

				// 获取文件信息以读取修改时间
				fileInfo, err := dir.Info()
				if err != nil {
					return err
				}

				// 转换路径分隔符为/
				zipPath := filepath.ToSlash(fullRelPath)
				return utils.AddToZip(zipWriter, fileInfo, path, zipPath)
			})
			if err != nil {
				logrus.Errorf("failed to packcage: %v", err)
				return http.StatusForbidden, fmt.Errorf("Failed to pacage")
			}
		}
	}

	return http.StatusOK, nil
}

func downloadByNames(zipWriter *zip.Writer, names []string, excludeRules []utils.ExculdeRule) (int, error) {
	for _, name := range names {
		workDir := filepath.Join(config.GetString("nfs.path"), name)
		if !utils.IsAllowedPath(workDir) {
			logrus.WithField("path", workDir).Errorf("invalid path access")
			return http.StatusInternalServerError, fmt.Errorf("Invalid path access")
		}
	}

	folderName := consts.DownloadFilename
	for _, name := range names {
		workDir := filepath.Join(config.GetString("nfs.path"), name)

		err := filepath.WalkDir(workDir, func(path string, dir fs.DirEntry, err error) error {
			if err != nil || dir.IsDir() {
				return err
			}

			relPath, _ := filepath.Rel(workDir, path)
			fullRelPath := filepath.Join(folderName, filepath.Base(workDir), relPath)
			fileName := filepath.Base(path)

			// 应用排除规则
			for _, rule := range excludeRules {
				if utils.MatchFile(fileName, rule) {
					return nil
				}
			}

			// 获取文件信息以读取修改时间
			fileInfo, err := dir.Info()
			if err != nil {
				return err
			}

			// 转换路径分隔符为/
			zipPath := filepath.ToSlash(fullRelPath)
			return utils.AddToZip(zipWriter, fileInfo, path, zipPath)
		})
		if err != nil {
			logrus.Errorf("failed to packcage: %v", err)
			return http.StatusForbidden, fmt.Errorf("Failed to pacage")
		}
	}

	return http.StatusOK, nil
}

// SubmitDatasetBuilding
//
//	@Summary		批量构建数据集
//	@Description	根据指定的时间范围和基准测试容器批量构建数据集。
//	@Tags			dataset
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.SubmitDatasetBuildingReq	true	"数据集构建请求列表，每个请求包含数据集名称、时间范围、基准测试和环境变量配置"
//	@Success		202		{object}	dto.GenericResponse[dto.SubmitResp]	"成功提交数据集构建任务，返回任务组ID和跟踪信息列表"
//	@Failure		400		{object}	dto.GenericResponse[any]	"请求参数错误：1) JSON格式不正确 2) 数据集名称为空 3) 时间范围无效 4) 基准测试不存在 5) 环境变量名称不支持"
//	@Failure		500		{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/datasets [post]
func SubmitDatasetBuilding(c *gin.Context) {
	groupID := c.GetString("groupID")
	logrus.Infof("SubmitDatasetBuilding, groupID: %s", groupID)

	var payloads []dto.DatasetBuildPayload
	if err := c.BindJSON(&payloads); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	payloads, err := repository.GetDatasetBuildPayloads(payloads)
	if err != nil {
		message := "failed to get dataset build payloads"
		logrus.Errorf("%s: %v", message, err)
		dto.ErrorResponse(c, http.StatusInternalServerError, message)
		return
	}

	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		logrus.Error("failed to get span context from gin.Context")
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get span context")
		return
	}

	spanCtx := ctx.(context.Context)
	for i := range payloads {
		for key := range payloads[i].EnvVars {
			if _, exists := dto.BuildEnvVarNameMap[key]; !exists {
				message := fmt.Sprintf("the key %s is invalid in env_vars", key)
				logrus.Errorf(message)
				dto.ErrorResponse(c, http.StatusInternalServerError, message)
				return
			}
		}
	}

	traces := make([]dto.Trace, 0, len(payloads))
	for idx, payload := range payloads {
		task := &dto.UnifiedTask{
			Type:      consts.TaskTypeBuildDataset,
			Payload:   utils.StructToMap(payload),
			Immediate: true,
			GroupID:   groupID,
		}
		task.SetGroupCtx(spanCtx)

		taskID, traceID, err := executor.SubmitTask(context.Background(), task)
		if err != nil {
			message := "failed to submit task"
			logrus.Error(message)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		traces = append(traces, dto.Trace{TraceID: traceID, HeadTaskID: taskID, Index: idx})
	}

	dto.JSONResponse(c, http.StatusAccepted, "Dataset building submitted successfully",
		dto.SubmitResp{GroupID: groupID, Traces: traces},
	)
}

func UploadDataset(c *gin.Context) {
}
