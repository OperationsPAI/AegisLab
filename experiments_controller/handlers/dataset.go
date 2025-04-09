package handlers

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/CUHK-SE-Group/rcabench/repository"
	"github.com/CUHK-SE-Group/rcabench/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// DeleteDataset
//
//	@Summary		删除数据集数据
//	@Description	删除数据集数据
//	@Tags			dataset
//	@Produce		application/json
//	@Param			datasetID	path		int					true	"数据集 ID"
//	@Success		200			{object}	dto.GenericResponse[int]
//	@Failure		400			{object}	dto.GenericResponse[any]
//	@Failure		500			{object}	dto.GenericResponse[any]
//	@Router			/api/v1/datasets/delete [delete]
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
//	@Summary		查询单个数据集
//	@Description	查询单个数据集
//	@Tags			dataset
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		[]dto.QueryDatasetReq	true	"请求体"
//	@Success		200		{object}	dto.GenericResponse[dto.QueryDatasetResp]
//	@Failure		400		{object}	dto.GenericResponse[any]
//	@Failure		500		{object}	dto.GenericResponse[any]
//	@Router			/api/v1/datasets [post]
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
		logrus.Errorf("failed to get injection record: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to retrieve dataset")
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

	dto.SuccessResponse(c, &dto.PaginationResp[dto.DatasetItem]{
		Total: total,
		Data:  items,
	})
}

// DownloadDataset 处理数据集下载请求
//
//	@Summary		下载数据集打包文件
//	@Description	将指定路径的多个数据集打包为 ZIP 文件下载（自动排除 result.csv 文件）
//	@Tags			dataset
//	@Produce		application/zip
//	@Consumes		application/json
//	@Success		200			{string} 	binary 		"ZIP 文件流"
//	@Failure		400			{object}	dto.GenericResponse[any] "参数绑定错误"
//	@Failure		403			{object}	dto.GenericResponse[any] "非法路径访问"
//	@Failure		500			{object}	dto.GenericResponse[any] "文件打包失败"
//	@Router			/api/v1/datasets/download [get]
func DownloadDataset(c *gin.Context) {
	filename := "package"
	excludeRules := []utils.ExculdeRule{{Pattern: "result.csv", IsGlob: false}}

	var req dto.DatasetDownloadReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, formatErrorMessage(err, dto.PaginationFieldMap))
		return
	}

	joinedResults, err := repository.GetDatasetWithGroupID(req.GroupIDs)
	if err != nil {
		message := "failed to query datasets"
		logrus.Errorf("%s: %v", message, err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to query datasets")
		return
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
				message := "invalid path access"
				logrus.WithField("path", workDir).Error(message)
				dto.ErrorResponse(c, http.StatusForbidden, message)
				return
			}
		}
	}

	// 设置响应头
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", filename))

	zipWriter := zip.NewWriter(c.Writer)
	defer zipWriter.Close()

	for groupID, datasets := range groupedResults {
		folderName := filepath.Join(filename, groupID)

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

				// 转换路径分隔符为/
				zipPath := filepath.ToSlash(fullRelPath)
				return utils.AddToZip(zipWriter, path, zipPath)
			})

			if err != nil {
				delete(c.Writer.Header(), "Content-Disposition")
				c.Header("Content-Type", "application/json; charset=utf-8")
				message := "failed to packcage"
				logrus.Errorf("%s: %v", message, err)
				dto.ErrorResponse(c, http.StatusInternalServerError, message)
				return
			}
		}
	}
}

// TODO 优化
// BuildDataset
//
//	@Summary		批量构建数据集
//	@Description	批量构建数据集
//	@Tags			dataset
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		[]dto.DatasetPayload	true	"请求体"
//	@Success		200		{object}	dto.GenericResponse[dto.SubmitResp]
//	@Failure		400		{object}	dto.GenericResponse[any]
//	@Failure		500		{object}	dto.GenericResponse[any]
//	@Router			/api/v1/datasets [post]
func SubmitDatasetBuilding(c *gin.Context) {
	groupID := c.GetString("groupID")
	logrus.Infof("SubmitDatasetBuilding, groupID: %s", groupID)

	var payloads []dto.DatasetPayload
	if err := c.BindJSON(&payloads); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	logrus.Infof("Received building dataset payloads: %+v", payloads)

	var traces []dto.Trace
	for _, payload := range payloads {
		taskID, traceID, err := executor.SubmitTask(c.Request.Context(), &executor.UnifiedTask{
			Type:      consts.TaskTypeBuildDataset,
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

		traces = append(traces, dto.Trace{TraceID: traceID, HeadTaskID: taskID})
	}

	dto.JSONResponse(c, http.StatusAccepted, "Dataset building submitted successfully", dto.SubmitResp{GroupID: groupID, Traces: traces})
}

func UploadDataset(c *gin.Context) {
}
