package v2

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/LGU-SE-Internal/rcabench/client"
	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/executor"
	"github.com/LGU-SE-Internal/rcabench/middleware"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/LGU-SE-Internal/rcabench/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SearchContainers handles complex container search with advanced filtering
//
//	@Summary Search containers
//	@Description Search containers with complex filtering, sorting and pagination. Supports all container types (algorithm, benchmark, etc.)
//	@Tags Containers
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.ContainerSearchRequest true "Container search request"
//	@Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.ContainerResponse]] "Containers retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers/search [post]
func SearchContainers(c *gin.Context) {
	// Check permission first
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canRead, err := checker.CanReadResource(consts.ResourceContainer)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}

	if !canRead {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to read containers")
		return
	}

	var req dto.ContainerSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Convert to SearchRequest
	searchReq := req.ConvertToSearchRequest()

	// Validate search request
	if err := searchReq.ValidateSearchRequest(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid search parameters: "+err.Error())
		return
	}

	// Execute search using query builder
	searchResult, err := repository.ExecuteSearch(database.DB, searchReq, database.Container{})
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to search containers: "+err.Error())
		return
	}

	// Convert database containers to response DTOs
	var containerResponses []dto.ContainerResponse
	for _, container := range searchResult.Items {
		containerResponse := dto.ContainerResponse{
			ID:        container.ID,
			Name:      container.Name,
			Type:      container.Type,
			Image:     container.Image,
			Tag:       container.Tag,
			Command:   container.Command,
			EnvVars:   container.EnvVars,
			UserID:    container.UserID,
			IsPublic:  container.IsPublic,
			Status:    container.Status,
			CreatedAt: container.CreatedAt,
			UpdatedAt: container.UpdatedAt,
		}

		containerResponses = append(containerResponses, containerResponse)
	}

	// Build final response
	response := dto.SearchResponse[dto.ContainerResponse]{
		Items:      containerResponses,
		Pagination: searchResult.Pagination,
		Filters:    searchResult.Filters,
		Sort:       searchResult.Sort,
	}

	dto.SuccessResponse(c, response)
}

// ListContainers handles simple container listing
//
//	@Summary List containers
//	@Description Get a simple list of containers with basic filtering
//	@Tags Containers
//	@Produce json
//	@Security BearerAuth
//	@Param page query int false "Page number" default(1)
//	@Param size query int false "Page size" default(20)
//	@Param type query string false "Container type filter" Enums(algorithm,benchmark)
//	@Param status query bool false "Container status filter"
//	@Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.ContainerResponse]] "Containers retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers [get]
func ListContainers(c *gin.Context) {
	// Check permission first
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canRead, err := checker.CanReadResource(consts.ResourceContainer)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}

	if !canRead {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to read containers")
		return
	}

	// Create a basic search request from query parameters
	req := dto.ContainerSearchRequest{
		AdvancedSearchRequest: dto.AdvancedSearchRequest{
			SearchRequest: dto.SearchRequest{
				Page: 1,
				Size: 20,
			},
		},
	}

	// Parse pagination from query parameters
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := parseIntParam(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}
	if sizeStr := c.Query("size"); sizeStr != "" {
		if size, err := parseIntParam(sizeStr); err == nil && size > 0 && size <= 1000 {
			req.Size = size
		}
	}

	// Parse filters from query parameters
	if containerType := c.Query("type"); containerType != "" {
		req.Type = &containerType
	}
	if statusStr := c.Query("status"); statusStr != "" {
		if status, err := strconv.ParseBool(statusStr); err == nil {
			req.IsActive = &status
		}
	}

	// Convert to SearchRequest
	searchReq := req.ConvertToSearchRequest()

	// Add default sorting by name
	searchReq.AddSort("name", dto.SortASC)

	// Validate search request
	if err := searchReq.ValidateSearchRequest(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid search parameters: "+err.Error())
		return
	}

	// Execute search using query builder
	searchResult, err := repository.ExecuteSearch(database.DB, searchReq, database.Container{})
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get container list: "+err.Error())
		return
	}

	// Convert database containers to response DTOs
	var containerResponses []dto.ContainerResponse
	for _, container := range searchResult.Items {
		containerResponse := dto.ContainerResponse{
			ID:        container.ID,
			Name:      container.Name,
			Type:      container.Type,
			Image:     container.Image,
			Tag:       container.Tag,
			Command:   container.Command,
			EnvVars:   container.EnvVars,
			UserID:    container.UserID,
			IsPublic:  container.IsPublic,
			Status:    container.Status,
			CreatedAt: container.CreatedAt,
			UpdatedAt: container.UpdatedAt,
		}

		containerResponses = append(containerResponses, containerResponse)
	}

	// Build final response
	response := dto.SearchResponse[dto.ContainerResponse]{
		Items:      containerResponses,
		Pagination: searchResult.Pagination,
		Filters:    searchResult.Filters,
		Sort:       searchResult.Sort,
	}

	dto.SuccessResponse(c, response)
}

