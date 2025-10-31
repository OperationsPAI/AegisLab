package dto

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"aegis/consts"
	"aegis/database"
	"aegis/utils"
)

const (
	InfoFileName  = "info.toml"
	InfoNameField = "name"
)

// BuildContainerRequest represents the request for building a container into platform registry
type BuildContainerRequest struct {
	// Container Meta
	ImageName string `json:"image_name" binding:"required"`
	Tag       string `json:"tag" binding:"omitempty"`

	// GitHub repository information
	GithubRepository string `json:"github_repository" binding:"required"`
	GithubBranch     string `json:"github_branch" binding:"omitempty"`
	GithubCommit     string `json:"github_commit" binding:"omitempty"`
	GithubToken      string `json:"github_token" binding:"omitempty"`
	SubPath          string `json:"sub_path" binding:"omitempty"`

	Options *BuildOptions `json:"build_options" binding:"omitempty"`
}

type BuildOptions struct {
	ContextDir     string            `json:"context_dir" binding:"omitempty" default:"."`
	DockerfilePath string            `json:"dockerfile_path" binding:"omitempty" default:"Dockerfile"`
	Target         string            `json:"target" binding:"omitempty"`
	BuildArgs      map[string]string `json:"build_args" binding:"omitempty" swaggertype:"object"`
	ForceRebuild   *bool             `json:"force_rebuild" binding:"omitempty"`
}

func (req *BuildContainerRequest) Validate() error {
	req.ImageName = strings.TrimSpace(req.ImageName)
	req.GithubRepository = strings.TrimSpace(req.GithubRepository)

	if req.ImageName == "" {
		return fmt.Errorf("container image name cannot be empty")
	}
	if req.Tag != "" {
		req.Tag = strings.TrimSpace(req.Tag)
	}
	parts := strings.Split(req.GithubRepository, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}
	if req.GithubBranch != "" {
		req.GithubBranch = strings.TrimSpace(req.GithubBranch)
		if err := utils.IsValidGitHubBranch(req.GithubBranch); err != nil {
			return err
		}
	}
	if req.GithubCommit != "" {
		req.GithubCommit = strings.TrimSpace(req.GithubCommit)
		if err := utils.IsValidGitHubCommit(req.GithubCommit); err != nil {
			return err
		}
	}
	if req.GithubToken != "" {
		req.GithubToken = strings.TrimSpace(req.GithubToken)
		if err := utils.IsValidGitHubToken(req.GithubToken); err != nil {
			return err
		}
	}

	if req.Tag == "" {
		req.Tag = "latest"
	}
	if req.GithubBranch == "" {
		req.GithubBranch = "main"
	}
	if req.SubPath == "" {
		req.SubPath = "."
	}

	return req.Options.Validate()
}

