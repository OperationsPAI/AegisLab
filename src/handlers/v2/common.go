package v2

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleServiceError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	processor := utils.NewErrorProcessor(err)
	innermostErr := processor.GetErrorByLevel(-1)
	if innermostErr == nil {
		return false
	}

	msg := innermostErr.Error()
	userFriendlyErr := processor.GetErrorByLevel(-2)
	if userFriendlyErr != nil {
		msg = userFriendlyErr.Error()
	}

	switch innermostErr {
	case consts.ErrNotFound:
		dto.ErrorResponse(c, http.StatusNotFound, msg)
	case consts.ErrAlreadyExists:
		dto.ErrorResponse(c, http.StatusConflict, msg)
	default:
		dto.ErrorResponse(c, http.StatusInternalServerError, consts.ErrInternal.Error())
	}

	return true
}
