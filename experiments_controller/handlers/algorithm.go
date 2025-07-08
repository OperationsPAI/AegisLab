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
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/LGU-SE-Internal/rcabench/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ListAlgorithms
//
//	@Summary		获取算法列表
//	@Description	获取系统中所有可用的算法列表，包括算法的镜像信息、标签和更新时间。只返回状态为激活的算法容器
//	@Tags			algorithm
//	@Produce		json
//	@Success		200	{object}	dto.GenericResponse[dto.ListAlgorithmsResp]	"成功返回算法列表"
//	@Failure		500	{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/algorithms [get]
func ListAlgorithms(c *gin.Context) {
	containers, err := repository.ListContainers(&dto.ListContainersFilterOptions{
		Type:   consts.ContainerTypeAlgorithm,
		Status: utils.BoolPtr(true),
	})
	if err != nil {
		logrus.Errorf("failed to list algorithms: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to list algorithms")
		return
	}

	dto.SuccessResponse(c, dto.ListAlgorithmsResp(containers))
}

// SubmitAlgorithmExecution
//
//	@Summary		提交算法执行任务
//	@Description	批量提交算法执行任务，支持多个算法和数据集的组合执行。系统将为每个执行任务分配唯一的 TraceID 用于跟踪任务状态和结果
//	@Tags			algorithm
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		dto.SubmitExecutionReq	true	"算法执行请求列表，包含算法名称、数据集和环境变量"
//	@Success		202		{object}	dto.GenericResponse[dto.SubmitResp]	"成功提交算法执行任务，返回任务跟踪信息"
//	@Failure		400		{object}	dto.GenericResponse[any]	"请求参数错误，如JSON格式不正确、算法名称或数据集名称无效、环境变量名称不支持等"
//	@Failure		500		{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/algorithms [post]
func SubmitAlgorithmExecution(c *gin.Context) {
	groupID := c.GetString("groupID")
	logrus.Infof("SubmitAlgorithmExecution called, groupID: %s", groupID)

	var req dto.SubmitExecutionReq
	if err := c.BindJSON(&req); err != nil {
		logrus.Error(err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		logrus.Error("failed to get span context from gin.Context")
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get span context")
		return
	}
	spanCtx := ctx.(context.Context)

	traces := make([]dto.Trace, 0, len(req))
	for idx, payload := range req {
		task := &dto.UnifiedTask{
			Type:      consts.TaskTypeRunAlgorithm,
			Payload:   utils.StructToMap(payload),
			Immediate: true,
			GroupID:   groupID,
		}
		task.SetGroupCtx(spanCtx)

		taskID, traceID, err := executor.SubmitTask(spanCtx, task)
		if err != nil {
			logrus.Errorf("failed to submit algorithm execution task: %v", err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to submit algorithm execution task")
			return
		}

		traces = append(traces, dto.Trace{TraceID: traceID, HeadTaskID: taskID, Index: idx})
	}

	dto.JSONResponse(c, http.StatusAccepted,
		"Algorithm executions submitted successfully",
		dto.SubmitResp{GroupID: groupID, Traces: traces},
	)
}

// SubmitAlgorithmBuilding
//
//	@Summary		提交算法构建任务
//	@Description	通过上传文件或指定GitHub仓库来构建算法Docker镜像。支持zip和tar.gz格式的文件上传，或从GitHub仓库自动拉取代码进行构建。系统会自动验证必需文件（Dockerfile）并设置执行权限
//	@Tags			algorithm
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			algorithm		formData	string	true	"算法名称，用于标识算法，将作为镜像构建的标识符"
//	@Param			image			formData	string	true	"Docker镜像名称。支持以下格式：1) image-name（自动添加默认Harbor地址和命名空间）2) namespace/image-name（自动添加默认Harbor地址）"
//	@Param			tag				formData	string	false	"Docker镜像标签，用于版本控制"	default(latest)
//	@Param			command			formData	string	false	"Docker镜像启动命令，默认为bash /entrypoint.sh"	default(bash /entrypoint.sh)
//	@Param			source_type		formData	string	false	"构建源类型，指定算法源码来源"	Enums(file,github)	default(file)
//	@Param			file			formData	file	false	"算法源码文件（支持zip或tar.gz格式），当source_type为file时必需，文件大小限制5MB"
//	@Param			github_token	formData	string	false	"GitHub访问令牌，用于访问私有仓库，公开仓库可不提供"
//	@Param			github_repo		formData	string	false	"GitHub仓库地址，格式：owner/repo，当source_type为github时必需"
//	@Param			github_branch	formData	string	false	"GitHub分支名，指定要构建的分支"	default(main)
//	@Param			github_commit	formData	string	false	"GitHub commit哈希值（支持短hash），如果指定commit则忽略branch参数"
//	@Param			github_path		formData	string	false	"仓库内的子目录路径，如果算法源码不在根目录"	default(.)
//	@Param			context_dir		formData	string	false	"Docker构建上下文路径，相对于源码根目录"	default(.)
//	@Param			dockerfile_path	formData	string	false	"Dockerfile路径，相对于源码根目录"	default(Dockerfile)
//	@Param			target			formData	string	false	"Dockerfile构建目标（multi-stage build时使用）"
//	@Param			force_rebuild	formData	bool	false	"是否强制重新构建镜像，忽略缓存"	default(false)
//	@Success		202				{object}	dto.GenericResponse[dto.SubmitResp]	"成功提交算法构建任务，返回任务跟踪信息"
//	@Failure		400				{object}	dto.GenericResponse[any]	"请求参数错误：文件格式不支持（仅支持zip、tar.gz）、文件大小超限（5MB）、参数验证失败、GitHub仓库地址无效、force_rebuild值格式错误等"
//	@Failure		404				{object}	dto.GenericResponse[any]	"资源不存在：构建上下文路径不存在、缺少必需文件（Dockerfile、entrypoint.sh）"
//	@Failure		500				{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/algorithms/build [post]
func SubmitAlgorithmBuilding(c *gin.Context) {
	groupID := c.GetString("groupID")
	logrus.Infof("SubmitAlgorithmBuilding, groupID: %s", groupID)

	req := &dto.SubmitImageBuildingReq{
		Algorithm: c.PostForm("algorithm"),
		Image:     c.PostForm("image"),
		Tag:       c.DefaultPostForm("tag", "latest"),
		Command:   c.DefaultPostForm("command", "bash /entrypoint.sh"),
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
			consts.BuildContainerType: consts.ContainerTypeAlgorithm,
			consts.BuildName:          req.Algorithm,
			consts.BuildImage:         req.Image,
			consts.BuildTag:           req.Tag,
			consts.BuildCommand:       req.Command,
			consts.BuildSourcePath:    sourcePath,
			consts.BuildBuildOptions:  req.BuildOptions,
		},
		Immediate: true,
		GroupID:   groupID,
	}
	task.SetGroupCtx(spanCtx)

	taskID, traceID, err := executor.SubmitTask(spanCtx, task)
	if err != nil {
		logrus.Errorf("failed to submit algorithm image building task: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to submit algorithm image building task")
		return
	}

	dto.JSONResponse(c, http.StatusAccepted,
		"Algorithm image building task submitted successfully",
		dto.SubmitResp{Traces: []dto.Trace{{TraceID: traceID, HeadTaskID: taskID, Index: 0}}},
	)
}

func processFileSource(c *gin.Context, req *dto.SubmitImageBuildingReq) (int, string, error) {
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

	tempDir, err := os.MkdirTemp("", "algorithm-upload-*")
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

	targetDir := path.Join(config.GetString("algo.storage_path"), req.Algorithm, fmt.Sprintf("build_%d", time.Now().Unix()))
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

func processGitHubSource(req *dto.SubmitImageBuildingReq) (int, string, error) {
	github := req.Source.GitHub

	targetDir := filepath.Join(config.GetString("algo.storage_path"), req.Algorithm, fmt.Sprintf("build_%d", time.Now().Unix()))
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

		newTargetDir := filepath.Join(config.GetString("algo.storage_path"), req.Algorithm, fmt.Sprintf("build_sub_%d", time.Now().Unix()))
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