func (req *BuildContainerRequest) ValidateInfoContent(sourcePath string) error {
	if req.ImageName == "" {
		tomlPath := filepath.Join(sourcePath, InfoFileName)
		content, err := utils.ReadTomlFile(tomlPath)
		if err != nil {
			return err
		}

		if name, ok := content[InfoNameField].(string); ok && name != "" {
			req.ImageName = name
		} else {
			return fmt.Errorf("%s does not contain a valid name field", InfoFileName)
		}
	}

	return nil
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

type CreateContainerRequest struct {
	Name     string               `json:"name" binding:"required"`
	Type     consts.ContainerType `json:"type" binding:"required"`
	README   string               `json:"readme" binding:"omitempty"`
	IsPublic *bool                `json:"is_public" binding:"omitempty"`

	VersionRequest *CreateContainerVersionRequest `json:"version" binding:"omitempty"`
}

func (req *CreateContainerRequest) Validate() error {
	req.Name = strings.TrimSpace(req.Name)

	if req.Name == "" {
		return fmt.Errorf("container name cannot be empty")
	}
	if req.Type == "" {
		return fmt.Errorf("container type cannot be empty")
	}
	if req.IsPublic == nil {
		req.IsPublic = utils.BoolPtr(true)
	}

	if req.Type != "" {
		if _, exists := consts.ValidContainerTypes[req.Type]; !exists {
			return fmt.Errorf("invalid container type: %s", req.Type)
		}
	}

	if req.VersionRequest != nil {
		if err := req.VersionRequest.Validate(); err != nil {
			return fmt.Errorf("invalid container version request: %v", err)
		}
	}

	return nil
}

func (req *CreateContainerRequest) ConvertToContainer() *database.Container {
	return &database.Container{
		Name:     req.Name,
		Type:     string(req.Type),
		README:   req.README,
		IsPublic: *req.IsPublic,
		Status:   consts.CommonEnabled,
	}
}

type CreateContainerVersionRequest struct {
	Name              string                   `json:"name" binding:"required"`
	GithubLink        string                   `json:"github_link" binding:"omitempty"`
	ImageRef          string                   `json:"image_ref" binding:"required"`
	Command           string                   `json:"command" binding:"omitempty"`
	EnvVars           []string                 `json:"env_vars" binding:"omitempty"`
	HelmConfigRequest *CreateHelmConfigRequest `json:"helm_config" binding:"omitempty"`
}

func (req *CreateContainerVersionRequest) Validate() error {
	req.Name = strings.TrimSpace(req.Name)
	req.ImageRef = strings.TrimSpace(req.ImageRef)

	if req.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if req.ImageRef == "" {
		return fmt.Errorf("docker image reference cannot be empty")
	}

	if req.GithubLink != "" {
		req.GithubLink = strings.TrimSpace(req.GithubLink)
		if err := utils.IsValidGitHubLink(req.GithubLink); err != nil {
			return fmt.Errorf("invalid github link: %s, %v", req.GithubLink, err)
		}
	}
	if _, _, _, err := utils.ParseSemanticVersion(req.Name); err != nil {
		return fmt.Errorf("invalid semantic version: %s, %v", req.Name, err)
	}
	if _, _, _, _, err := utils.ParseFullImageRefernce(req.ImageRef); err != nil {
		return fmt.Errorf("invalid docker image reference: %s, %v", req.ImageRef, err)
	}

	if req.HelmConfigRequest != nil {
		if err := req.HelmConfigRequest.Validate(); err != nil {
			return fmt.Errorf("invalid helm config: %v", err)
		}
	}

	return nil
}

func (req *CreateContainerVersionRequest) ConvertToContainerVersion() *database.ContainerVersion {
	version := &database.ContainerVersion{
		Name:     req.Name,
		ImageRef: req.ImageRef,
		Command:  req.Command,
		EnvVars:  strings.Join(req.EnvVars, ","),
		Status:   consts.CommonEnabled,
	}

	return version
}

type CreateHelmConfigRequest struct {
	ChartName    string         `json:"chart_name" binding:"required"`
	RepoName     string         `json:"repo_name" binding:"required"`
	RepoURL      string         `json:"repo_url" binding:"required"`
	NsPrefix     string         `json:"ns_prefix" binding:"required"`
	PortTemplate string         `json:"port_template" binding:"omitempty"`
	Values       map[string]any `json:"values" binding:"omitempty" swaggertype:"object"`
}

func (req *CreateHelmConfigRequest) Validate() error {
	req.ChartName = strings.TrimSpace(req.ChartName)
	req.RepoName = strings.TrimSpace(req.RepoName)
	req.RepoURL = strings.TrimSpace(req.RepoURL)
	req.NsPrefix = strings.TrimSpace(req.NsPrefix)

	if req.ChartName == "" {
		return fmt.Errorf("chart name cannot be empty")
	}
	if req.RepoName == "" {
		return fmt.Errorf("repository name cannot be empty")
	}
	if req.RepoURL == "" {
		return fmt.Errorf("repository URL cannot be empty")
	}
	if req.NsPrefix == "" {
		return fmt.Errorf("namespace prefix cannot be empty")
	}

	if _, err := url.ParseRequestURI(req.RepoURL); err != nil {
		return fmt.Errorf("invalid repository URL: %s, %v", req.RepoURL, err)
	}
	if !utils.CheckNsPrefixExists(req.NsPrefix) {
		return fmt.Errorf("invalid namespace prefix: %s", req.NsPrefix)
	}
	if req.PortTemplate != "" {
		req.PortTemplate = strings.TrimSpace(req.PortTemplate)
		if !strings.Contains(req.PortTemplate, "{{.port}}") {
			return fmt.Errorf("port template must contain '{{.port}}' placeholder")
		}
	}

	return nil
}

func (req *CreateHelmConfigRequest) ConvertToHelmConfig() (*database.HelmConfig, error) {
	var valuesJSON string

	if len(req.Values) > 0 {
		data, err := json.Marshal(req.Values)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize helm values to JSON: %v", err)
		}
		valuesJSON = string(data)
	} else {
		valuesJSON = "{}"
	}

	return &database.HelmConfig{
		ChartName:    req.ChartName,
		RepoName:     req.RepoName,
		RepoURL:      req.RepoURL,
		NsPrefix:     req.NsPrefix,
		PortTemplate: req.PortTemplate,
		Values:       valuesJSON,
	}, nil
}

