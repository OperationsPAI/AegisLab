package handlers

import "github.com/gin-gonic/gin"

type GetDatasetReq struct {
	PageNumber int `json:"page_number"`
	PageSize   int `json:"page_size"`
}

// GetDatasetResp
type GetDatasetResp struct {
	Name string `json:"name"`
}

// GetDatasetList
//
//	@Summary		获取所有数据集列表
//	@Description	获取所有数据集列表
//	@Tags			algorithm
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		GetDatasetReq	true	"请求体"
//	@Success		200		{object}	GenericResponse[GetDatasetResp]
//	@Failure		400		{object}	GenericResponse[GetDatasetResp]
//	@Failure		500		{object}	GenericResponse[GetDatasetResp]
//	@Router			/api/v1/dataset/getlist [post]
func GetDatasetList(c *gin.Context) {
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
//	@Tags			algorithm
//	@Router			/api/v1/dataset/delete [post]
func DeleteDataset(c *gin.Context) {

}
