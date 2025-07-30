package v2

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/gin-gonic/gin"
)

// CreateDataset 创建数据集
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
	// Check permission
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canCreate, err := checker.CanWriteResource(consts.ResourceDataset)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}
	if !canCreate {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to create datasets")
		return
	}

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
		req.Format = "json"
	}

	// Check if dataset with same name and version already exists
	existing, err := repository.GetDatasetByNameAndVersion(req.Name, req.Version)
	if err == nil && existing != nil {
		dto.ErrorResponse(c, http.StatusConflict, "Dataset with same name and version already exists")
		return
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create dataset
	dataset := &database.Dataset{
		Name:        req.Name,
		Version:     req.Version,
		Description: req.Description,
		Type:        req.Type,
		DataSource:  req.DataSource,
		Format:      req.Format,
		ProjectID:   req.ProjectID,
		Status:      1, // Active
		IsPublic:    req.IsPublic != nil && *req.IsPublic,
	}

	if err := tx.Create(dataset).Error; err != nil {
		tx.Rollback()
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create dataset: "+err.Error())
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

	// Associate with injections if provided
	if len(req.InjectionIDs) > 0 {
		for _, injectionID := range req.InjectionIDs {
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
	if err := database.DB.Preload("Project").First(dataset, dataset.ID).Error; err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to load dataset: "+err.Error())
		return
	}

	response := dto.ToDatasetV2Response(dataset, false)
	dto.JSONResponse(c, http.StatusCreated, "Dataset created successfully", response)
}

// GetDataset 获取单个数据集
//
//	@Summary Get dataset by ID
//	@Description Get detailed information about a specific dataset
//	@Tags Datasets
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Dataset ID"
//	@Param include query string false "Include related data (project,injections,labels)"
//	@Success 200 {object} dto.GenericResponse[dto.DatasetV2Response] "Dataset retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid dataset ID"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Dataset not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/datasets/{id} [get]
func GetDataset(c *gin.Context) {
	// Check permission
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canRead, err := checker.CanReadResource(consts.ResourceDataset)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}
	if !canRead {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to read datasets")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	include := c.Query("include")
	includeInjections := strings.Contains(include, "injections")
	includeLabels := strings.Contains(include, "labels")

	// Build query with preloads
	query := database.DB.Model(&database.Dataset{})
	if strings.Contains(include, "project") {
		query = query.Preload("Project")
	}

	var dataset database.Dataset
	if err := query.First(&dataset, id).Error; err != nil {
		if strings.Contains(err.Error(), "not found") {
			dto.ErrorResponse(c, http.StatusNotFound, "Dataset not found")
		} else {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get dataset: "+err.Error())
		}
		return
	}

	response := dto.ToDatasetV2Response(&dataset, false)

	// Load injection relations if requested
	if includeInjections {
		relations, err := repository.GetDatasetFaultInjections(dataset.ID)
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to load injections: "+err.Error())
			return
		}

		response.Injections = make([]dto.DatasetV2InjectionRelationResponse, len(relations))
		for i, rel := range relations {
			response.Injections[i] = dto.DatasetV2InjectionRelationResponse{
				ID:               rel.ID,
				FaultInjectionID: rel.FaultInjectionID,
				CreatedAt:        rel.CreatedAt,
				UpdatedAt:        rel.UpdatedAt,
			}
		}
	}

	// Load label relations if requested
	if includeLabels {
		labels, err := repository.GetDatasetLabels(dataset.ID)
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to load labels: "+err.Error())
			return
		}
		response.Labels = labels
	}

	dto.SuccessResponse(c, response)
}

// ListDatasets 获取数据集列表
//
//	@Summary List datasets
//	@Description Get a paginated list of datasets with filtering and sorting
//	@Tags Datasets
//	@Produce json
//	@Security BearerAuth
//	@Param page query int false "Page number (default 1)"
//	@Param size query int false "Page size (default 20, max 100)"
//	@Param project_id query int false "Filter by project ID"
//	@Param type query string false "Filter by dataset type"
//	@Param status query int false "Filter by status"
//	@Param is_public query bool false "Filter by public status"
//	@Param search query string false "Search in name and description"
//	@Param sort_by query string false "Sort field (id,name,created_at,updated_at)"
//	@Param sort_order query string false "Sort order (asc,desc)"
//	@Param include query string false "Include related data (project,injections,labels)"
//	@Success 200 {object} dto.GenericResponse[dto.DatasetSearchResponse] "Datasets retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request parameters"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/datasets [get]
func ListDatasets(c *gin.Context) {
	// Check permission
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canRead, err := checker.CanReadResource(consts.ResourceDataset)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}
	if !canRead {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to read datasets")
		return
	}

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
	datasets, total, err := repository.ListDatasets(req.Page, req.Size, req.ProjectID, req.Type, req.Status, req.IsPublic)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to list datasets: "+err.Error())
		return
	}

	// Convert to response
	items := make([]dto.DatasetV2Response, len(datasets))
	includeLabels := strings.Contains(req.Include, "labels")
	for i, dataset := range datasets {
		items[i] = *dto.ToDatasetV2Response(&dataset, false)

		// Load labels if requested
		if includeLabels {
			labels, err := repository.GetDatasetLabels(dataset.ID)
			if err == nil {
				items[i].Labels = labels
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

// UpdateDataset 更新数据集
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
	// Check permission
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canUpdate, err := checker.CanWriteResource(consts.ResourceDataset)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}
	if !canUpdate {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to update datasets")
		return
	}

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
	if req.InjectionIDs != nil {
		// Remove existing associations
		if err := tx.Where("dataset_id = ?", dataset.ID).Delete(&database.DatasetFaultInjection{}).Error; err != nil {
			tx.Rollback()
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove existing associations: "+err.Error())
			return
		}

		// Add new associations
		for _, injectionID := range req.InjectionIDs {
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
	if err := database.DB.Preload("Project").First(dataset, dataset.ID).Error; err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to reload dataset: "+err.Error())
		return
	}

	response := dto.ToDatasetV2Response(dataset, false)
	dto.SuccessResponse(c, response)
}

// DeleteDataset 删除数据集（逻辑删除）
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
	// Check permission
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canDelete, err := checker.CanDeleteResource(consts.ResourceDataset)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}
	if !canDelete {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to delete datasets")
		return
	}

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

// SearchDatasets 搜索数据集（复杂查询）
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
	// Check permission
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canRead, err := checker.CanReadResource(consts.ResourceDataset)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}
	if !canRead {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to read datasets")
		return
	}

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
		var projectID *int
		var datasetType string
		var status *int
		if len(req.ProjectIDs) > 0 {
			projectID = &req.ProjectIDs[0]
		}
		if len(req.Types) > 0 {
			datasetType = req.Types[0]
		}
		if len(req.Statuses) > 0 {
			status = &req.Statuses[0]
		}

		datasets, total, err = repository.ListDatasets(req.Page, req.Size, projectID, datasetType, status, req.IsPublic)
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

// ManageDatasetInjections 管理数据集中的故障注入
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
	// Check permission
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canUpdate, err := checker.CanWriteResource(consts.ResourceDataset)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}
	if !canUpdate {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to update datasets")
		return
	}

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

// ManageDatasetLabels 管理数据集中的标签
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
	// Check permission
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canUpdate, err := checker.CanWriteResource(consts.ResourceDataset)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}
	if !canUpdate {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to update datasets")
		return
	}

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