// ListContainerRequest represents container list query parameters
type ListContainerRequest struct {
	PaginationRequest
	Type   consts.ContainerType `json:"type" binding:"omitempty"`
	Status *int                 `json:"status" binding:"omitempty"`
}

func (req *ListContainerRequest) Validate() error {
	if req.Type != "" {
		if _, exists := consts.ValidContainerTypes[req.Type]; !exists {
			return fmt.Errorf("invalid container type: %s", req.Type)
		}
	}

	return validateStatusField(req.Status, false)
}

// SearchContainerRequest represents advanced role search with complex filtering
type SearchContainerRequest struct {
	AdvancedSearchRequest

	// Container-specific filters
	Name    *string `json:"name,omitempty"`
	Image   *string `json:"image,omitempty"`
	Tag     *string `json:"tag,omitempty"`
	Type    *string `json:"type,omitempty"`
	Command *string `json:"command,omitempty"`
	Status  *int    `json:"status,omitempty"`
}

// ConvertToSearchRequest converts SearchContainerRequest to SearchRequest
func (csr *SearchContainerRequest) ConvertToSearchRequest() *SearchRequest {
	sr := csr.AdvancedSearchRequest.ConvertAdvancedToSearch()

	// Add container-specific filters
	if csr.Name != nil {
		sr.AddFilter("name", OpLike, *csr.Name)
	}
	if csr.Image != nil {
		sr.AddFilter("image", OpLike, *csr.Image)
	}
	if csr.Tag != nil {
		sr.AddFilter("tag", OpEqual, *csr.Tag)
	}
	if csr.Type != nil {
		sr.AddFilter("type", OpEqual, *csr.Type)
	}
	if csr.Command != nil {
		sr.AddFilter("command", OpLike, *csr.Command)
	}

	return sr
}

// UpdateContainerRequest represents the request for updating a container
type UpdateContainerRequest struct {
	README   *string `json:"readme" binding:"omitempty"`
	IsPublic *bool   `json:"is_public" binding:"omitempty"`
	Status   *int    `json:"status" binding:"omitempty"`
}

func (req *UpdateContainerRequest) Validate() error {
	if req.Status != nil {
		return validateStatusField(req.Status, true)
	}
	return nil
}

func (req *UpdateContainerRequest) PatchContainerModel(target *database.Container) {
	if req.README != nil {
		target.README = *req.README
	}
	if req.IsPublic != nil {
		target.IsPublic = *req.IsPublic
	}
	if req.Status != nil {
		target.Status = *req.Status
	}
}

// UpdateContainerVersionRequest represents the request for updating a container version
type UpdateContainerVersionRequest struct {
	GithubLink        *string                  `json:"github_link" binding:"omitempty"`
	Command           *string                  `json:"command" binding:"omitempty"`
	EnvVars           *[]string                `json:"env_vars" binding:"omitempty"`
	Status            *int                     `json:"status" binding:"omitempty"`
	HelmConfigRequest *UpdateHelmConfigRequest `json:"helm_config" binding:"omitempty"`
}

