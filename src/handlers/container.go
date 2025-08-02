package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
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

const (
	Command     = "bash /entrypoint.sh"
	MaxFileSize = 5 * 1024 * 1024 // 5MB
)

// SubmitContainerBuilding
//
//	@Summary		Submit container build task
//	@Description	Build Docker images by uploading files, specifying GitHub repositories, or Harbor images. Supports zip and tar.gz file uploads, or automatically pulls code from GitHub for building, or directly updates the database from existing Harbor images. The system automatically validates required files (Dockerfile) and sets execution permissions.
//	@Tags			container
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			type			formData	string		false	"Container type, specifies the purpose of the container"	Enums(algorithm, benchmark)	default(algorithm)
//	@Param			name			formData	string		false	"Container name, used to identify the container, will be used as the image build identifier, defaults to the name field in info.toml"
//	@Param			image			formData	string		true	"Docker image name. When source_type is harbor, specify the existing image name in Harbor; otherwise, supports the following formats: 1) image-name (automatically adds default Harbor address and namespace) 2) namespace/image-name (automatically adds default Harbor address)"
//	@Param			tag				formData	string		false	"Docker image tag. When source_type is harbor, specify the existing image tag in Harbor; otherwise, used for version control"	default(latest)
//	@Param			command			formData	string		false	"Docker image startup command, defaults to bash /entrypoint.sh"	default(bash /entrypoint.sh)
//	@Param			env_vars		formData	[]string	false	"List of environment variable names, supports multiple variables"
//	@Param			source_type		formData	string		false	"Build source type, specifies the source of the code"	Enums(file,github,harbor)	default(file)
//	@Param			file			formData	file		false	"Source file (supports zip or tar.gz format), required when source_type is file, file size limit 5MB"
//	@Param			github_token	formData	string		false	"GitHub access token, used for private repositories, not required for public repositories"
//	@Param			github_repo		formData	string		false	"GitHub repository address, format: owner/repo, required when source_type is github"
//	@Param			github_branch	formData	string		false	"GitHub branch name, specifies the branch to build"	default(main)
//	@Param			github_commit	formData	string		false	"GitHub commit hash (supports short hash), if specified, branch parameter is ignored"
//	@Param			github_path		formData	string		false	"Subdirectory path in the repository, if the source code is not in the root directory"	default(.)
//	@Param			context_dir		formData	string		false	"Docker build context path, relative to the source root directory"	default(.)
//	@Param			dockerfile_path	formData	string		false	"Dockerfile path, relative to the source root directory"	default(Dockerfile)
//	@Param			target			formData	string		false	"Dockerfile build target (used for multi-stage builds)"
//	@Param			force_rebuild	formData	bool		false	"Whether to force rebuild the image, ignore cache"	default(false)
//	@Success		202				{object}	dto.GenericResponse[dto.SubmitResp]	"Successfully submitted container build task, returns task tracking information"
//	@Failure		400				{object}	dto.GenericResponse[any]	"Request parameter error: unsupported file format (only zip, tar.gz), file size exceeds limit (5MB), parameter validation failed, invalid GitHub repository address, invalid Harbor image parameter, invalid force_rebuild value, etc."
//	@Failure		404				{object}	dto.GenericResponse[any]	"Resource not found: build context path does not exist, missing required files (Dockerfile, entrypoint.sh), image not found in Harbor"
//	@Failure		500				{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v1/containers [post]
func SubmitContainerBuilding(c *gin.Context) {
	groupID := c.GetString("groupID")
	logrus.Infof("SubmitContainerBuilding, groupID: %s", groupID)

	req := &dto.SubmitContainerBuildingReq{
		ContainerType: consts.ContainerType(c.DefaultPostForm("type", string(consts.ContainerTypeAlgorithm))),
		Name:          c.PostForm("name"),
		Image:         c.PostForm("image"),
		Tag:           c.DefaultPostForm("tag", "latest"),
		Command:       c.DefaultPostForm("command", Command),
		EnvVars:       c.PostFormArray("env_vars"),
		Source: dto.BuildSource{
			Type: consts.BuildSourceType(c.DefaultPostForm("source_type", "file")),
		},
	}

	switch req.Source.Type {
	case consts.BuildSourceTypeFile:
		req.Source.File = &dto.FileSource{}
	case consts.BuildSourceTypeGitHub:
		req.Source.GitHub = &dto.GitHubSource{
			Token:      c.PostForm("github_token"),
			Repository: c.PostForm("github_repo"),
			Branch:     c.DefaultPostForm("github_branch", "main"),
			Commit:     c.PostForm("github_commit"),
			Path:       c.DefaultPostForm("github_path", "."),
		}
	case consts.BuildSourceTypeHarbor:
		req.Source.Harbor = &dto.HarborSource{
			Image: req.Image,
			Tag:   req.Tag,
		}
	}

	forceRebuildStr := c.DefaultPostForm("force_rebuild", "false")
	forceRebuild, err := strconv.ParseBool(forceRebuildStr)
	if err != nil {
		logrus.Errorf("invalid force_rebuild value: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid force_rebuild value")
		return
	}

	req.BuildOptions = &dto.BuildOptions{
		ContextDir:     c.DefaultPostForm("context_path", "."),
		DockerfilePath: c.DefaultPostForm("dockerfile_path", "Dockerfile"),
		Target:         c.DefaultPostForm("target", ""),
		ForceRebuild:   forceRebuild,
	}

	if err := req.Validate(); err != nil {
		logrus.Errorf("build request validation failed: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var code int
	var sourcePath string
	switch req.Source.Type {
	case consts.BuildSourceTypeFile:
		code, sourcePath, err = processFileSource(c, req)
		if err != nil {
			logrus.Errorf("failed to process file source: %v", err)
			dto.ErrorResponse(c, code, err.Error())
			return
		}
	case consts.BuildSourceTypeGitHub:
		code, sourcePath, err = processGitHubSource(req)
		if err != nil {
			logrus.Errorf("failed to process github source: %v", err)
			dto.ErrorResponse(c, code, err.Error())
			return
		}
	case consts.BuildSourceTypeHarbor:
		code, err = processHarborSource(req)
		if err != nil {
			logrus.Errorf("failed to process harbor source: %v", err)
			dto.ErrorResponse(c, code, err.Error())
			return
		}
		sourcePath = ""
	default:
		dto.ErrorResponse(c, http.StatusBadRequest, "Unsupported source type")
		return
	}

	if req.Source.Type != consts.BuildSourceTypeHarbor {
		if code, err := req.ValidateInfoContent(sourcePath); err != nil {
			logrus.Errorf("info content validation failed: %v", err)
			dto.ErrorResponse(c, code, err.Error())
			return
		}

		if err := validateRequiredFiles(sourcePath,
			req.BuildOptions.ContextDir,
			req.BuildOptions.DockerfilePath,
		); err != nil {
			logrus.Errorf("source validation failed: %v", err)
			dto.ErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}
	}

	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		logrus.Error("failed to get span context from gin.Context")
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get span context")
		return
	}
	spanCtx := ctx.(context.Context)

	if req.Source.Type == consts.BuildSourceTypeHarbor {
		taskID, traceID, err := processHarborDirectUpdate(spanCtx, req)
		if err != nil {
			logrus.Errorf("failed to process harbor direct update: %v", err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to process harbor direct update")
			return
		}

		dto.JSONResponse(c, http.StatusOK,
			"Container information updated successfully from Harbor",
			dto.SubmitResp{Traces: []dto.Trace{{TraceID: traceID, HeadTaskID: taskID, Index: 0}}},
		)
		return
	}

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
		GroupID:   groupID,
	}
	task.SetGroupCtx(spanCtx)

	taskID, traceID, err := executor.SubmitTask(spanCtx, task)
	if err != nil {
		logrus.Errorf("failed to submit container building task: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to submit container building task")
		return
	}

	dto.JSONResponse(c, http.StatusAccepted,
		"Container building task submitted successfully",
		dto.SubmitResp{Traces: []dto.Trace{{TraceID: traceID, HeadTaskID: taskID, Index: 0}}},
	)
}

func processFileSource(c *gin.Context, req *dto.SubmitContainerBuildingReq) (int, string, error) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		return http.StatusBadRequest, "", fmt.Errorf("failed to get uploaded file: %v", err)
	}
	defer file.Close()

	const maxSize = 5 * 1024 * 1024 // 5MB
	if header.Size > maxSize {
		return http.StatusBadRequest, "", fmt.Errorf("file size exceeds %dMB limit", maxSize/(1024*1024))
	}

	req.Source.File = &dto.FileSource{
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

	targetDir := path.Join(config.GetString("container.storage_path"), string(req.ContainerType), req.Name, fmt.Sprintf("build_%d", time.Now().Unix()))
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
		logrus.WithField("temp_dir", tempDir).Warnf("failed to remove temporary directory %s: %v", tempDir, err)
	}

	return 0, targetDir, nil
}

