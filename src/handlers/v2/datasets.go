package v2

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/gin-gonic/gin"
)

// CreateDataset creates a dataset
//
//	@Summary Create dataset
//	@Description Create a new dataset with optional injection and label associations
//	@Tags Datasets
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param dataset body dto.DatasetV2CreateReq true "Dataset creation request"
//	@Success 201 {object} dto.GenericResponse[dto.DatasetV2Response] "Dataset created successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 409 {object} dto.GenericResponse[any] "Dataset already exists"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/datasets [post]
func CreateDataset(c *gin.Context) {
	var req dto.DatasetV2CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Set defaults
	if req.Version == "" {
		req.Version = "v1.0"
	}
	if req.Format == "" {
		req.Format = "parquet"
	}

	// Check if dataset with same name and version already exists
	existing, err := repository.GetDatasetByNameAndVersion(req.Name, req.Version)
	if err == nil && existing != nil {
		dto.ErrorResponse(c, http.StatusConflict, "Dataset with same name and version already exists")
		return
	}

	// Check if there's a deleted dataset with same name and version (for recovery/overwrite)
	var deletedDataset *database.Dataset
	deletedDataset, err = repository.GetDeletedDatasetByNameAndVersion(req.Name, req.Version)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to check deleted dataset: "+err.Error())
		return
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var dataset *database.Dataset

	if deletedDataset != nil {
		// Overwrite existing deleted dataset
		dataset = deletedDataset
		dataset.Description = req.Description
		dataset.Type = req.Type
		dataset.DataSource = req.DataSource
		dataset.Format = req.Format
		dataset.Status = consts.DatasetEnabled
		dataset.IsPublic = req.IsPublic != nil && *req.IsPublic

		if err := tx.Save(dataset).Error; err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to update deleted dataset: "+err.Error())
			return
		}

		// Remove all existing associations for overwrite
		if err := repository.RemoveAllLabelsFromDataset(dataset.ID); err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove existing labels: "+err.Error())
			return
		}

		if err := repository.RemoveAllInjectionsFromDataset(dataset.ID); err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove existing injections: "+err.Error())
			return
		}
	} else {
		// Create new dataset
		dataset = &database.Dataset{
			Name:        req.Name,
			Version:     req.Version,
			Description: req.Description,
			Type:        req.Type,
			DataSource:  req.DataSource,
			Format:      req.Format,
			Status:      consts.DatasetEnabled,
			IsPublic:    req.IsPublic != nil && *req.IsPublic,
		}

		if err := tx.Create(dataset).Error; err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create dataset: "+err.Error())
			return
		}
	}

	// Create new labels if provided
	var createdLabelIDs []int
	for _, labelReq := range req.NewLabels {
		label := &database.Label{
			Key:         labelReq.Key,
			Value:       labelReq.Value,
			Category:    labelReq.Category,
			Description: labelReq.Description,
			Color:       labelReq.Color,
			IsSystem:    false,
			Usage:       1, // Initially used by this dataset
		}
		if label.Category == "" {
			label.Category = "dataset"
		}
		if label.Color == "" {
			label.Color = "#1890ff"
		}

		if err := tx.Create(label).Error; err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create label: "+err.Error())
			return
		}
		createdLabelIDs = append(createdLabelIDs, label.ID)
	}

	// Combine existing label IDs and newly created label IDs
	allLabelIDs := append(req.LabelIDs, createdLabelIDs...)

	// Associate with labels
	for _, labelID := range allLabelIDs {
		if err := repository.AddLabelToDataset(dataset.ID, labelID); err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to associate with labels: "+err.Error())
			return
		}
	}

	// Process injection references (ID or name) - batch query for better performance
	var injectionIDs []int
	if len(req.InjectionRefs) > 0 {
		// Collect IDs and names for batch query
		var ids []int
		var names []string
		for _, ref := range req.InjectionRefs {
			if ref.ID != nil {
				ids = append(ids, *ref.ID)
			} else if ref.Name != nil {
				names = append(names, *ref.Name)
			} else {
				tx.Rollback()
				dto.ErrorResponse(c, http.StatusBadRequest, "Injection reference must specify either ID or name")
				return
			}
		}

		// Batch query injections
		injections, err := repository.GetInjectionsByIDsAndNames(ids, names)
		if err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to query injections: "+err.Error())
			return
		}

		// Create maps for quick lookup
		idMap := make(map[int]*database.FaultInjectionSchedule)
		nameMap := make(map[string]*database.FaultInjectionSchedule)
		for i := range injections {
			injection := &injections[i]
			idMap[injection.ID] = injection
			nameMap[injection.InjectionName] = injection
		}

		// Process each reference and collect valid IDs
		for _, ref := range req.InjectionRefs {
			var injection *database.FaultInjectionSchedule
			var injectionID int

			if ref.ID != nil {
				injection = idMap[*ref.ID]
				if injection == nil {
					tx.Rollback()
					dto.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Injection with ID %d not found", *ref.ID))
					return
				}
				injectionID = injection.ID
			} else if ref.Name != nil {
				injection = nameMap[*ref.Name]
				if injection == nil {
					tx.Rollback()
					dto.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Injection with name '%s' not found", *ref.Name))
					return
				}
				injectionID = injection.ID
			}

			injectionIDs = append(injectionIDs, injectionID)
		}
	}

	// Associate with injections if provided
	if len(injectionIDs) > 0 {
		for _, injectionID := range injectionIDs {
			relation := &database.DatasetFaultInjection{
				DatasetID:        dataset.ID,
				FaultInjectionID: injectionID,
			}
			if err := tx.Create(relation).Error; err != nil {
				tx.Rollback()
				dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to associate with injections: "+err.Error())
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
		return
	}

	// Load relations for response
	if err := database.DB.First(dataset, dataset.ID).Error; err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to load dataset: "+err.Error())
		return
	}

	response := dto.ToDatasetV2Response(dataset, false)
	dto.JSONResponse(c, http.StatusCreated, "Dataset created successfully", response)
}

