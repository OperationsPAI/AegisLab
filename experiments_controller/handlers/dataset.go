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

// QueryDataset
//
//	@Summary		查询单个数据集详情
//	@Description	根据数据集名称查询单个数据集的详细信息，包括检测器结果和执行记录
//	@Tags			dataset
//	@Produce		json
//	@Param			name	query		string	true	"数据集名称"
//	@Param			sort	query		string	false	"排序方式"
//	@Success		200		{object}	dto.GenericResponse[dto.QueryDatasetResp]
//	@Failure		400		{object}	dto.GenericResponse[any]
//	@Failure		500		{object}	dto.GenericResponse[any]
//	@Router			/api/v1/datasets/query [get]
func QueryDataset(c *gin.Context) {
	var req dto.QueryDatasetReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, formatErrorMessage(err, dto.PaginationFieldMap))
		return
	}

	sortOrder, err := validateSortOrder(req.Sort)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	item, err := repository.GetDatasetByName(req.Name, consts.DatasetBuildSuccess)
	if err != nil {
		logrus.Errorf("failed to get dataset: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get dataset")
		return
	}

	logEntry := logrus.WithField("dataset", item.Name)

	detectorRecord, err := repository.GetDetectorRecordByDatasetID(item.ID)
	if err != nil {
		logEntry.Errorf("failed to retrieve detector record: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to load detector data")
		return
	}

	executionRecords, err := repository.ListExecutionRecordByDatasetID(item.ID, sortOrder)
	if err != nil {
		logEntry.Errorf("failed to retrieve execution records: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to load execution data")
		return
	}

	if len(executionRecords) == 0 {
		logEntry.Warn("no execution records found for dataset")
	}

	dto.SuccessResponse(c, &dto.QueryDatasetResp{
		DatasetItem:      item.DatasetItem,
		DetectorResult:   detectorRecord,
		ExecutionResults: executionRecords,
	})
}

// GetDatasetList
//
//	@Summary		分页查询数据集列表
//	@Description	获取状态为成功的注入数据集列表（支持分页参数）
//	@Tags			dataset
//	@Produce		json
//	@Param			page_num	query		int		true	"页码（从1开始）" minimum(1) default(1)
//	@Param			page_size	query		int		true	"每页数量" minimum(5) maximum(20) default(10)
//	@Success		200			{object}	dto.GenericResponse[dto.PaginationResp[dto.DatasetItem]] "成功响应"
//	@Failure		400			{object}	dto.GenericResponse[any] "参数校验失败"
//	@Failure		500			{object}	dto.GenericResponse[any] "服务器内部错误"
//	@Router			/api/v1/datasets [get]
func GetDatasetList(c *gin.Context) {
	// 获取查询参数并校验是否合法
	var req dto.DatasetListReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, formatErrorMessage(err, dto.PaginationFieldMap))
		return
	}

	total, records, err := repository.ListDatasetWithPagination(req.PageNum, req.PageSize)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	items := make([]dto.DatasetItem, 0, len(records))
	for _, record := range records {
		var item dto.DatasetItem
		if err := item.Convert(record); err != nil {
			logrus.WithField("dataset", record.InjectionName).Error(err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "invalid injection configuration")
			return
		}

		items = append(items, item)
	}

	totalPages := (total + int64(req.PageSize) - 1) / int64(req.PageSize)
	dto.SuccessResponse(c, dto.PaginationResp[dto.DatasetItem]{
		Total:      total,
		TotalPages: totalPages,
		Items:      items,
	})
}

// DownloadByGroupIDs 处理数据集下载请求
//
//		@Summary		下载数据集打包文件
//		@Description	将指定路径的多个数据集打包为 ZIP 文件下载（自动排除 result.csv 文件）
//		@Tags			dataset
//		@Produce		application/zip
//		@Consumes		application/json
//	 @Param          group_ids    query       []string    false   "数据集组ID列表，与names参数二选一"
//	 @Param          names        query       []string    false   "数据集名称列表，与group_ids参数二选一"
//		@Success		200			{string} 	binary 		"ZIP 文件流"
//		@Failure		400			{object}	dto.GenericResponse[any] "参数绑定错误"
//		@Failure		403			{object}	dto.GenericResponse[any] "非法路径访问"
//		@Failure		500			{object}	dto.GenericResponse[any] "文件打包失败"
//		@Router			/api/v1/datasets/download [get]
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
//	@Description	批量构建数据集
//	@Tags			dataset
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		[]dto.DatasetBuildPayload	true	"请求体"
//	@Success		202		{object}	dto.GenericResponse[dto.SubmitResp]
//	@Failure		400		{object}	dto.GenericResponse[any]
//	@Failure		500		{object}	dto.GenericResponse[any]
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