func processGitHubSource(req *dto.SubmitContainerBuildingReq) (int, string, error) {
	github := req.Source.GitHub

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
		checkoutCmd := exec.Command("git", "checkout", github.Commit)
		checkoutCmd.Dir = targetDir
		if output, err := checkoutCmd.CombinedOutput(); err != nil {
			return http.StatusInternalServerError, "", fmt.Errorf("failed to checkout commit %s: %v, output: %s", github.Commit, err, string(output))
		}
	}

	cmd := exec.Command(gitCmd[0], gitCmd[1:]...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("failed to clone repository: %v, output: %s", err, string(output))
	}

	if github.Path != "." {
		subPath := filepath.Join(targetDir, github.Path)
		if _, err := os.Stat(subPath); os.IsNotExist(err) {
			return http.StatusInternalServerError, "", fmt.Errorf("specified path %s does not exist in repository", github.Path)
		}

		newTargetDir := filepath.Join(config.GetString("container.storage_path"), string(req.ContainerType), req.Name, fmt.Sprintf("build_sub_%d", time.Now().Unix()))
		if err := os.MkdirAll(newTargetDir, 0755); err != nil {
			return http.StatusInternalServerError, "", fmt.Errorf("failed to create sub path target directory: %v", err)
		}

		if err := utils.CopyDir(subPath, newTargetDir); err != nil {
			return http.StatusInternalServerError, "", fmt.Errorf("failed to copy sub path: %v", err)
		}

		if err := os.RemoveAll(targetDir); err != nil {
			logrus.WithField("target_dir", targetDir).Warnf("failed to remove original target directory %s: %v", targetDir, err)
		}

		targetDir = newTargetDir
	}

	return 0, targetDir, nil
}