// GetDataset gets a single dataset
//
//	@Summary Get dataset by ID
//	@Description Get detailed information about a specific dataset
//	@Tags Datasets
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Dataset ID"
//	@Param include_injections query bool false "Include related fault injections"
//	@Param include_labels query bool false "Include related labels"
//	@Success 200 {object} dto.GenericResponse[dto.DatasetV2Response] "Dataset retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid dataset ID"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Dataset not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/datasets/{id} [get]
func GetDataset(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	var req dto.DatasetV2GetReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	// Use GORM preloading for simplicity and better performance in single dataset queries
	dataset, err := repository.GetDatasetByIDWithRelations(id, req.IncludeLabels, req.IncludeInjections)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			dto.ErrorResponse(c, http.StatusNotFound, "Dataset not found")
		} else {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get dataset: "+err.Error())
		}
		return
	}

	response := dto.ToDatasetV2Response(dataset, false)

	// Convert GORM associations to response format if loaded
	if req.IncludeLabels && len(dataset.Labels) > 0 {
		response.Labels = dataset.Labels
	}

	if req.IncludeInjections && len(dataset.FaultInjections) > 0 {
		items, err := toInjectionV2ResponsesWithLabels(dataset.FaultInjections, false)
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to convert injections: "+err.Error())
			return
		}
		response.Injections = items
	}

	dto.SuccessResponse(c, response)
}

