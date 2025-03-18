package handlers

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/executor"
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

	var payloads []executor.DatasetPayload
	if err := c.BindJSON(&payloads); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	logrus.Infof("Received building dataset payloads: %+v", payloads)

	var traces []dto.Trace
	for _, payload := range payloads {
		taskID, traceID, err := executor.SubmitTask(c.Request.Context(), &executor.UnifiedTask{
			Type:      executor.TaskTypeBuildDataset,
			Payload:   utils.StructToMap(payload),
			Immediate: true,
			GroupID:   groupID,
		})
		if err != nil {
			message := "Failed to submit task"
			logrus.Error(message)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		traces = append(traces, dto.Trace{TraceID: traceID, HeadTaskID: taskID})
	}

	dto.JSONResponse(c, http.StatusAccepted, "Dataset building submitted successfully", dto.SubmitResp{GroupID: groupID, Traces: traces})
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
		dto.ErrorResponse(c, http.StatusBadRequest, dto.FormatErrorMessage(err, dto.PaginationFieldMap))
		return
	}

	pageNum := *req.PageNum
	pageSize := *req.PageSize

	db := database.DB.Model(&database.FaultInjectionSchedule{}).Where("status = ?", executor.DatasetSuccess)
	db.Scopes(
		database.Sort("created_at desc"),
		database.Paginate(pageNum, pageSize),
	).Select("SQL_CALC_FOUND_ROWS *")

	// 查询总记录数
	var total int64
	if err := db.Raw("SELECT FOUND_ROWS()").Scan(&total).Error; err != nil {
		message := "Failed to count injection schedules"
		logrus.WithError(err).Error(message)
		dto.ErrorResponse(c, http.StatusInternalServerError, message)
		return
	}

	// 查询分页数据
	var records []database.FaultInjectionSchedule
	if err := db.Select("id, injection_name").Find(&records).Error; err != nil {
		message := "Failed to retrieve datasets"
		logrus.WithError(err).Error(message)
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
		dto.ErrorResponse(c, http.StatusBadRequest, dto.FormatErrorMessage(err, dto.PaginationFieldMap))
		return
	}

	var joinedResults []joinedResult
	if err := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Joins("JOIN tasks ON tasks.id = fault_injection_schedules.task_id").
		Where("tasks.group_id IN ? AND fault_injection_schedules.status = ?", req.GroupIDs, executor.DatasetSuccess).
		Select("tasks.group_id, fault_injection_schedules.injection_name").
		Scan(&joinedResults).
		Error; err != nil {
		message := "Failed to query datasets"
		logrus.WithError(err).Error(message)
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
				message := "Invalid path access"
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
				message := "Failed to packcage"
				logrus.WithError(err).Error(message)
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

	var existingIDs []int
	if err := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Select("id").
		Where("id IN ? AND status != ?", req.IDs, executor.DatesetDeleted).
		Pluck("id", &existingIDs).
		Error; err != nil {
		message := "Failed to query datasets"
		logrus.WithError(err).Error(message)
		dto.ErrorResponse(c, http.StatusInternalServerError, message)
		return
	}

	// 所有目标记录已被删除，直接返回成功
	if len(existingIDs) == 0 {
		dto.ErrorResponse(c, http.StatusOK, "No records to delete")
		return
	}

	if err := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Where("id IN ?", existingIDs).
		Update("status", executor.DatesetDeleted).
		Error; err != nil {
		message := "Failed to delete datasets"
		logrus.WithError(err).Error(message)
		dto.ErrorResponse(c, http.StatusInternalServerError, message)
		return
	}

	dto.SuccessResponse[any](c, nil)
}