func (req *UpdateContainerVersionRequest) Validate() error {
	if req.GithubLink != nil {
		trimmedLink := strings.TrimSpace(*req.GithubLink)
		*req.GithubLink = trimmedLink

		if trimmedLink != "" {
			if err := utils.IsValidGitHubLink(trimmedLink); err != nil {
				return fmt.Errorf("invalid GitHub link '%s': %v", trimmedLink, err)
			}
		}
	}
	if req.Command != nil {
		*req.Command = strings.TrimSpace(*req.Command)
	}
	if req.Status != nil {
		if _, exists := consts.ValidCommonStatus[*req.Status]; !exists {
			return fmt.Errorf("invalid status value: %d", *req.Status)
		}
	}

	if req.HelmConfigRequest != nil {
		if err := req.HelmConfigRequest.Validate(); err != nil {
			return fmt.Errorf("invalid helm config: %v", err)
		}
	}

	return nil
}

func (req *UpdateContainerVersionRequest) PatchContainerVersionModel(target *database.ContainerVersion) {
	if req.GithubLink != nil {
		target.GithubLink = *req.GithubLink
	}
	if req.Command != nil {
		target.Command = *req.Command
	}
	if req.EnvVars != nil {
		target.EnvVars = strings.Join(*req.EnvVars, ",")
	}
	if req.Status != nil {
		target.Status = *req.Status
	}
}

type UpdateHelmConfigRequest struct {
	RepoURL      *string         `json:"repo_url" binding:"omitempty"`
	RepoName     *string         `json:"repo_name" binding:"omitempty"`
	ChartName    *string         `json:"chart_name" binding:"omitempty"`
	NsPrefix     *string         `json:"ns_prefix" binding:"omitempty"`
	PortTemplate *string         `json:"port_template" binding:"omitempty"`
	Values       *map[string]any `json:"values" binding:"omitempty" swaggertype:"object"`
}

func (req *UpdateHelmConfigRequest) Validate() error {
	if req.RepoURL != nil {
		trimmedURL := strings.TrimSpace(*req.RepoURL)
		*req.RepoURL = trimmedURL

		if trimmedURL == "" {
			return fmt.Errorf("repository URL cannot be empty if provided")
		}
		if _, err := url.Parse(trimmedURL); err != nil {
			return fmt.Errorf("invalid repository URL format: %s. Error: %v", trimmedURL, err)
		}
	}
	if req.RepoName != nil {
		*req.RepoName = strings.TrimSpace(*req.RepoName)
	}
	if req.ChartName != nil {
		*req.ChartName = strings.TrimSpace(*req.ChartName)
	}
	if req.NsPrefix != nil {
		trimmedPrefix := strings.TrimSpace(*req.NsPrefix)
		*req.NsPrefix = trimmedPrefix

		if trimmedPrefix == "" {
			return fmt.Errorf("namespace prefix cannot be empty if provided")
		}
		if !utils.CheckNsPrefixExists(trimmedPrefix) {
			return fmt.Errorf("invalid namespace prefix: '%s'. It must exist or adhere to naming rules", trimmedPrefix)
		}
	}
	if req.PortTemplate != nil {
		trimmedTemplate := strings.TrimSpace(*req.PortTemplate)
		*req.PortTemplate = trimmedTemplate

		if trimmedTemplate != "" {
			const placeholder = "{{.port}}"
			if !strings.Contains(trimmedTemplate, placeholder) {
				return fmt.Errorf("port template '%s' must contain the dynamic placeholder '%s'", trimmedTemplate, placeholder)
			}
		}
	}

	return nil
}