// ListDatasets gets the dataset list
//
//	@Summary List datasets
//	@Description Get a paginated list of datasets with filtering and sorting
//	@Tags Datasets
//	@Produce json
//	@Security BearerAuth
//	@Param page query int false "Page number (default 1)"
//	@Param size query int false "Page size (default 20, max 100)"
//	@Param type query string false "Filter by dataset type"
//	@Param status query int false "Filter by status"
//	@Param is_public query bool false "Filter by public status"
//	@Param search query string false "Search in name and description"
//	@Param sort_by query string false "Sort field (id,name,created_at,updated_at)"
//	@Param sort_order query string false "Sort order (asc,desc)"
//	@Param include query string false "Include related data (injections,labels)"
//	@Success 200 {object} dto.GenericResponse[dto.DatasetSearchResponse] "Datasets retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request parameters"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/datasets [get]
func ListDatasets(c *gin.Context) {
	var req dto.DatasetV2ListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	// Call repository
	datasets, total, err := repository.ListDatasets(req.Page, req.Size, req.Type, req.Status, req.IsPublic)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to list datasets: "+err.Error())
		return
	}

	// Convert to response
	items := make([]dto.DatasetV2Response, len(datasets))
	includeLabels := strings.Contains(req.Include, "labels")
	includeInjections := strings.Contains(req.Include, "injections")

	// Batch load related data if requested for better performance
	var labelsMap map[int][]database.Label
	var injectionsMap map[int][]database.FaultInjectionSchedule

	if includeLabels || includeInjections {
		datasetIDs := make([]int, len(datasets))
		for i, dataset := range datasets {
			datasetIDs[i] = dataset.ID
		}

		// Load labels in batch if requested
		if includeLabels {
			labelsMap, err = repository.GetDatasetLabelsMap(datasetIDs)
			if err != nil {
				dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to load labels: "+err.Error())
				return
			}
		}

		// Load injections in batch if requested
		if includeInjections {
			injectionsMap, err = repository.GetDatasetInjectionsMap(datasetIDs)
			if err != nil {
				dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to load injections: "+err.Error())
				return
			}
		}
	}

	for i, dataset := range datasets {
		items[i] = *dto.ToDatasetV2Response(&dataset, false)

		// Add labels if requested and available
		if includeLabels {
			if labels, ok := labelsMap[dataset.ID]; ok {
				items[i].Labels = labels
			}
		}

		// Add injections if requested and available
		if includeInjections {
			if injections, ok := injectionsMap[dataset.ID]; ok {
				injectionItems, err := toInjectionV2ResponsesWithLabels(injections, false)
				if err == nil {
					items[i].Injections = injectionItems
				}
			}
		}
	}

	// Create pagination info
	pagination := dto.PaginationInfo{
		Page:       req.Page,
		Size:       req.Size,
		Total:      total,
		TotalPages: int((total + int64(req.Size) - 1) / int64(req.Size)),
	}

	response := dto.DatasetSearchResponse{
		Items:      items,
		Pagination: pagination,
	}

	dto.SuccessResponse(c, response)
}

// UpdateDataset updates a dataset
//
//	@Summary Update dataset
//	@Description Update dataset information, injection and label associations
//	@Tags Datasets
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Dataset ID"
//	@Param dataset body dto.DatasetV2UpdateReq true "Dataset update request"
//	@Success 200 {object} dto.GenericResponse[dto.DatasetV2Response] "Dataset updated successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Dataset not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/datasets/{id} [put]
func UpdateDataset(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	var req dto.DatasetV2UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Get existing dataset
	dataset, err := repository.GetDatasetByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			dto.ErrorResponse(c, http.StatusNotFound, "Dataset not found")
		} else {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get dataset: "+err.Error())
		}
		return
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update dataset fields
	if req.Name != nil {
		dataset.Name = *req.Name
	}
	if req.Version != nil {
		dataset.Version = *req.Version
	}
	if req.Description != nil {
		dataset.Description = *req.Description
	}
	if req.Type != nil {
		dataset.Type = *req.Type
	}
	if req.DataSource != nil {
		dataset.DataSource = *req.DataSource
	}
	if req.Format != nil {
		dataset.Format = *req.Format
	}
	if req.IsPublic != nil {
		dataset.IsPublic = *req.IsPublic
	}

	if err := tx.Save(dataset).Error; err != nil {
		tx.Rollback()
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to update dataset: "+err.Error())
		return
	}

	// Create new labels if provided
	var createdLabelIDs []int
	for _, labelReq := range req.NewLabels {
		label := &database.Label{
			Key:         labelReq.Key,
			Value:       labelReq.Value,
			Category:    labelReq.Category,
			Description: labelReq.Description,
			Color:       labelReq.Color,
			IsSystem:    false,
			Usage:       1,
		}
		if label.Category == "" {
			label.Category = "dataset"
		}
		if label.Color == "" {
			label.Color = "#1890ff"
		}

		if err := tx.Create(label).Error; err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create label: "+err.Error())
			return
		}
		createdLabelIDs = append(createdLabelIDs, label.ID)
	}

	// Update label associations if provided
	if req.LabelIDs != nil || len(createdLabelIDs) > 0 {
		// Remove existing label associations
		if err := tx.Where("dataset_id = ?", dataset.ID).Delete(&database.DatasetLabel{}).Error; err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove existing label associations: "+err.Error())
			return
		}

		// Add new label associations
		allLabelIDs := append(req.LabelIDs, createdLabelIDs...)
		for _, labelID := range allLabelIDs {
			datasetLabel := &database.DatasetLabel{
				DatasetID: dataset.ID,
				LabelID:   labelID,
			}
			if err := tx.Create(datasetLabel).Error; err != nil {
				tx.Rollback()
				dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create label association: "+err.Error())
				return
			}
		}
	}

	// Update injection associations if provided
	if req.InjectionRefs != nil {
		// Remove existing associations
		if err := tx.Where("dataset_id = ?", dataset.ID).Delete(&database.DatasetFaultInjection{}).Error; err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove existing associations: "+err.Error())
			return
		}

		// Process injection references (ID or name) - batch query for better performance
		var injectionIDs []int
		// Collect IDs and names for batch query
		var ids []int
		var names []string
		for _, ref := range req.InjectionRefs {
			if ref.ID != nil {
				ids = append(ids, *ref.ID)
			} else if ref.Name != nil {
				names = append(names, *ref.Name)
			} else {
				tx.Rollback()
				dto.ErrorResponse(c, http.StatusBadRequest, "Injection reference must specify either ID or name")
				return
			}
		}

		// Batch query injections
		injections, err := repository.GetInjectionsByIDsAndNames(ids, names)
		if err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to query injections: "+err.Error())
			return
		}

		// Create maps for quick lookup
		idMap := make(map[int]*database.FaultInjectionSchedule)
		nameMap := make(map[string]*database.FaultInjectionSchedule)
		for i := range injections {
			injection := &injections[i]
			idMap[injection.ID] = injection
			nameMap[injection.InjectionName] = injection
		}

		// Process each reference and collect valid IDs
		for _, ref := range req.InjectionRefs {
			var injection *database.FaultInjectionSchedule
			var injectionID int

			if ref.ID != nil {
				injection = idMap[*ref.ID]
				if injection == nil {
					tx.Rollback()
					dto.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Injection with ID %d not found", *ref.ID))
					return
				}
				injectionID = injection.ID
			} else if ref.Name != nil {
				injection = nameMap[*ref.Name]
				if injection == nil {
					tx.Rollback()
					dto.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Injection with name '%s' not found", *ref.Name))
					return
				}
				injectionID = injection.ID
			}

			injectionIDs = append(injectionIDs, injectionID)
		}

		// Add new associations
		for _, injectionID := range injectionIDs {
			relation := &database.DatasetFaultInjection{
				DatasetID:        dataset.ID,
				FaultInjectionID: injectionID,
			}
			if err := tx.Create(relation).Error; err != nil {
				tx.Rollback()
				dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create new associations: "+err.Error())
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
		return
	}

	// Reload dataset with relations
	if err := database.DB.First(dataset, dataset.ID).Error; err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to reload dataset: "+err.Error())
		return
	}

	response := dto.ToDatasetV2Response(dataset, false)
	dto.SuccessResponse(c, response)
}

