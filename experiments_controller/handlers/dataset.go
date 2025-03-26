package handlers

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/CUHK-SE-Group/rcabench/repository"
	"github.com/CUHK-SE-Group/rcabench/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type joinedResult struct {
	GroupID string `gorm:"column:group_id"`
	Dataset string `gorm:"column:injection_name"`
}

// BuildDataset
//
//	@Summary		制作数据集
//	@Description	制作数据集
//	@Tags			dataset
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		[]executor.DatasetPayload	true	"请求体"
//	@Success		200		{object}	GenericResponse[BuildResp]
//	@Failure		400		{object}	GenericResponse[any]
//	@Failure		500		{object}	GenericResponse[any]
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

	fiRecord, err := repository.GetInjectionRecordByDataset(req.Name)
	if err != nil {
		logrus.Errorf("failed to get fault injection record: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to retrieve injection data")
		return
	}

	logEntry := logrus.WithField("dataset", req.Name)

	meta, err := executor.ParseInjectionMeta(fiRecord.Config)
	if err != nil {
		logEntry.Errorf("failed to parse injection config: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "invalid injection configuration")
		return
	}

	param := dto.InjectionParam{
		Duration:  meta.Duration,
		FaultType: handler.ChaosTypeMap[handler.ChaosType(meta.FaultType)],
		Namespace: meta.Namespace,
		Pod:       meta.Pod,
		Spec:      meta.InjectSpec,
	}
	if fiRecord.Status != consts.DatasetBuildSuccess {
		dto.SuccessResponse(c, &dto.QueryDatasetResp{
			Param:            param,
			StartTime:        fiRecord.StartTime,
			EndTime:          fiRecord.EndTime,
			DetectorResult:   dto.DetectorRecord{},
			ExecutionResults: []dto.ExecutionRecord{},
		})

		return
	}

	detectorRecord, err := repository.GetDetectorRecordByDatasetID(fiRecord.ID)
	if err != nil {
		logEntry.Errorf("failed to retrieve detector record: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to load execution data")
		return
	}

	executionRecords, err := repository.GetExecutionRecordsByDatasetID(fiRecord.ID, sortOrder)
	if err != nil {
		logEntry.Error("Failed to retrieve execution records")
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to load execution data")
		return
	}

	if len(executionRecords) == 0 {
		logEntry.Info("No execution records found for dataset")
	}

	dto.SuccessResponse(c, &dto.QueryDatasetResp{
		Param:            param,
		StartTime:        fiRecord.StartTime,
		EndTime:          fiRecord.EndTime,
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
//	@Success		200			{object}	GenericResponse[DatasetResp] "成功响应"
//	@Failure		400			{object}	GenericResponse[any] "参数校验失败"
//	@Failure		500			{object}	GenericResponse[any] "服务器内部错误"
//	@Router			/api/v1/datasets [get]
func GetDatasetList(c *gin.Context) {
	// 获取查询参数并校验是否合法
	var req dto.DatasetListReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, formatErrorMessage(err, dto.PaginationFieldMap))
		return
	}

	pageNum := *req.PageNum
	pageSize := *req.PageSize

	db := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Where("status = ?", consts.DatasetBuildSuccess)
	db.Scopes(
		database.Sort("created_at desc"),
		database.Paginate(pageNum, pageSize),
	).Select("SQL_CALC_FOUND_ROWS *")

	// 查询总记录数
	var total int64
	if err := db.Raw("SELECT FOUND_ROWS()").Scan(&total).Error; err != nil {
		message := "failed to count injection schedules"
		logrus.Errorf("%s: %v", message, err)
		dto.ErrorResponse(c, http.StatusInternalServerError, message)
		return
	}

	// 查询分页数据
	var records []database.FaultInjectionSchedule
	if err := db.Select("id, injection_name").Find(&records).Error; err != nil {
		message := "failed to retrieve datasets"
		logrus.Errorf("%s: %v", message, err)
		dto.ErrorResponse(c, http.StatusInternalServerError, message)
		return
	}

	items := make([]dto.DatasetItem, 0, len(records))
	for _, record := range records {
		items = append(items, *dto.ConvertToDatasetItem(&record))
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
//	@Failure		400			{object}	GenericResponse[any] "参数绑定错误"
//	@Failure		403			{object}	GenericResponse[any] "非法路径访问"
//	@Failure		500			{object}	GenericResponse[any] "文件打包失败"
//	@Router			/api/v1/datasets/download [get]
func DownloadDataset(c *gin.Context) {
	filename := "package"
	excludeRules := []utils.ExculdeRule{{Pattern: "result.csv", IsGlob: false}}

	var req dto.DatasetDownloadReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, formatErrorMessage(err, dto.PaginationFieldMap))
		return
	}

	var joinedResults []joinedResult
	if err := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Joins("JOIN tasks ON tasks.id = fault_injection_schedules.task_id").
		Where("tasks.group_id IN ? AND fault_injection_schedules.status = ?", req.GroupIDs, consts.DatasetBuildSuccess).
		Select("tasks.group_id, fault_injection_schedules.injection_name").
		Scan(&joinedResults).
		Error; err != nil {
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
		groupedResults[joinedResult.GroupID] = append(groupedResults[joinedResult.GroupID], joinedResult.Dataset)
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

// UploadDataset
//
//	@Summary		上传数据集数据
//	@Description	上传数据集数据
//	@Tags			dataset
//	@Router			/api/v1/dataset/upload [post]
func UploadDataset(c *gin.Context) {
}

// DeleteDataset
//
//	@Summary		删除数据集数据
//	@Description	删除数据集数据
//	@Tags			dataset
//	@Produce		application/json
//	@Param			datasetID	path		int					true	"数据集 ID"
//	@Success		200			{object}	GenericResponse[int]
//	@Failure		400			{object}	GenericResponse[any]
//	@Failure		500			{object}	GenericResponse[any]
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