// validateRequiredFiles validates whether the required files for building exist
func validateRequiredFiles(sourcePath, ContextDir string, DockerfilePath string) error {
	buildContextPath := filepath.Join(sourcePath, ContextDir)
	if _, err := os.Stat(buildContextPath); os.IsNotExist(err) {
		return fmt.Errorf("build context path '%s' does not exist", ContextDir)
	}

	buildDockerfilePath := filepath.Join(sourcePath, DockerfilePath)
	if _, err := os.Stat(buildDockerfilePath); os.IsNotExist(err) {
		return fmt.Errorf("dockerfile '%s' does not exist", DockerfilePath)
	}

	return nil
}

func processHarborSource(req *dto.SubmitContainerBuildingReq) (int, error) {
	harbor := req.Source.Harbor

	// Parse image name, extract the last part from the full image path
	// Example: 10.10.10.240/library/rca-algo-random -> rca-algo-random
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

func processHarborDirectUpdate(ctx context.Context, req *dto.SubmitContainerBuildingReq) (string, string, error) {
	taskID := fmt.Sprintf("harbor-%d", time.Now().UnixNano())
	traceID := fmt.Sprintf("trace-%d", time.Now().UnixNano())

	container := &database.Container{
		Type:    string(req.ContainerType),
		Name:    req.Name,
		Image:   req.Source.Harbor.Image,
		Tag:     req.Source.Harbor.Tag,
		Command: req.Command,
		EnvVars: strings.Join(req.EnvVars, ","),
		Status:  true,
		UserID:  1, // TODO: Need to get actual user ID from authentication context
	}

	if err := repository.CreateContainer(container); err != nil {
		return "", "", fmt.Errorf("failed to create container record: %v", err)
	}

	logrus.Infof("Harbor container %s:%s updated successfully in database", req.Source.Harbor.Image, req.Source.Harbor.Tag)
	return taskID, traceID, nil
}