// DeleteDataset deletes a dataset (soft delete)
//
//	@Summary Delete dataset
//	@Description Soft delete a dataset (sets status to -1)
//	@Tags Datasets
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Dataset ID"
//	@Success 200 {object} dto.GenericResponse[any] "Dataset deleted successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid dataset ID"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Dataset not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/datasets/{id} [delete]
func DeleteDataset(c *gin.Context) {

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	// Check if dataset exists
	if _, err := repository.GetDatasetByID(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			dto.ErrorResponse(c, http.StatusNotFound, "Dataset not found")
		} else {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get dataset: "+err.Error())
		}
		return
	}

	// Soft delete
	if err := repository.DeleteDataset(id); err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete dataset: "+err.Error())
		return
	}

	dto.SuccessResponse(c, gin.H{"message": "Dataset deleted successfully"})
}

// SearchDatasets advanced search for datasets
//
//	@Summary Search datasets
//	@Description Advanced search for datasets with complex filtering
//	@Tags Datasets
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param search body dto.DatasetV2SearchReq true "Search criteria"
//	@Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.DatasetV2Response]] "Search results"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/datasets/search [post]
func SearchDatasets(c *gin.Context) {
	var req dto.DatasetV2SearchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}

	var datasets []database.Dataset
	var total int64
	var err error

	// Search by labels if provided
	if len(req.LabelKeys) > 0 || len(req.LabelValues) > 0 {
		datasets, err = repository.SearchDatasetsByLabels(req.LabelKeys, req.LabelValues)
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to search datasets by labels: "+err.Error())
			return
		}
		total = int64(len(datasets))

		// Apply pagination manually
		start := (req.Page - 1) * req.Size
		end := start + req.Size
		if start >= len(datasets) {
			datasets = []database.Dataset{}
		} else if end > len(datasets) {
			datasets = datasets[start:]
		} else {
			datasets = datasets[start:end]
		}
	} else {
		// Use regular search
		var datasetType string
		var status *int
		if len(req.Types) > 0 {
			datasetType = req.Types[0]
		}
		if len(req.Statuses) > 0 {
			status = &req.Statuses[0]
		}

		datasets, total, err = repository.ListDatasets(req.Page, req.Size, datasetType, status, req.IsPublic)
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to search datasets: "+err.Error())
			return
		}
	}

	// Convert to response
	items := make([]dto.DatasetV2Response, len(datasets))
	for i, dataset := range datasets {
		items[i] = *dto.ToDatasetV2Response(&dataset, false)
	}

	// Create response using existing SearchResponse
	pagination := dto.PaginationInfo{
		Page:       req.Page,
		Size:       req.Size,
		Total:      total,
		TotalPages: int((total + int64(req.Size) - 1) / int64(req.Size)),
	}

	response := dto.SearchResponse[dto.DatasetV2Response]{
		Items:      items,
		Pagination: pagination,
	}

	dto.SuccessResponse(c, response)
}

