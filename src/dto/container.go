package dto

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/utils"
)

const (
	InfoFileName  = "info.toml"
	InfoNameField = "name"
)

type GetContainerFilterOptions struct {
	Type  consts.ContainerType
	Name  string
	Image string
	Tag   string
}

type ListContainersFilterOptions struct {
	Status *bool
	Type   consts.ContainerType
	Names  []string
}

// BuildSource represents the source configuration for container building
// @Description Build source configuration with different source types
type BuildSource struct {
	// @Description Build source type (file, github, or harbor)
	Type consts.BuildSourceType `json:"type" binding:"required" swaggertype:"string"`
	// @Description File source configuration (for file uploads)
	File *FileSource `json:"file" binding:"omitempty"`
	// @Description GitHub source configuration
	GitHub *GitHubSource `json:"github" binding:"omitempty"`
	// @Description Harbor source configuration
	Harbor *HarborSource `json:"harbor" binding:"omitempty"`
}

func (s *BuildSource) Validate() error {
	if s.Type == consts.BuildSourceTypeGitHub {
		return s.GitHub.Validate()
	}

	if s.Type == consts.BuildSourceTypeHarbor {
		return s.Harbor.Validate()
	}

	return nil
}

// FileSource File source configuration
// @Description File source configuration for uploads
type FileSource struct {
	// Files uploaded via multipart/form-data are automatically processed
	// This is just for documentation purposes
	// @Description Filename of the uploaded file
	Filename string `json:"file_name,omitempty"`
	// @Description Size of the uploaded file in bytes
	Size int64 `json:"size,omitempty"`
}

// GitHubSource GitHub source configuration
// @Description GitHub source configuration
type GitHubSource struct {
	// @Description GitHub repository in format 'owner/repo'
	Repository string `json:"repository" binding:"required"`
	// @Description GitHub access token (optional)
	Token string `json:"token" binding:"omitempty"`
	// @Description Branch name (optional, defaults to main)
	Branch string `json:"branch" binding:"omitempty"`
	// @Description Specific commit hash (optional)
	Commit string `json:"commit" binding:"omitempty"`
	// @Description Path within the repository (optional)
	Path string `json:"path" binding:"omitempty"`
}

func (s *GitHubSource) Validate() error {
	parts := strings.Split(s.Repository, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}

	if s.Token != "" {
		if err := utils.IsValidGitHubToken(s.Token); err != nil {
			return err
		}
	}

	if s.Branch != "" {
		if err := utils.IsValidGitHubBranch(s.Branch); err != nil {
			return err
		}
	}

	if s.Commit != "" {
		if err := utils.IsValidGitHubCommit(s.Commit); err != nil {
			return err
		}
	}

	// Validate path format
	if err := utils.IsValidGitHubPath(s.Path); err != nil {
		return err
	}

	return nil
}

// HarborSource Harbor source configuration
// @Description Harbor source configuration
type HarborSource struct {
	// @Description Harbor image name
	Image string `json:"image" binding:"required"`
	// @Description Image tag (optional, defaults to latest)
	Tag string `json:"tag" binding:"omitempty"`
}

func (s *HarborSource) Validate() error {
	if s.Image == "" {
		return fmt.Errorf("image name cannot be empty")
	}

	if s.Tag != "" {
		if err := utils.IsValidDockerTag(s.Tag); err != nil {
			return fmt.Errorf("invalid Docker tag: %s, %v", s.Tag, err)
		}
	}

	return nil
}

// BuildOptions Build options
// @Description Build options for container creation
type BuildOptions struct {
	// @Description Context directory for build (optional)
	ContextDir string `json:"context_dir" binding:"omitempty"`
	// @Description Path to Dockerfile (optional, defaults to Dockerfile)
	DockerfilePath string `json:"dockerfile_path" binding:"omitempty"`
	// @Description Build target (optional)
	Target string `json:"target" binding:"omitempty"`
	// @Description Build arguments (optional)
	BuildArgs map[string]string `json:"build_args" binding:"omitempty" swaggertype:"object"`
	// @Description Build labels (optional)
	Labels map[string]string `json:"labels" binding:"omitempty" swaggertype:"object"`
	// @Description Force rebuild even if image exists
	ForceRebuild bool `json:"force_rebuild" binding:"omitempty"`
}

func (opts *BuildOptions) Validate() error {
	if opts.BuildArgs != nil {
		for key := range opts.BuildArgs {
			if key == "" {
				return fmt.Errorf("build arg key cannot be empty")
			}
		}
	}

	if opts.Labels != nil {
		for key := range opts.Labels {
			if key == "" {
				return fmt.Errorf("label key cannot be empty")
			}
		}
	}

	return nil
}

