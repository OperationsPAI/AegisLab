package handlers

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ParsePositiveID parses a string ID parameter and validates it's a positive integer
// Returns the parsed ID and true if valid, or writes error response and returns false
func ParsePositiveID(c *gin.Context, idStr, fieldName string) (int, bool) {
	logrus.WithFields(logrus.Fields{
		"idStr":     idStr,
		"fieldName": fieldName,
	}).Debug("ParsePositiveID: attempting to parse ID")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"idStr":     idStr,
			"fieldName": fieldName,
			"error":     err.Error(),
		}).Warn("ParsePositiveID: failed to parse ID as integer")
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid "+fieldName)
		return 0, false
	}

	if id <= 0 {
		logrus.WithFields(logrus.Fields{
			"id":        id,
			"fieldName": fieldName,
		}).Warn("ParsePositiveID: ID is not positive")
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid "+fieldName)
		return 0, false
	}

	logrus.WithFields(logrus.Fields{
		"id":        id,
		"fieldName": fieldName,
	}).Debug("ParsePositiveID: successfully parsed positive ID")
	return id, true
}

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
