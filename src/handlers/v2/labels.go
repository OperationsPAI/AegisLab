package v2

import (
	"net/http"
	"strings"

	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/utils"

	"github.com/gin-gonic/gin"
)

// CreateLabels creates a new label
//
//	@Summary Create label
//	@Description Create a new label with key-value pair. If a deleted label with same key-value exists, it will be restored and updated.
//	@Tags Labels
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param label body dto.LabelCreateReq true "Label creation request"
//	@Success 201 {object} dto.GenericResponse[dto.LabelResponse] "Label created successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 409 {object} dto.GenericResponse[any] "Label already exists"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/labels [post]
func CreateLabels(c *gin.Context) {
	var req dto.LabelCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	existing, err := repository.GetLabelByKeyandValue(req.Key, req.Value)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to check existing label: "+err.Error())
		return
	}

	if existing != nil {
		dto.ErrorResponse(c, http.StatusConflict, "Label with same key and value already exists")
		return
	}

	deletedLabel, err := repository.GetLabelByKeyandValue(req.Key, req.Value, consts.LabelDeleted)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to check deleted label: "+err.Error())
		return
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var label *database.Label

	if deletedLabel != nil {
		if removeFunc, exists := repository.RemoveRelationsFromLabel[deletedLabel.Category]; exists {
			if err := removeFunc(deletedLabel.ID); err != nil {
				tx.Rollback()
				dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove relations from deleted label: "+err.Error())
				return
			}
		}

		label = deletedLabel
		label.Category = req.Category
		label.Description = req.Description
		label.Color = utils.GetStringValue(req.Color, "#1890ff")

		if err := tx.Save(label).Error; err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to update deleted dataset: "+err.Error())
			return
		}
	} else {
		label = req.ToEntity()
		if err := tx.Create(label).Error; err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create label: "+err.Error())
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
		return
	}

	response := dto.ToLabelResponse(label)
	dto.JSONResponse(c, http.StatusCreated, "Label created successfully", response)
}