// GetContainer handles getting a single container by ID
//
//	@Summary Get container by ID
//	@Description Get detailed information about a specific container
//	@Tags Containers
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Container ID"
//	@Success 200 {object} dto.GenericResponse[dto.ContainerResponse] "Container retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid container ID"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Container not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers/{id} [get]
func GetContainer(c *gin.Context) {
	// Check permission first
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canRead, err := checker.CanReadResource(consts.ResourceContainer)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}

	if !canRead {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to read containers")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	var container database.Container
	if err := database.DB.First(&container, id).Error; err != nil {
		if err.Error() == "record not found" {
			dto.ErrorResponse(c, http.StatusNotFound, "Container not found")
		} else {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get container: "+err.Error())
		}
		return
	}

	response := dto.ContainerResponse{
		ID:        container.ID,
		Name:      container.Name,
		Type:      container.Type,
		Image:     container.Image,
		Tag:       container.Tag,
		Command:   container.Command,
		EnvVars:   container.EnvVars,
		UserID:    container.UserID,
		IsPublic:  container.IsPublic,
		Status:    container.Status,
		CreatedAt: container.CreatedAt,
		UpdatedAt: container.UpdatedAt,
	}

	dto.SuccessResponse(c, response)
}

// CreateContainer handles container creation for v2 API
//
//	@Summary Create or update container
//	@Description Create a new container with build configuration or update existing one if it already exists. Containers are associated with the authenticated user. If a container with the same name, type, image, and tag already exists, it will be updated instead of creating a new one.
//	@Tags Containers
//	@Accept multipart/form-data
//	@Produce json
//	@Security BearerAuth
//	@Param type formData string true "Container type" Enums(algorithm,benchmark) default(algorithm)
//	@Param name formData string true "Container name"
//	@Param image formData string true "Docker image name"
//	@Param tag formData string false "Docker image tag" default(latest)
//	@Param command formData string false "Container startup command" default(/bin/bash)
//	@Param env_vars formData []string false "Environment variables (can be specified multiple times)"
//	@Param is_public formData boolean false "Whether the container is public" default(false)
//	@Param build_source_type formData string false "Build source type" Enums(file,github,harbor) default(file)
//	@Param file formData file false "Source code file (zip or tar.gz format, max 5MB) - required when build_source_type=file"
//	@Param github_repository formData string false "GitHub repository (owner/repo) - required when build_source_type=github"
//	@Param github_branch formData string false "GitHub branch name" default(main)
//	@Param github_commit formData string false "GitHub commit hash (if specified, branch is ignored)"
//	@Param github_path formData string false "Path within repository" default(.)
//	@Param github_token formData string false "GitHub access token for private repositories"
//	@Param harbor_image formData string false "Harbor image name - required when build_source_type=harbor"
//	@Param harbor_tag formData string false "Harbor image tag - required when build_source_type=harbor"
//	@Param context_dir formData string false "Docker build context directory" default(.)
//	@Param dockerfile_path formData string false "Dockerfile path relative to source root" default(Dockerfile)
//	@Success 202 {object} dto.GenericResponse[dto.SubmitResp] "Container creation/update task submitted successfully"
//	@Success 200 {object} dto.GenericResponse[dto.SubmitResp] "Container information updated successfully from Harbor"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers [post]
func CreateContainer(c *gin.Context) {
	// Check authentication
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Check permission
	checker := repository.NewPermissionChecker(userID.(int), nil)
	canCreate, err := checker.CanWriteResource(consts.ResourceContainer)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}

	if !canCreate {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to create containers")
		return
	}

	// Parse multipart form
	err = c.Request.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Failed to parse multipart form: "+err.Error())
		return
	}

	// Build CreateContainerRequest from form fields
	req := dto.CreateContainerRequest{
		ContainerType: consts.ContainerType(c.PostForm("type")),
		Name:          c.PostForm("name"),
		Image:         c.PostForm("image"),
		Tag:           c.DefaultPostForm("tag", "latest"),
		Command:       c.DefaultPostForm("command", "/bin/bash"),
		EnvVars:       c.PostFormArray("env_vars"),
		IsPublic:      c.PostForm("is_public") == "true",
	}

	// Validate required fields
	if req.ContainerType == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "Missing required field: type")
		return
	}
	if req.Name == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "Missing required field: name")
		return
	}
	if req.Image == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "Missing required field: image")
		return
	}

	// Handle build source
	buildSourceType := c.DefaultPostForm("build_source_type", "file")
	if buildSourceType != "" {
		req.BuildSource = &dto.BuildSource{
			Type: consts.BuildSourceType(buildSourceType),
		}

		switch consts.BuildSourceType(buildSourceType) {
		case consts.BuildSourceTypeFile:
			req.BuildSource.File = &dto.FileSource{}
		case consts.BuildSourceTypeGitHub:
			repository := c.PostForm("github_repository")
			if repository == "" && buildSourceType == "github" {
				dto.ErrorResponse(c, http.StatusBadRequest, "github_repository is required when build_source_type is github")
				return
			}
			req.BuildSource.GitHub = &dto.GitHubSource{
				Repository: repository,
				Branch:     c.DefaultPostForm("github_branch", "main"),
				Commit:     c.PostForm("github_commit"),
				Path:       c.DefaultPostForm("github_path", "."),
				Token:      c.PostForm("github_token"),
			}
		case consts.BuildSourceTypeHarbor:
			harborImage := c.PostForm("harbor_image")
			harborTag := c.PostForm("harbor_tag")
			if harborImage == "" && buildSourceType == "harbor" {
				dto.ErrorResponse(c, http.StatusBadRequest, "harbor_image is required when build_source_type is harbor")
				return
			}
			if harborTag == "" && buildSourceType == "harbor" {
				dto.ErrorResponse(c, http.StatusBadRequest, "harbor_tag is required when build_source_type is harbor")
				return
			}
			req.BuildSource.Harbor = &dto.HarborSource{
				Image: harborImage,
				Tag:   harborTag,
			}
		}
	}

	// Handle build options
	contextDir := c.PostForm("context_dir")
	dockerfilePath := c.PostForm("dockerfile_path")
	if contextDir != "" || dockerfilePath != "" {
		req.BuildOptions = &dto.BuildOptions{
			ContextDir:     c.DefaultPostForm("context_dir", "."),
			DockerfilePath: c.DefaultPostForm("dockerfile_path", "Dockerfile"),
		}
	}

	// Validate request
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	// Check if container with same name, type, image, tag already exists for this user
	var existingContainer database.Container
	result := database.DB.Where("name = ? AND type = ? AND image = ? AND tag = ? AND user_id = ? AND status = true",
		req.Name, req.ContainerType, req.Image, req.Tag, userID).First(&existingContainer)

	var isUpdate bool
	if result.Error == nil {
		// Container already exists, we'll update it
		isUpdate = true
		logrus.WithFields(logrus.Fields{
			"user_id": userID,
			"name":    req.Name,
			"type":    req.ContainerType,
			"image":   req.Image,
			"tag":     req.Tag,
		}).Info("Container already exists, will be updated")
	}

	// Handle different source types
	var code int
	var sourcePath string
	switch req.BuildSource.Type {
	case consts.BuildSourceTypeFile:
		code, sourcePath, err = processFileSourceV2(c, &req)
		if err != nil {
			dto.ErrorResponse(c, code, err.Error())
			return
		}
	case consts.BuildSourceTypeGitHub:
		code, sourcePath, err = processGitHubSourceV2(&req)
		if err != nil {
			dto.ErrorResponse(c, code, err.Error())
			return
		}
	case consts.BuildSourceTypeHarbor:
		code, err = processHarborSourceV2(&req, userID.(int))
		if err != nil {
			dto.ErrorResponse(c, code, err.Error())
			return
		}
		sourcePath = ""
	default:
		dto.ErrorResponse(c, http.StatusBadRequest, "Unsupported source type")
		return
	}

	// For non-Harbor sources, validate info content and required files
	if req.BuildSource.Type != consts.BuildSourceTypeHarbor {
		if code, err := req.ValidateInfoContent(sourcePath); err != nil {
			dto.ErrorResponse(c, code, err.Error())
			return
		}

		if err := validateRequiredFilesV2(sourcePath, req.BuildOptions); err != nil {
			dto.ErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}
	}

	// Get span context for tracing
	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get span context")
		return
	}
	spanCtx := ctx.(context.Context)

	// Handle Harbor direct update
	if req.BuildSource.Type == consts.BuildSourceTypeHarbor {
		taskID, traceID, err := processHarborDirectUpdateV2(spanCtx, &req, userID.(int), isUpdate)
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to process harbor direct update")
			return
		}

		message := "Container information created successfully from Harbor"
		if isUpdate {
			message = "Container information updated successfully from Harbor (existing record was overwritten)"
		}

		dto.JSONResponse(c, http.StatusOK,
			message,
			dto.SubmitResp{Traces: []dto.Trace{{TraceID: traceID, HeadTaskID: taskID, Index: 0}}},
		)
		return
	}

	// Submit build task
	task := &dto.UnifiedTask{
		Type: consts.TaskTypeBuildImage,
		Payload: map[string]any{
			consts.BuildContainerType: req.ContainerType,
			consts.BuildName:          req.Name,
			consts.BuildImage:         req.Image,
			consts.BuildTag:           req.Tag,
			consts.BuildCommand:       req.Command,
			consts.BuildImageEnvVars:  req.EnvVars,
			consts.BuildSourcePath:    sourcePath,
			consts.BuildBuildOptions:  req.BuildOptions,
		},
		Immediate: true,
		ProjectID: nil, // No project association for user containers
	}
	task.SetGroupCtx(spanCtx)

	taskID, traceID, err := executor.SubmitTask(spanCtx, task)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to submit container building task")
		return
	}

	message := "Container building task submitted successfully"
	if isUpdate {
		message = "Container building task submitted successfully (existing record will be overwritten)"
	}

	dto.JSONResponse(c, http.StatusAccepted,
		message,
		dto.SubmitResp{Traces: []dto.Trace{{TraceID: traceID, HeadTaskID: taskID, Index: 0}}},
	)
}

