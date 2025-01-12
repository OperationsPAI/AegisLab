package handlers

import "github.com/gin-gonic/gin"

// GetAlgorithmResp
// 去获取每个算法目录里的 toml 描述文件
type GetAlgorithmResp struct {
	Name string `json:"name"`
}

// GetAlgorithms
//
//	@Summary		获取算法列表
//	@Description	获取算法列表
//	@Tags			algorithm
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		InjectReq	true	"请求体"
//	@Success		200		{object}	GenericResponse[GetAlgorithmResp]
//	@Failure		400		{object}	GenericResponse[GetAlgorithmResp]
//	@Failure		500		{object}	GenericResponse[GetAlgorithmResp]
//	@Router			/api/v1/algo/injectstatus [post]
func GetAlgorithms(c *gin.Context) {
}
