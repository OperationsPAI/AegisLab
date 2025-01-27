package handlers

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/CUHK-SE-Group/rcabench/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type DatasetReq struct {
	PageNum  *int `form:"page_num" binding:"required,min=1"`
	PageSize *int `form:"page_size" binding:"required,min=5,max=20"`
}

type DatasetResp struct {
	Total    int64    `json:"total"`
	Datasets []string `json:"datasets"`
}

type DatasetDownloadReq struct {
	Datasets []string `json:"datasets" binding:"required,min=1,dive,required"`
}

var DatasetStatusMap = map[int]string{
	executor.DatasetInitial: "initial",
	executor.DatasetSuccess: "success",
	executor.DatasetFailed:  "failed",
	executor.DatesetDeleted: "deleted",
}

var DatasetFieldMap = map[string]string{
	"PageNum":  "page_num",
	"PageSize": "page_size",
}

// GetDatasetList 获取数据集列表（带分页）
//
//	@Summary		分页查询数据集列表
//	@Description	获取状态为成功的注入数据集列表（支持分页参数）
//	@Tags			dataset
//	@Produce		application/json
//	@Param			page_num	query		int		true	"页码（从1开始）" minimum(1) default(1)
//	@Param			page_size	query		int		true	"每页数量" minimum(1) maximum(100) default(20)
//	@Success		200			{object}	GenericResponse[DatasetResp] "成功响应"
//	@Failure		400			{object}	GenericResponse[any] "参数校验失败"
//	@Failure		500			{object}	GenericResponse[any] "服务器内部错误"
//	@Router			/api/v1/datasets [get]
func GetDatasetList(c *gin.Context) {
	// 获取查询参数并校验是否合法
	var datasetReq DatasetReq
	if err := c.BindQuery(&datasetReq); err != nil {
		JSONResponse[interface{}](c, http.StatusBadRequest, executor.FormatErrorMessage(err, DatasetFieldMap), nil)
		return
	}

	// 计算偏移量
	pageNum := *datasetReq.PageNum
	pageSize := *datasetReq.PageSize
	offset := (pageNum - 1) * pageSize

	// 查询总记录数
	var total int64
	err := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Select("injection_name").
		Where("status = ?", executor.DatasetSuccess).
		Count(&total).Error
	if err != nil {
		logrus.Errorf("Failed to count fault injection schedules: %v", err)
		JSONResponse[interface{}](c, http.StatusInternalServerError, "Failed to retrieve datasets", nil)
		return
	}

	// 查询分页数据
	var faultRecords []database.FaultInjectionSchedule
	err = database.DB.
		Select("injection_name").
		Where("status = ?", executor.DatasetSuccess).
		Offset(offset).
		Limit(pageSize).
		Find(&faultRecords).Error
	if err != nil {
		logrus.Errorf("Failed to query fault injection schedules with pagination: %v", err)
		JSONResponse[interface{}](c, http.StatusInternalServerError, "Failed to retrieve datasets", nil)
		return
	}

	var datasetResp DatasetResp
	datasetResp.Total = total
	for _, record := range faultRecords {
		datasetResp.Datasets = append(datasetResp.Datasets, record.InjectionName)
	}

	JSONResponse(c, http.StatusOK, "OK", datasetResp)
}

// DownloadDataset 处理数据集下载请求
//
//	@Summary		下载数据集打包文件
//	@Description	将指定路径的多个数据集打包为 ZIP 文件下载（自动排除 result.csv 文件）
//	@Tags			dataset
//	@Produce		application/zip
//	@Consumes		application/json
//	@Param			body		body		executor.FaultInjectionPayload	true	"下载请求参数"
//	@Success		200			{string} 	binary 		"ZIP 文件流"
//	@Failure		400			{object}	GenericResponse[any] "参数绑定错误"
//	@Failure		403			{object}	GenericResponse[any] "非法路径访问"
//	@Failure		500			{object}	GenericResponse[any] "文件打包失败"
//	@Router			/api/v1/datasets/download [get]
func DownloadDataset(c *gin.Context) {
	var req DatasetDownloadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONResponse[interface{}](c, http.StatusBadRequest, executor.FormatErrorMessage(err, DatasetFieldMap), nil)
		return
	}

	excludeRules := []utils.ExculdeRule{{Pattern: "result.csv", IsGlob: false}}

	// 预先检查所有数据集路径合法性
	for _, dataset := range req.Datasets {
		workDir := filepath.Join(config.GetString("nfs.path"), dataset)
		if !utils.IsAllowedPath(workDir) {
			JSONResponse[interface{}](c, http.StatusForbidden, "Invalid path access", nil)
			return
		}
	}

	// 设置响应头
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=packages.zip")

	zipWriter := zip.NewWriter(c.Writer)
	defer zipWriter.Close()

	for _, dataset := range req.Datasets {
		workDir := filepath.Join(config.GetString("nfs.path"), dataset)

		err := filepath.WalkDir(workDir, func(path string, dir fs.DirEntry, err error) error {
			if err != nil || dir.IsDir() {
				return err
			}

			relPath, _ := filepath.Rel(workDir, path)
			fullRelPath := filepath.Join(filepath.Base(workDir), relPath)
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
			c.Error(fmt.Errorf("packaging failed: %v", err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
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
//	@Router			/api/v1/dataset/delete [delete]
func DeleteDataset(c *gin.Context) {
	idStr := c.Param("datasetID")
	if idStr == "" {
		JSONResponse[interface{}](c, http.StatusBadRequest, "Dataset id is required", nil)
		return
	}

	var id int
	var err error
	if id, err = strconv.Atoi(idStr); err != nil {
		JSONResponse[interface{}](c, http.StatusBadRequest, "Dataset id must be an integer", nil)
		return
	}

	var faultRecord database.FaultInjectionSchedule
	err = database.DB.
		Model(&faultRecord).
		Where("id = ?", id).
		Update("status", executor.DatesetDeleted).Error
	if err != nil {
		logrus.Errorf("Failed to update status to DatasetDeleted for dataset %d: %v", id, err)
		JSONResponse[interface{}](c, http.StatusInternalServerError, fmt.Sprintf("Failed to delete dataset %d", id), nil)
		return
	}

	JSONResponse[interface{}](c, http.StatusOK, "Delete dataset successfully", id)
}
