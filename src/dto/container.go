package dto

import (
	"fmt"
	"path/filepath"
	"strings"

	"aegis/model"
	"aegis/utils"
)

const (
	InfoFileName  = "info.toml"
	InfoNameField = "name"
)

type ParameterItem struct {
	Key            string `json:"key"`
	Value          any    `json:"value,omitempty"`
	TemplateString string `json:"template_string,omitempty"`
}

type HelmConfigItem struct {
	Version       string          `json:"version"`
	RepoURL       string          `json:"repo_url"`
	RepoName      string          `json:"repo_name"`
	ChartName     string          `json:"chart_name"`
	LocalPath     string          `json:"local_path,omitempty"`
	ValueFile     string          `json:"value_file,omitempty"`
	DynamicValues []ParameterItem `json:"values,omitempty"`
}

func NewHelmConfigItem(cfg *model.HelmConfig) *HelmConfigItem {
	return &HelmConfigItem{
		Version:   cfg.Version,
		RepoURL:   cfg.RepoURL,
		RepoName:  cfg.RepoName,
		ChartName: cfg.ChartName,
		LocalPath: cfg.LocalPath,
		ValueFile: cfg.ValueFile,
	}
}

// GetValuesMap constructs a nested map of Helm values by merging file and dynamic values.
func (hci *HelmConfigItem) GetValuesMap() map[string]any {
	root := make(map[string]any)

	if hci.ValueFile != "" {
		if fileValues, err := utils.LoadYAMLFile(hci.ValueFile); err == nil {
			root = fileValues
		}
	}

	for _, item := range hci.DynamicValues {
		value := item.Value
		keys := utils.ParseHelmKey(item.Key)
		cur := root

		for i, k := range keys {
			if i == len(keys)-1 {
				if k.IsArray {
					if arr, ok := cur[k.Key].([]any); ok {
						for len(arr) <= k.Index {
							arr = append(arr, make(map[string]any))
						}
						arr[k.Index] = value
						cur[k.Key] = arr
					} else {
						arr := make([]any, k.Index+1)
						for j := 0; j < k.Index; j++ {
							arr[j] = make(map[string]any)
						}
						arr[k.Index] = value
						cur[k.Key] = arr
					}
				} else {
					cur[k.Key] = value
				}
				break
			}

			if k.IsArray {
				if _, exists := cur[k.Key]; !exists {
					arr := make([]any, k.Index+1)
					for j := 0; j <= k.Index; j++ {
						arr[j] = make(map[string]any)
					}
					cur[k.Key] = arr
				}

				if arr, ok := cur[k.Key].([]any); ok {
					for len(arr) <= k.Index {
						arr = append(arr, make(map[string]any))
					}
					cur[k.Key] = arr

					if nextMap, ok := arr[k.Index].(map[string]any); ok {
						cur = nextMap
					} else {
						newMap := make(map[string]any)
						arr[k.Index] = newMap
						cur = newMap
					}
				}
			} else {
				if _, exists := cur[k.Key]; !exists {
					cur[k.Key] = make(map[string]any)
				}
				if nextMap, ok := cur[k.Key].(map[string]any); ok {
					cur = nextMap
				} else {
					newMap := make(map[string]any)
					cur[k.Key] = newMap
					cur = newMap
				}
			}
		}
	}

	return root
}

type ContainerVersionItem struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ImageRef string `json:"image_ref"`
	Command  string `json:"command,omitempty"`

	ContainerID   int    `json:"container_id"`
	ContainerName string `json:"container_name"`

	EnvVars []ParameterItem `json:"env_vars,omitempty"`
	Extra   *HelmConfigItem `json:"extra,omitempty"`
}

func NewContainerVersionItem(version *model.ContainerVersion) ContainerVersionItem {
	item := ContainerVersionItem{
		ID:       version.ID,
		Name:     version.Name,
		ImageRef: version.ImageRef,
		Command:  version.Command,
	}

	if version.Container != nil {
		item.ContainerID = version.Container.ID
		item.ContainerName = version.Container.Name
	}

	return item
}

type ContainerRef struct {
	Name    string `json:"name" binding:"required"`
	Version string `json:"version" binding:"omitempty"`
}

func (ref *ContainerRef) Validate() error {
	if ref.Name == "" {
		return fmt.Errorf("algorithm name is required")
	}
	if ref.Version != "" {
		if _, _, _, err := utils.ParseSemanticVersion(ref.Version); err != nil {
			return fmt.Errorf("invalid semantic version: %s, %v", ref.Version, err)
		}
	}
	return nil
}

type ContainerSpec struct {
	ContainerRef
	EnvVars []ParameterSpec `json:"env_vars" binding:"omitempty"`
	Payload map[string]any  `json:"payload,omitempty" swaggertype:"object"`
}

func (item *ContainerSpec) Validate() error {
	if err := item.ContainerRef.Validate(); err != nil {
		return err
	}
	for _, envVar := range item.EnvVars {
		if err := utils.IsValidEnvVar(envVar.Key); err != nil {
			return fmt.Errorf("invalid environment variable key %s: %v", envVar.Key, err)
		}
	}
	return nil
}

type ParameterSpec struct {
	Key   string `json:"key"`
	Value any    `json:"value,omitempty"`
}

type BuildOptions struct {
	ContextDir     string            `json:"context_dir" binding:"omitempty" default:"."`
	DockerfilePath string            `json:"dockerfile_path" binding:"omitempty" default:"Dockerfile"`
	Target         string            `json:"target" binding:"omitempty"`
	BuildArgs      map[string]string `json:"build_args" binding:"omitempty" swaggertype:"object"`
	ForceRebuild   *bool             `json:"force_rebuild" binding:"omitempty"`
}

func (opts *BuildOptions) Validate() error {
	if opts.ContextDir != "" {
		opts.ContextDir = strings.TrimSpace(opts.ContextDir)
	}
	if opts.DockerfilePath != "" {
		opts.DockerfilePath = strings.TrimSpace(opts.DockerfilePath)
	}
	if opts.Target != "" {
		opts.Target = strings.TrimSpace(opts.Target)
	}
	if opts.BuildArgs != nil {
		for key := range opts.BuildArgs {
			if key == "" {
				return fmt.Errorf("build arg key cannot be empty")
			}
		}
	}

	if opts.ContextDir == "" {
		opts.ContextDir = "."
	}
	if opts.DockerfilePath == "" {
		opts.DockerfilePath = "Dockerfile"
	}
	if opts.ForceRebuild == nil {
		opts.ForceRebuild = utils.BoolPtr(false)
	}

	return nil
}

func (opts *BuildOptions) ValidateRequiredFiles(sourcePath string) error {
	contextPath := filepath.Join(sourcePath, opts.ContextDir)
	if utils.CheckFileExists(contextPath) {
		return fmt.Errorf("build context path '%s' does not exist", contextPath)
	}

	dockerfilePath := filepath.Join(sourcePath, opts.DockerfilePath)
	if !utils.CheckFileExists(dockerfilePath) {
		return fmt.Errorf("dockerfile not found at path: %s", dockerfilePath)
	}

	return nil
}
