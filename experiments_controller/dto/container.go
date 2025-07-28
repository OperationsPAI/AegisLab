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

type BuildSource struct {
	Type   consts.BuildSourceType `json:"type" binding:"required"`
	File   *FileSource            `json:"file" binding:"omitempty"`
	GitHub *GitHubSource          `json:"github" binding:"omitempty"`
	Harbor *HarborSource          `json:"harbor" binding:"omitempty"`
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

// FileSource 文件源配置
type FileSource struct {
	// 通过 multipart/form-data 上传的文件会自动处理
	// 这里只是为了文档说明
	Filename string `json:"file_name,omitempty"`
	Size     int64  `json:"size,omitempty"`
}

// GitHubSource GitHub源配置
type GitHubSource struct {
	Repository string `json:"repository" binding:"required"`
	Token      string `json:"token" binding:"omitempty"`
	Branch     string `json:"branch" binding:"omitempty"`
	Commit     string `json:"commit" binding:"omitempty"`
	Path       string `json:"path" binding:"omitempty"`
} //@name GitHubSource

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

	// 验证 path 格式
	if err := utils.IsValidGitHubPath(s.Path); err != nil {
		return err
	}

	return nil
}

// HarborSource Harbor源配置
type HarborSource struct {
	Image string `json:"image" binding:"required"`
	Tag   string `json:"tag" binding:"omitempty"`
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

// BuildOptions 构建选项
type BuildOptions struct {
	ContextDir     string            `json:"context_dir" binding:"omitempty"`
	DockerfilePath string            `json:"dockerfile_path" binding:"omitempty"`
	Target         string            `json:"target" binding:"omitempty"`
	BuildArgs      map[string]string `json:"build_args" binding:"omitempty" swaggertype:"object"`
	Labels         map[string]string `json:"labels" binding:"omitempty" swaggertype:"object"`
	ForceRebuild   bool              `json:"force_rebuild" binding:"omitempty"`
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
	ProjectName   string               `json:"project_name" binding:"required"`
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
	if req.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}

	if req.ContainerType != "" {
		if _, exists := ValidContainerTypes[req.ContainerType]; !exists {
			return fmt.Errorf("Invalid container type: %s", req.ContainerType)
		}
	}

	if req.Image == "" {
		return fmt.Errorf("Docker image cannot be empty")
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
			return fmt.Errorf("Invalid Docker tag: %s, %v", req.Tag, err)
		}
	}

	if req.EnvVars != nil {
		for i, envVar := range req.EnvVars {
			if err := utils.IsValidEnvVar(envVar); err != nil {
				return fmt.Errorf("Invalid environment variable %s at index %d: %v", envVar, i, err)
			}
		}
	}

	if err := req.Source.Validate(); err != nil {
		return fmt.Errorf("Invalid build source: %v", err)
	}

	if err := req.BuildOptions.Validate(); err != nil {
		return fmt.Errorf("Invalid build options: %v", err)
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
		return http.StatusBadRequest, fmt.Errorf("Container name cannot be empty")
	}

	if req.Name == config.GetString("algo.detector") {
		return http.StatusBadRequest, fmt.Errorf("Name '%s' is reserved and cannot be used for building images", config.GetString("algo.detector"))
	}

	return http.StatusOK, nil
}

func getInfoFileContent(sourcePath string) (map[string]any, error) {
	tomlPath := filepath.Join(sourcePath, InfoFileName)

	data, err := os.ReadFile(tomlPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s file: %v", InfoFileName, err)
	}

	var config map[string]any
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("Failed to parse %s file: %v", InfoFileName, err)
	}

	return config, nil
}

var ValidContainerTypes = map[consts.ContainerType]struct{}{
	consts.ContainerTypeAlgorithm: {},
	consts.ContainerTypeBenchmark: {},
}
