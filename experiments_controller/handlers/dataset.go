package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	chaosCli "github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type DatasetReq struct {
	PageNum  *int `form:"page_num" binding:"required,min=1"`
	PageSize *int `form:"page_size" binding:"required,min=1"`
}

type DatasetResp struct {
	Total    int64    `json:"total"`
	Datasets []string `json:"datasets"`
}

var fieldMap = map[string]string{
	"ID":       "datasetID",
	"PageNum":  "page_num",
	"PageSize": "page_size",
}

// GetDatasetList
//
//	@Summary		获取所有数据集列表
//	@Description	获取所有数据集列表
//	@Tags			dataset
//	@Produce		application/json
//	@Param			page_num	query		int		true	"页面数目"
//	@Param			page_size	query		int		true	"页面大小"
//	@Success		200			{object}	GenericResponse[GetDatasetResp]
//	@Failure		400			{object}	GenericResponse[any]
//	@Failure		500			{object}	GenericResponse[any]
//	@Router			/api/v1/dataset/getlist [get]
func GetDatasetList(c *gin.Context) {
	// 获取查询参数并校验是否合法
	var datasetReq DatasetReq
	if err := c.ShouldBindQuery(&datasetReq); err != nil {
		JSONResponse[interface{}](c, http.StatusBadRequest, convertValidationErrors(err, fieldMap), nil)
		return
	}

	// 计算偏移量
	pageNum := *datasetReq.PageNum
	pageSize := *datasetReq.PageSize
	offset := (pageNum - 1) * pageSize

	currentTime := time.Now()

	// 查询总记录数
	var total int64

	err := database.DB.
		Model(&database.FaultInjectionSchedule{}).
		Where("status = ?", database.DatasetSuccess).
		Where("proposed_end_time < ?", currentTime).
		Count(&total).Error
	if err != nil {
		logrus.Errorf("Failed to count fault injection schedules: %v", err)
		JSONResponse[interface{}](c, http.StatusInternalServerError, "Failed to retrieve datasets", nil)
		return
	}

	// 查询分页数据
	var faultRecords []database.FaultInjectionSchedule

	err = database.DB.
		Where("status IN ?", []int{database.DatasetInitial, database.DatasetSuccess}).
		Where("proposed_end_time < ?", currentTime).
		Offset(offset).
		Limit(pageSize).
		Find(&faultRecords).Error
	if err != nil {
		logrus.Errorf("Failed to query fault injection schedules with pagination: %v", err)
		JSONResponse[interface{}](c, http.StatusInternalServerError, "Failed to retrieve datasets", nil)
		return
	}

	// 用于存储最终成功的记录
	var successfulRecords []database.FaultInjectionSchedule

	for _, record := range faultRecords {
		datasetName := record.InjectionName
		var startTime, endTime time.Time

		// 如果状态为初始，查询 CRD 并更新记录
		if record.Status == database.DatasetInitial {
			startTime, endTime, err = chaosCli.QueryCRDByName("ts", datasetName)
			if err != nil {
				logrus.Errorf("Failed to QueryCRDByName for dataset %s: %v", datasetName, err)

				// 更新状态为失败
				if updateErr := database.DB.Model(&record).Where("injection_name = ?", datasetName).
					Update("status", database.DatasetFailed).Error; updateErr != nil {
					logrus.Errorf("Failed to update status to DatasetFailed for dataset %s: %v", datasetName, updateErr)
				}
				continue
			}

			// 更新数据库中的 start_time、end_time 和状态为成功
			if updateErr := database.DB.Model(&record).Where("injection_name = ?", datasetName).
				Updates(map[string]interface{}{
					"start_time": startTime,
					"end_time":   endTime,
					"status":     database.DatasetSuccess,
				}).Error; updateErr != nil {
				logrus.Errorf("Failed to update record for dataset %s: %v", datasetName, updateErr)
				continue
			}
			// 更新成功的记录状态到内存
			record.StartTime = startTime
			record.EndTime = endTime
			record.Status = database.DatasetSuccess
		}

		// 仅保留状态为成功的记录
		if record.Status == database.DatasetSuccess {
			successfulRecords = append(successfulRecords, record)
		}
	}

	var datasetResp DatasetResp
	datasetResp.Total = total
	for _, record := range successfulRecords {
		datasetResp.Datasets = append(datasetResp.Datasets, record.InjectionName)
	}

	JSONResponse(c, http.StatusOK, "OK", datasetResp)
}

// DownloadDataset
//
//	@Summary		下载数据集数据
//	@Description	下载数据集数据
//	@Tags			algorithm
//	@Router			/api/v1/dataset/download [post]
func DownloadDataset(c *gin.Context) {
}

// UploadDataset
//
//	@Summary		上传数据集数据
//	@Description	上传数据集数据
//	@Tags			algorithm
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
		Update("status", database.DatesetDeleted).Error
	if err != nil {
		logrus.Errorf("Failed to update status to DatasetDeleted for dataset %d: %v", id, err)
		JSONResponse[interface{}](c, http.StatusInternalServerError, fmt.Sprintf("Failed to delete dataset %d", id), nil)
		return
	}

	JSONResponse[interface{}](c, http.StatusOK, "Delete dataset successfully", id)
}