// Helper functions for v2 container creation

func processFileSourceV2(c *gin.Context, req *dto.CreateContainerRequest) (int, string, error) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		return http.StatusBadRequest, "", fmt.Errorf("failed to get uploaded file: %v", err)
	}
	defer file.Close()

	const maxSize = 5 * 1024 * 1024 // 5MB
	if header.Size > maxSize {
		return http.StatusBadRequest, "", fmt.Errorf("file size exceeds %dMB limit", maxSize/(1024*1024))
	}

	req.BuildSource.File = &dto.FileSource{
		Filename: header.Filename,
		Size:     header.Size,
	}

	fileName := header.Filename
	fileExt := strings.ToLower(filepath.Ext(fileName))

	isZip := fileExt == ".zip"
	isTarGz := (fileExt == ".gz" && strings.HasSuffix(strings.ToLower(fileName), ".tar.gz")) ||
		(fileExt == ".tgz")

	if !isZip && !isTarGz {
		return http.StatusBadRequest, "", fmt.Errorf("only zip and tar.gz files are allowed")
	}

	tempDir, err := os.MkdirTemp("", "container-upload-*")
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("failed to create temporary directory: %v", err)
	}

	filePath := filepath.Join(tempDir, header.Filename)
	out, err := os.Create(filePath)
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("failed to save file: %v", err)
	}

	targetDir := filepath.Join(config.GetString("container.storage_path"), string(req.ContainerType), req.Name, fmt.Sprintf("build_%d", time.Now().Unix()))
	if err = os.MkdirAll(targetDir, 0755); err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("failed to create target directory: %v", err)
	}

	var extractErr error
	if isZip {
		extractErr = utils.ExtractZip(filePath, targetDir)
	} else if isTarGz {
		extractErr = utils.ExtractTarGz(filePath, targetDir)
	}

	if extractErr != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("failed to extract file: %v", extractErr)
	}

	if err := os.RemoveAll(tempDir); err != nil {
		// Log warning but don't fail the request
		logrus.WithField("temp_dir", tempDir).Warnf("failed to remove temporary directory %s: %v", tempDir, err)
	}

	return 0, targetDir, nil
}