type SubmitContainerBuildingReq struct {
	ContainerType consts.ContainerType `json:"type" binding:"required,oneof=algorithm benchmark"`
	Name          string               `json:"name" binding:"required"`
	Image         string               `json:"image" binding:"required"`
	Tag           string               `json:"tag" binding:"omitempty"`
	Command       string               `json:"command" binding:"omitempty"`
	EnvVars       []string             `json:"env_vars" binding:"omitempty" swaggertype:"array"`
	Source        BuildSource          `json:"source" binding:"required"`
	BuildOptions  *BuildOptions        `json:"build_options" binding:"omitempty"`
}

func (req *SubmitContainerBuildingReq) Validate() error {
	if req.ContainerType != "" {
		if _, exists := ValidContainerTypes[req.ContainerType]; !exists {
			return fmt.Errorf("invalid container type: %s", req.ContainerType)
		}
	}

	if req.Image == "" {
		return fmt.Errorf("docker image cannot be empty")
	} else {
		if len(req.Image) > 255 {
			return fmt.Errorf("image name %s too long (max 255 characters)", req.Image)
		}

		if !strings.Contains(req.Image, "/") {
			req.Image = fmt.Sprintf("%s/%s/%s", config.GetString("harbor.host"), config.GetString("harbor.project"), req.Image)
		} else if !strings.Contains(req.Image, ".") {
			req.Image = fmt.Sprintf("%s/%s", config.GetString("harbor.host"), req.Image)
		}
	}

	if req.Tag != "" {
		if err := utils.IsValidDockerTag(req.Tag); err != nil {
			return fmt.Errorf("invalid docker tag: %s, %v", req.Tag, err)
		}
	}

	if req.EnvVars != nil {
		for i, envVar := range req.EnvVars {
			if err := utils.IsValidEnvVar(envVar); err != nil {
				return fmt.Errorf("invalid environment variable %s at index %d: %v", envVar, i, err)
			}
		}
	}

	if err := req.Source.Validate(); err != nil {
		return fmt.Errorf("invalid build source: %v", err)
	}

	if err := req.BuildOptions.Validate(); err != nil {
		return fmt.Errorf("invalid build options: %v", err)
	}

	return nil
}

func (req *SubmitContainerBuildingReq) ValidateInfoContent(sourcePath string) (int, error) {
	if req.Name == "" {
		content, err := getInfoFileContent(sourcePath)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		if name, ok := content[InfoNameField].(string); ok && name != "" {
			req.Name = name
		} else {
			return http.StatusNotFound, fmt.Errorf("%s does not contain a valid name field", InfoFileName)
		}
	}

	if req.Name == "" {
		return http.StatusBadRequest, fmt.Errorf("container name cannot be empty")
	}

	if req.Name == config.GetString("algo.detector") {
		return http.StatusBadRequest, fmt.Errorf("name '%s' is reserved and cannot be used for building images", config.GetString("algo.detector"))
	}

	return http.StatusOK, nil
}

func getInfoFileContent(sourcePath string) (map[string]any, error) {
	tomlPath := filepath.Join(sourcePath, InfoFileName)

	data, err := os.ReadFile(tomlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file: %v", InfoFileName, err)
	}

	var config map[string]any
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse %s file: %v", InfoFileName, err)
	}

	return config, nil
}

var ValidContainerTypes = map[consts.ContainerType]struct{}{
	consts.ContainerTypeAlgorithm: {},
	consts.ContainerTypeBenchmark: {},
}

// CreateContainerRequest represents the v2 API request for creating a container
// Containers are associated with users, not projects
// @Description Container creation request for v2 API
type CreateContainerRequest struct {
	// @Description Container type (algorithm or benchmark)
	// @example algorithm
	ContainerType consts.ContainerType `json:"type" binding:"required,oneof=algorithm benchmark" swaggertype:"string"`
	// @Description Container name
	// @example my-container
	Name string `json:"name" binding:"required"`
	// @Description Docker image name
	// @example my-image
	Image string `json:"image" binding:"required"`
	// @Description Docker image tag
	// @example latest
	Tag string `json:"tag" binding:"omitempty"`
	// @Description Container startup command
	// @example /bin/bash
	Command string `json:"command" binding:"omitempty"`
	// @Description Environment variables
	EnvVars []string `json:"env_vars" binding:"omitempty"`
	// @Description Whether the container is public
	IsPublic bool `json:"is_public" binding:"omitempty"`
	// @Description Container build source configuration
	BuildSource *BuildSource `json:"build_source" binding:"omitempty"`
	// @Description Container build options
	BuildOptions *BuildOptions `json:"build_options" binding:"omitempty"`
}

