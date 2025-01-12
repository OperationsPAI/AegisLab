package handlers

import (
	"time"

	"github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/gin-gonic/gin"
)

type InjectReq struct {
}
type InjectResp struct {
}
type InjectListResp struct {
	TaskID     string    `json:"task_id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	InjectTime time.Time `json:"inject_time"`
	Duration   int       `json:"duration"` // minutes
	FaultType  string    `json:"fault_type"`
	Para       string    `json:"para"`
}
type InjectStatusReq struct {
	TaskID string `json:"task_id"`
}
type InjectStatusResp struct {
}
type InjectCancelReq struct {
	TaskID string `json:"task_id"`
}
type InjectCancelResp struct {
}

// InjectFault
//
//	@Summary		注入故障
//	@Description	注入故障
//	@Tags			injection
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		InjectReq					true	"请求体"
//	@Success		200		{array}		GenericResponse[InjectResp]	""
//	@Failure		400		{object}	GenericResponse[InjectResp]	""
//	@Failure		500		{object}	GenericResponse[InjectResp]	""
//	@Router			/api/v1/injection/submit [post]
func InjectFault(c *gin.Context) {
}

// GetInjectionList
//
//	@Summary		获取注入列表和必要的简略信息
//	@Description	获取注入列表和必要的简略信息
//	@Tags			injection
//	@Produce		json
//	@Consumes		application/json
//	@Success		200	{array}		GenericResponse[[]InjectListResp]
//	@Failure		400	{object}	GenericResponse[[]InjectListResp]
//	@Failure		500	{object}	GenericResponse[[]InjectListResp]
//	@Router			/api/v1/injection/getlist [post]
func GetInjectionList(c *gin.Context) {
}

// GetInjectionStatus
//
//	@Summary		获取注入列表和必要的简略信息
//	@Description	获取注入列表和必要的简略信息
//	@Tags			injection
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		InjectStatusReq	true	"请求体"
//	@Success		200		{array}		GenericResponse[InjectStatusResp]
//	@Failure		400		{object}	GenericResponse[InjectStatusResp]
//	@Failure		500		{object}	GenericResponse[InjectStatusResp]
//	@Router			/api/v1/injection/getstatus [post]
func GetInjectionStatus(c *gin.Context) {
}

// CancelInjection
//
//	@Summary		取消注入
//	@Description	取消注入
//	@Tags			injection
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		InjectCancelReq	true	"请求体"
//	@Success		200		{array}		GenericResponse[InjectCancelResp]
//	@Failure		400		{object}	GenericResponse[InjectCancelResp]
//	@Failure		500		{object}	GenericResponse[InjectCancelResp]
//	@Router			/api/v1/injection/cancel [post]
func CancelInjection(c *gin.Context) {
}

// GetInjectionPara 获取注入参数
//
//	@Summary		获取故障注入参数
//	@Description	获取可用的故障注入参数和类型映射
//	@Tags			injection
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"返回故障注入参数和类型映射"
//	@Failure		500	{object}	map[string]string		"服务器内部错误"
//	@Router			/api/v1/injection/getpara [post]
func GetInjectionPara(c *gin.Context) {
	choice := make(map[string][]handler.ActionSpace, 0)
	for tp, spec := range handler.SpecMap {
		actionSpace, err := handler.GenerateActionSpace(spec)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate action space"})
			return
		}
		name := handler.GetChaosTypeName(tp)
		choice[name] = actionSpace
	}
	c.JSON(200, gin.H{
		"specification": choice,
		"keymap":        handler.ChaosTypeMap,
	})
}