func processGitHubSourceV2(req *dto.CreateContainerRequest) (int, string, error) {
	github := req.BuildSource.GitHub

	targetDir := filepath.Join(config.GetString("container.storage_path"), string(req.ContainerType), req.Name, fmt.Sprintf("build_%d", time.Now().Unix()))
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("failed to create target directory: %v", err)
	}

	var gitCmd []string
	repoURL := fmt.Sprintf("https://github.com/%s.git", github.Repository)

	if github.Token != "" {
		repoURL = fmt.Sprintf("https://%s@github.com/%s.git", github.Token, github.Repository)
	}

	gitCmd = []string{"git", "clone"}
	if github.Commit != "" {
		gitCmd = append(gitCmd, repoURL, targetDir)
	} else {
		gitCmd = append(gitCmd, "--branch", github.Branch, "--single-branch", repoURL, targetDir)
	}

	if github.Commit != "" {
		cmd := exec.Command(gitCmd[0], gitCmd[1:]...)
		if err := cmd.Run(); err != nil {
			return http.StatusInternalServerError, "", fmt.Errorf("failed to clone repository: %v", err)
		}

		// Checkout specific commit
		cmd = exec.Command("git", "-C", targetDir, "checkout", github.Commit)
		if err := cmd.Run(); err != nil {
			return http.StatusBadRequest, "", fmt.Errorf("failed to checkout commit %s: %v", github.Commit, err)
		}
	} else {
		cmd := exec.Command(gitCmd[0], gitCmd[1:]...)
		if err := cmd.Run(); err != nil {
			return http.StatusInternalServerError, "", fmt.Errorf("failed to clone repository: %v", err)
		}
	}

	// If a specific path is provided, copy only that subdirectory
	if github.Path != "" && github.Path != "." {
		sourcePath := filepath.Join(targetDir, github.Path)
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			return http.StatusNotFound, "", fmt.Errorf("specified path '%s' does not exist in repository", github.Path)
		}

		newTargetDir := filepath.Join(config.GetString("container.storage_path"), string(req.ContainerType), req.Name, fmt.Sprintf("build_final_%d", time.Now().Unix()))
		if err := utils.CopyDir(sourcePath, newTargetDir); err != nil {
			return http.StatusInternalServerError, "", fmt.Errorf("failed to copy subdirectory: %v", err)
		}

		// Clean up the full clone
		if err := os.RemoveAll(targetDir); err != nil {
			logrus.WithField("target_dir", targetDir).Warnf("failed to remove temporary directory: %v", err)
		}

		targetDir = newTargetDir
	}

	return 0, targetDir, nil
}