func (req *CreateContainerRequest) Validate() error {
	if req.ContainerType != "" {
		if _, exists := ValidContainerTypes[req.ContainerType]; !exists {
			return fmt.Errorf("invalid container type: %s", req.ContainerType)
		}
	}

	if req.Name == "" {
		return fmt.Errorf("container name cannot be empty")
	}

	if req.Image == "" {
		return fmt.Errorf("docker image cannot be empty")
	} else {
		if len(req.Image) > 255 {
			return fmt.Errorf("image name %s too long (max 255 characters)", req.Image)
		}

		// Auto-complete image name with Harbor host and project if needed
		if !strings.Contains(req.Image, "/") {
			req.Image = fmt.Sprintf("%s/%s/%s", config.GetString("harbor.host"), config.GetString("harbor.project"), req.Image)
		} else if !strings.Contains(req.Image, ".") {
			req.Image = fmt.Sprintf("%s/%s", config.GetString("harbor.host"), req.Image)
		}
	}

	if req.Tag != "" {
		if err := utils.IsValidDockerTag(req.Tag); err != nil {
			return fmt.Errorf("invalid docker tag: %s, %v", req.Tag, err)
		}
	}

	if req.BuildSource != nil {
		if err := req.BuildSource.Validate(); err != nil {
			return fmt.Errorf("invalid build source: %v", err)
		}
	}

	if req.BuildOptions != nil {
		if err := req.BuildOptions.Validate(); err != nil {
			return fmt.Errorf("invalid build options: %v", err)
		}
	}

	return nil
}

// ValidateInfoContent validates the info.toml content in the source path
func (req *CreateContainerRequest) ValidateInfoContent(sourcePath string) (int, error) {
	infoFilePath := filepath.Join(sourcePath, InfoFileName)

	if _, err := os.Stat(infoFilePath); os.IsNotExist(err) {
		return http.StatusNotFound, fmt.Errorf("required %s file not found in source", InfoFileName)
	}

	config, err := getInfoFileContent(sourcePath)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// If name is not provided in request, get it from info.toml
	if req.Name == "" {
		if nameValue, exists := config[InfoNameField]; exists {
			if nameStr, ok := nameValue.(string); ok && nameStr != "" {
				req.Name = nameStr
			} else {
				return http.StatusBadRequest, fmt.Errorf("invalid or empty 'name' field in %s", InfoFileName)
			}
		} else {
			return http.StatusBadRequest, fmt.Errorf("missing 'name' field in %s and not provided in request", InfoFileName)
		}
	}

	return http.StatusOK, nil
}

// UpdateContainerRequest represents the request for updating a container
// @Description Request structure for updating container information
type UpdateContainerRequest struct {
	// @Description Container name (optional)
	Name *string `json:"name" binding:"omitempty"`
	// @Description Container type (optional)
	Type *consts.ContainerType `json:"type" binding:"omitempty,oneof=algorithm benchmark" swaggertype:"string"`
	// @Description Docker image name (optional)
	Image *string `json:"image" binding:"omitempty"`
	// @Description Docker image tag (optional)
	Tag *string `json:"tag" binding:"omitempty"`
	// @Description Container startup command (optional)
	Command *string `json:"command" binding:"omitempty"`
	// @Description Environment variables (optional)
	EnvVars *string `json:"env_vars" binding:"omitempty"`
	// @Description Whether the container is public (optional)
	IsPublic *bool `json:"is_public" binding:"omitempty"`
	// @Description Container status (optional)
	Status *bool `json:"status" binding:"omitempty"`
}

func (req *UpdateContainerRequest) Validate() error {
	if req.Name != nil && *req.Name == "" {
		return fmt.Errorf("container name cannot be empty")
	}

	if req.Type != nil {
		if _, exists := ValidContainerTypes[*req.Type]; !exists {
			return fmt.Errorf("invalid container type: %s", *req.Type)
		}
	}

	if req.Image != nil {
		if *req.Image == "" {
			return fmt.Errorf("docker image cannot be empty")
		}
		if len(*req.Image) > 255 {
			return fmt.Errorf("image name %s too long (max 255 characters)", *req.Image)
		}
	}

	if req.Tag != nil && *req.Tag != "" {
		if err := utils.IsValidDockerTag(*req.Tag); err != nil {
			return fmt.Errorf("invalid docker tag: %s, %v", *req.Tag, err)
		}
	}

	return nil
}
