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

	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/executor"
	"github.com/LGU-SE-Internal/rcabench/middleware"
	"github.com/LGU-SE-Internal/rcabench/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	MaxFileSize = 5 * 1024 * 1024 // 5MB
)

// SubmitContainerBuilding
//
//	@Summary		提交镜像构建任务
//	@Description	通过上传文件或指定GitHub仓库来构建Docker镜像。支持zip和tar.gz格式的文件上传，或从GitHub仓库自动拉取代码进行构建。系统会自动验证必需文件（Dockerfile）并设置执行权限
//	@Tags			container
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			type			formData	string		false	"容器类型，指定容器的用途"	Enums(algorithm, benchmark)	default(algorithm)
//	@Param			name			formData	string		false	"容器名称，用于标识容器，将作为镜像构建的标识符，默认使用info.toml中的name字段"
//	@Param			image			formData	string		true	"Docker镜像名称。支持以下格式：1) image-name（自动添加默认Harbor地址和命名空间）2) namespace/image-name（自动添加默认Harbor地址）"
//	@Param			tag				formData	string		false	"Docker镜像标签，用于版本控制"	default(latest)
//	@Param			command			formData	string		false	"Docker镜像启动命令，默认为bash /entrypoint.sh"	default(bash /entrypoint.sh)
//	@Param			env_vars		formData	[]string	false	"环境变量名称列表，支持多个环境变量"
//	@Param			source_type		formData	string		false	"构建源类型，指定源码来源"	Enums(file,github)	default(file)
//	@Param			file			formData	file		false	"源码文件（支持zip或tar.gz格式），当source_type为file时必需，文件大小限制5MB"
//	@Param			github_token	formData	string		false	"GitHub访问令牌，用于访问私有仓库，公开仓库可不提供"
//	@Param			github_repo		formData	string		false	"GitHub仓库地址，格式：owner/repo，当source_type为github时必需"
//	@Param			github_branch	formData	string		false	"GitHub分支名，指定要构建的分支"	default(main)
//	@Param			github_commit	formData	string		false	"GitHub commit哈希值（支持短hash），如果指定commit则忽略branch参数"
//	@Param			github_path		formData	string		false	"仓库内的子目录路径，如果源码不在根目录"	default(.)
//	@Param			context_dir		formData	string		false	"Docker构建上下文路径，相对于源码根目录"	default(.)
//	@Param			dockerfile_path	formData	string		false	"Dockerfile路径，相对于源码根目录"	default(Dockerfile)
//	@Param			target			formData	string		false	"Dockerfile构建目标（multi-stage build时使用）"
//	@Param			force_rebuild	formData	bool		false	"是否强制重新构建镜像，忽略缓存"	default(false)
//	@Success		202				{object}	dto.GenericResponse[dto.SubmitResp]	"成功提交容器构建任务，返回任务跟踪信息"
//	@Failure		400				{object}	dto.GenericResponse[any]	"请求参数错误：文件格式不支持（仅支持zip、tar.gz）、文件大小超限（5MB）、参数验证失败、GitHub仓库地址无效、force_rebuild值格式错误等"
//	@Failure		404				{object}	dto.GenericResponse[any]	"资源不存在：构建上下文路径不存在、缺少必需文件（Dockerfile、entrypoint.sh）"
//	@Failure		500				{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/containers [post]
func SubmitContainerBuilding(c *gin.Context) {
	groupID := c.GetString("groupID")
	logrus.Infof("SubmitContainerBuilding, groupID: %s", groupID)

	req := &dto.SubmitContainerBuildingReq{
		ContainerType: consts.ContainerType(c.DefaultPostForm("type", string(consts.ContainerTypeAlgorithm))),
		Name:          c.PostForm("name"),
		Image:         c.PostForm("image"),
		Tag:           c.DefaultPostForm("tag", "latest"),
		Command:       c.DefaultPostForm("command", "bash /entrypoint.sh"),
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
	default:
		dto.ErrorResponse(c, http.StatusBadRequest, "Unsupported source type")
		return
	}

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

	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		logrus.Error("failed to get span context from gin.Context")
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get span context")
		return
	}
	spanCtx := ctx.(context.Context)

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

	const maxSize = 5 * 1024 * 1024 // 50MB
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

// validateRequiredFiles 验证构建所需的文件是否存在
func validateRequiredFiles(sourcePath, ContextDir string, DockerfilePath string) error {
	buildContextPath := filepath.Join(sourcePath, ContextDir)
	if _, err := os.Stat(buildContextPath); os.IsNotExist(err) {
		return fmt.Errorf("build context path '%s' does not exist", ContextDir)
	}

	buildDockerfilePath := filepath.Join(sourcePath, DockerfilePath)
	if _, err := os.Stat(buildDockerfilePath); os.IsNotExist(err) {
		return fmt.Errorf("Dockerfile '%s' does not exist", DockerfilePath)
	}

	return nil
}