func processHarborSourceV2(req *dto.CreateContainerRequest, userID int) (int, error) {
	harbor := req.BuildSource.Harbor

	// Extract image name from full image path
	imageName := harbor.Image
	if strings.Contains(imageName, "/") {
		parts := strings.Split(imageName, "/")
		imageName = parts[len(parts)-1]
	}

	harborClient := client.GetHarborClient()
	exists, err := harborClient.CheckImageExists(imageName, harbor.Tag)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to check harbor image: %v", err)
	}

	if !exists {
		return http.StatusNotFound, fmt.Errorf("image %s:%s not found in harbor", imageName, harbor.Tag)
	}

	return http.StatusOK, nil
}

func validateRequiredFilesV2(sourcePath string, buildOptions *dto.BuildOptions) error {
	contextDir := "."
	dockerfilePath := "Dockerfile"

	if buildOptions != nil {
		if buildOptions.ContextDir != "" {
			contextDir = buildOptions.ContextDir
		}
		if buildOptions.DockerfilePath != "" {
			dockerfilePath = buildOptions.DockerfilePath
		}
	}

	buildContextPath := filepath.Join(sourcePath, contextDir)
	if _, err := os.Stat(buildContextPath); os.IsNotExist(err) {
		return fmt.Errorf("build context path '%s' does not exist", contextDir)
	}

	buildDockerfilePath := filepath.Join(sourcePath, dockerfilePath)
	if _, err := os.Stat(buildDockerfilePath); os.IsNotExist(err) {
		return fmt.Errorf("Dockerfile '%s' does not exist", dockerfilePath)
	}

	return nil
}

