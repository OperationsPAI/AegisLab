package handlers

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func HandleServiceError(c *gin.Context, err error) bool {
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
	case consts.ErrAuthenticationFailed:
		dto.ErrorResponse(c, http.StatusUnauthorized, msg)
	case consts.ErrNotFound:
		dto.ErrorResponse(c, http.StatusNotFound, msg)
	case consts.ErrAlreadyExists:
		dto.ErrorResponse(c, http.StatusConflict, msg)
	default:
		// Log full error details for debugging
		logrus.WithFields(logrus.Fields{
			"path":   c.Request.URL.Path,
			"method": c.Request.Method,
			"error":  err.Error(),
		}).Error("Service error")
		// Return user-friendly message but expose more details in development
		dto.ErrorResponse(c, http.StatusInternalServerError, msg)
	}

	return true
}