func (req *UpdateHelmConfigRequest) PatchHelmConfigModel(target *database.HelmConfig) error {
	if req.RepoURL != nil {
		target.RepoURL = *req.RepoURL
	}
	if req.RepoName != nil {
		target.RepoName = *req.RepoName
	}
	if req.ChartName != nil {
		target.ChartName = *req.ChartName
	}
	if req.NsPrefix != nil {
		target.NsPrefix = *req.NsPrefix
	}
	if req.PortTemplate != nil {
		target.PortTemplate = *req.PortTemplate
	}
	if req.Values != nil {
		if len(*req.Values) > 0 {
			data, err := json.Marshal(*req.Values)
			if err != nil {
				return fmt.Errorf("failed to serialize helm values to JSON: %v", err)
			}
			target.Values = string(data)
		} else {
			target.Values = "{}"
		}
	}

	return nil
}

// ContainerResponse is basic container info used
type ContainerResponse struct {
	ID        int                  `json:"id"`
	Name      string               `json:"name"`
	Type      consts.ContainerType `json:"type"`
	IsPublic  bool                 `json:"is_public"`
	Status    int                  `json:"status"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// ConvertFromContainer converts database Container to ContainerResponse DTO
func (c *ContainerResponse) ConvertFromContainer(container *database.Container) {
	c.ID = container.ID
	c.Name = container.Name
	c.Type = consts.ContainerType(container.Type)
	c.IsPublic = container.IsPublic
	c.Status = container.Status
	c.CreatedAt = container.CreatedAt
	c.UpdatedAt = container.UpdatedAt
}

// ContainerDetailResponse is used for single resource retrieval.
type ContainerDetailResponse struct {
	ContainerResponse

	README string `json:"readme"`

	Versions []ContainerVersionResponse `json:"versions"`
}

func (c *ContainerDetailResponse) ConvertFromContainer(container *database.Container) {
	c.ContainerResponse.ConvertFromContainer(container)
	c.README = container.README
}

type ContainerVersionResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	ImageRef  string    `json:"image_ref"`
	Usage     int       `json:"usage"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c *ContainerVersionResponse) ConvertFromContainerVersion(version *database.ContainerVersion) {
	c.ID = version.ID
	c.Name = version.Name
	c.ImageRef = version.ImageRef
	c.Usage = version.Usage
	c.UpdatedAt = version.UpdatedAt
}

type ContainerVersionDetailResponse struct {
	ContainerVersionResponse

	GithubLink string `json:"github_link"`
	Command    string `json:"command"`
	EnvVars    string `json:"env_vars"`

	HelmConfig *HelmConfigDetailResponse `json:"helm_config,omitempty"`
}

func (c *ContainerVersionDetailResponse) ConvertFromContainerVersion(version *database.ContainerVersion) {
	c.ContainerVersionResponse.ConvertFromContainerVersion(version)
	c.GithubLink = version.GithubLink
	c.Command = version.Command
	c.EnvVars = version.EnvVars
}

type ListContainerVersionResponse struct {
	Items      []ContainerResponse `json:"items"`
	Pagination PaginationInfo      `json:"pagination"`
}

type HelmConfigDetailResponse struct {
	ID           int            `json:"id"`
	RepoURL      string         `json:"repo_url"`
	FullChart    string         `json:"full_chart"`
	NsPrefix     string         `json:"ns_prefix"`
	PortTemplate string         `json:"port_template"`
	Values       map[string]any `json:"values"`
}

func (h *HelmConfigDetailResponse) ConvertFromHelmConfig(cfg *database.HelmConfig) error {
	h.ID = cfg.ID
	h.RepoURL = cfg.RepoURL
	h.FullChart = cfg.FullChart
	h.NsPrefix = cfg.NsPrefix
	h.PortTemplate = cfg.PortTemplate

	if cfg.Values != "" {
		var valuesMap map[string]any
		if err := json.Unmarshal([]byte(cfg.Values), &valuesMap); err != nil {
			return fmt.Errorf("failed to unmarshal Helm values JSON for config ID %d: %w", cfg.ID, err)
		}
		h.Values = valuesMap
	} else {
		h.Values = make(map[string]any)
	}

	return nil
}