func processHarborDirectUpdateV2(ctx context.Context, req *dto.CreateContainerRequest, userID int, isUpdate bool) (string, string, error) {
	taskID := fmt.Sprintf("harbor-%d", time.Now().UnixNano())
	traceID := fmt.Sprintf("trace-%d", time.Now().UnixNano())

	// Convert environment variables array to string
	envVarsStr := ""
	if len(req.EnvVars) > 0 {
		envVarsStr = strings.Join(req.EnvVars, ",")
	}

	container := &database.Container{
		Type:     string(req.ContainerType),
		Name:     req.Name,
		Image:    req.BuildSource.Harbor.Image,
		Tag:      req.BuildSource.Harbor.Tag,
		Command:  req.Command,
		EnvVars:  envVarsStr,
		UserID:   userID,
		IsPublic: req.IsPublic,
		Status:   true,
	}

	var err error
	if isUpdate {
		// Update existing container
		var existingContainer database.Container
		if err := database.DB.Where("name = ? AND type = ? AND image = ? AND tag = ? AND user_id = ? AND status = true",
			req.Name, req.ContainerType, req.BuildSource.Harbor.Image, req.BuildSource.Harbor.Tag, userID).First(&existingContainer).Error; err != nil {
			return "", "", fmt.Errorf("failed to find existing container record: %v", err)
		}

		// Update the existing record
		container.ID = existingContainer.ID
		container.CreatedAt = existingContainer.CreatedAt
		container.UpdatedAt = time.Now()

		err = database.DB.Save(container).Error
	} else {
		// Create new container
		err = database.DB.Create(container).Error
	}

	if err != nil {
		return "", "", fmt.Errorf("failed to save container record: %v", err)
	}

	return taskID, traceID, nil
}