// ManageDatasetInjections manages dataset injections
//
//	@Summary Manage dataset injections
//	@Description Add or remove injection associations for a dataset
//	@Tags Datasets
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Dataset ID"
//	@Param manage body dto.DatasetV2InjectionManageReq true "Injection management request"
//	@Success 200 {object} dto.GenericResponse[dto.DatasetV2Response] "Injections managed successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Dataset not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/datasets/{id}/injections [patch]
func ManageDatasetInjections(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	var req dto.DatasetV2InjectionManageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Check if dataset exists
	dataset, err := repository.GetDatasetByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			dto.ErrorResponse(c, http.StatusNotFound, "Dataset not found")
		} else {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get dataset: "+err.Error())
		}
		return
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Remove injections
	for _, injectionID := range req.RemoveInjections {
		if err := tx.Where("dataset_id = ? AND fault_injection_id = ?", id, injectionID).
			Delete(&database.DatasetFaultInjection{}).Error; err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove injection: "+err.Error())
			return
		}
	}

	// Add injections
	for _, injectionID := range req.AddInjections {
		relation := &database.DatasetFaultInjection{
			DatasetID:        id,
			FaultInjectionID: injectionID,
		}
		if err := tx.Create(relation).Error; err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to add injection: "+err.Error())
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
		return
	}

	// Return updated dataset
	response := dto.ToDatasetV2Response(dataset, false)
	dto.SuccessResponse(c, response)
}

// ManageDatasetLabels manages dataset labels
//
//	@Summary Manage dataset labels
//	@Description Add, remove labels or create new labels for a dataset
//	@Tags Datasets
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Dataset ID"
//	@Param manage body dto.DatasetV2LabelManageReq true "Label management request"
//	@Success 200 {object} dto.GenericResponse[dto.DatasetV2Response] "Labels managed successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Dataset not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/datasets/{id}/labels [patch]
func ManageDatasetLabels(c *gin.Context) {

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	var req dto.DatasetV2LabelManageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Check if dataset exists
	dataset, err := repository.GetDatasetByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			dto.ErrorResponse(c, http.StatusNotFound, "Dataset not found")
		} else {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get dataset: "+err.Error())
		}
		return
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Remove labels
	for _, labelID := range req.RemoveLabels {
		if err := repository.RemoveLabelFromDataset(id, labelID); err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove label: "+err.Error())
			return
		}
	}

	// Create new labels
	var createdLabelIDs []int
	for _, labelReq := range req.NewLabels {
		label := &database.Label{
			Key:         labelReq.Key,
			Value:       labelReq.Value,
			Category:    labelReq.Category,
			Description: labelReq.Description,
			Color:       labelReq.Color,
			IsSystem:    false,
			Usage:       1,
		}
		if label.Category == "" {
			label.Category = "dataset"
		}
		if label.Color == "" {
			label.Color = "#1890ff"
		}

		if err := tx.Create(label).Error; err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create label: "+err.Error())
			return
		}
		createdLabelIDs = append(createdLabelIDs, label.ID)
	}

	// Add existing and new labels
	allLabelIDs := append(req.AddLabels, createdLabelIDs...)
	for _, labelID := range allLabelIDs {
		if err := repository.AddLabelToDataset(id, labelID); err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to add label: "+err.Error())
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
		return
	}

	// Return updated dataset with labels
	response := dto.ToDatasetV2Response(dataset, false)

	// Load labels
	labels, err := repository.GetDatasetLabels(id)
	if err == nil {
		response.Labels = labels
	}

	dto.SuccessResponse(c, response)
}
