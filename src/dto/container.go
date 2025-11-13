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

// =====================================================================
// Container Service DTOs
// =====================================================================

type HelmConfigItem struct {
	RepoURL      string         `json:"repo_url"`
	RepoName     string         `json:"repo_name"`
	ChartName    string         `json:"chart_name"`
	FullChart    string         `json:"full_chart"`
	NsPrefix     string         `json:"ns_prefix"`
	PortTemplate string         `json:"port_template"`
	Values       map[string]any `json:"values"`
}

func NewHelmConfigItem(cfg *database.HelmConfig) (*HelmConfigItem, error) {
	var values map[string]any
	if err := json.Unmarshal([]byte(cfg.Values), &values); err != nil {
		return nil, fmt.Errorf("failed to unmarshal helm config values: %w", err)
	}

	return &HelmConfigItem{
		RepoURL:      cfg.RepoURL,
		RepoName:     cfg.RepoName,
		ChartName:    cfg.ChartName,
		FullChart:    cfg.FullChart,
		NsPrefix:     cfg.NsPrefix,
		PortTemplate: cfg.PortTemplate,
		Values:       values,
	}, nil
}

type PedestalInfo struct {
	Registry   string          `json:"registry"`
	Namespace  string          `json:"namespace"`
	Tag        string          `json:"tag"`
	HelmConfig *HelmConfigItem `json:"helm_config"`
}

func NewPedestalInfo(version *database.ContainerVersion, helmConfig *database.HelmConfig) (*PedestalInfo, error) {
	helmConfigItem, err := NewHelmConfigItem(helmConfig)
	if err != nil {
		return nil, err
	}

	return &PedestalInfo{
		Registry:   version.Registry,
		Namespace:  version.Namespace,
		Tag:        version.Tag,
		HelmConfig: helmConfigItem,
	}, nil
}

type ContainerVersionItem struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ImageRef    string `json:"image_ref"`
	Command     string `json:"command,omitempty"`
	EnvVarsKeys string `json:"env_vars_keys,omitempty"`

	ContainerID   int    `json:"container_id"`
	ContainerName string `json:"container_name"`

	EnvVars map[string]string `json:"env_vars,omitempty"`
	Extra   *PedestalInfo     `json:"extra,omitempty"`
}

func NewContainerVersionItem(version *database.ContainerVersion, envVars map[string]string) ContainerVersionItem {
	item := ContainerVersionItem{
		ID:          version.ID,
		Name:        version.Name,
		ImageRef:    version.ImageRef,
		Command:     version.Command,
		EnvVarsKeys: version.EnvVars,
	}

	if version.Container != nil {
		item.ContainerID = version.Container.ID
		item.ContainerName = version.Container.Name
	}

	item.EnvVars = envVars
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
	EnvVars map[string]string `json:"env_vars" binding:"omitempty" swaggertype:"object"`
}

func (item *ContainerSpec) Validate() error {
	if err := item.ContainerRef.Validate(); err != nil {
		return err
	}
	for key := range item.EnvVars {
		if err := utils.IsValidEnvVar(key); err != nil {
			return fmt.Errorf("invalid environment variable key %s: %v", key, err)
		}
	}
	return nil
}

// =====================================================================
// Container CRUD DTOs
// =====================================================================

type CreateContainerReq struct {
	Name     string               `json:"name" binding:"required"`
	Type     consts.ContainerType `json:"type" binding:"required"`
	README   string               `json:"readme" binding:"omitempty"`
	IsPublic *bool                `json:"is_public" binding:"omitempty"`

	VersionReq *CreateContainerVersionReq `json:"version" binding:"omitempty"`
}

func (req *CreateContainerReq) Validate() error {
	req.Name = strings.TrimSpace(req.Name)

	if req.Name == "" {
		return fmt.Errorf("container name cannot be empty")
	}
	if req.IsPublic == nil {
		req.IsPublic = utils.BoolPtr(true)
	}

	if _, exists := consts.ValidContainerTypes[req.Type]; !exists {
		return fmt.Errorf("invalid container type: %d", req.Type)
	}

	if req.VersionReq != nil {
		if err := req.VersionReq.Validate(); err != nil {
			return fmt.Errorf("invalid container version request: %v", err)
		}
	}

	return nil
}

func (req *CreateContainerReq) ConvertToContainer() *database.Container {
	return &database.Container{
		Name:     req.Name,
		Type:     req.Type,
		README:   req.README,
		IsPublic: *req.IsPublic,
		Status:   consts.CommonEnabled,
	}
}

type CreateContainerVersionReq struct {
	Name              string                   `json:"name" binding:"required"`
	GithubLink        string                   `json:"github_link" binding:"omitempty"`
	ImageRef          string                   `json:"image_ref" binding:"required"`
	Command           string                   `json:"command" binding:"omitempty"`
	EnvVars           []string                 `json:"env_vars" binding:"omitempty"`
	HelmConfigRequest *CreateHelmConfigRequest `json:"helm_config" binding:"omitempty"`
}

func (req *CreateContainerVersionReq) Validate() error {
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

func (req *CreateContainerVersionReq) ConvertToContainerVersion() *database.ContainerVersion {
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

// ListContainerReq represents container list query parameters
type ListContainerReq struct {
	PaginationReq
	Type     *consts.ContainerType `json:"type" binding:"omitempty"`
	IsPublic *bool                 `json:"is_public" binding:"omitempty"`
	Status   *consts.StatusType    `json:"status" binding:"omitempty"`
}

func (req *ListContainerReq) Validate() error {
	if req.Type != nil {
		if _, exists := consts.ValidContainerTypes[*req.Type]; !exists {
			return fmt.Errorf("invalid container type: %d", req.Type)
		}
	}

	return validateStatusField(req.Status, false)
}

// ListContainerVersionReq represents container version list query parameters
type ListContainerVersionReq struct {
	PaginationReq
	Status *consts.StatusType `json:"status" binding:"omitempty"`
}

func (req *ListContainerVersionReq) Validate() error {
	return validateStatusField(req.Status, false)
}

// SearchContainerReq represents container search request
type SearchContainerReq struct {
	AdvancedSearchReq

	// Container-specific filters
	Name    *string `json:"name,omitempty"`
	Image   *string `json:"image,omitempty"`
	Tag     *string `json:"tag,omitempty"`
	Type    *string `json:"type,omitempty"`
	Command *string `json:"command,omitempty"`
	Status  *int    `json:"status,omitempty"`
}

// ConvertToSearchRequest converts SearchContainerReq to SearchRequest
func (csr *SearchContainerReq) ConvertToSearchRequest() *SearchReq {
	sr := csr.AdvancedSearchReq.ConvertAdvancedToSearch()

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

// UpdateContainerReq represents the request for updating a container
type UpdateContainerReq struct {
	README   *string            `json:"readme" binding:"omitempty"`
	IsPublic *bool              `json:"is_public" binding:"omitempty"`
	Status   *consts.StatusType `json:"status" binding:"omitempty"`
}

func (req *UpdateContainerReq) Validate() error {
	return validateStatusField(req.Status, true)
}

func (req *UpdateContainerReq) PatchContainerModel(target *database.Container) {
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

// SubmitBuildContainerReq represents the request for building a container into platform registry
type SubmitBuildContainerReq struct {
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

func (req *SubmitBuildContainerReq) Validate() error {
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

func (req *SubmitBuildContainerReq) ValidateInfoContent(sourcePath string) error {
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

// UpdateContainerVersionReq represents the request for updating a container version
type UpdateContainerVersionReq struct {
	GithubLink        *string                  `json:"github_link" binding:"omitempty"`
	Command           *string                  `json:"command" binding:"omitempty"`
	EnvVars           *[]string                `json:"env_vars" binding:"omitempty"`
	Status            *consts.StatusType       `json:"status" binding:"omitempty"`
	HelmConfigRequest *UpdateHelmConfigRequest `json:"helm_config" binding:"omitempty"`
}

func (req *UpdateContainerVersionReq) Validate() error {
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
		if err := validateStatusField(req.Status, true); err != nil {
			return err
		}
	}

	if req.HelmConfigRequest != nil {
		if err := req.HelmConfigRequest.Validate(); err != nil {
			return fmt.Errorf("invalid helm config: %v", err)
		}
	}

	return nil
}

func (req *UpdateContainerVersionReq) PatchContainerVersionModel(target *database.ContainerVersion) {
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

// ContainerResp is basic container info used
type ContainerResp struct {
	ID        int                  `json:"id"`
	Name      string               `json:"name"`
	Type      consts.ContainerType `json:"type"`
	IsPublic  bool                 `json:"is_public"`
	Status    consts.StatusType    `json:"status"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`

	Labels []LabelItem `json:"labels,omitempty"`
}

func NewContainerResp(container *database.Container) *ContainerResp {
	resp := &ContainerResp{
		ID:        container.ID,
		Name:      container.Name,
		Type:      consts.ContainerType(container.Type),
		IsPublic:  container.IsPublic,
		Status:    container.Status,
		CreatedAt: container.CreatedAt,
		UpdatedAt: container.UpdatedAt,
	}

	if len(container.Labels) > 0 {
		resp.Labels = make([]LabelItem, 0, len(container.Labels))
		for _, l := range container.Labels {
			resp.Labels = append(resp.Labels, LabelItem{
				Key:   l.Key,
				Value: l.Value,
			})
		}
	}
	return resp
}

// ContainerDetailResp is used for single resource retrieval.
type ContainerDetailResp struct {
	ContainerResp

	README string `json:"readme"`

	Versions []ContainerVersionResp `json:"versions"`
}

func NewContainerDetailResp(container *database.Container) *ContainerDetailResp {
	return &ContainerDetailResp{
		ContainerResp: *NewContainerResp(container),
		README:        container.README,
	}
}

type ContainerVersionResp struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	ImageRef  string    `json:"image_ref"`
	Usage     int       `json:"usage"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewContainerVersionResp(version *database.ContainerVersion) *ContainerVersionResp {
	return &ContainerVersionResp{
		ID:        version.ID,
		Name:      version.Name,
		ImageRef:  version.ImageRef,
		Usage:     version.Usage,
		UpdatedAt: version.UpdatedAt,
	}
}

type ContainerVersionDetailResp struct {
	ContainerVersionResp

	GithubLink string `json:"github_link"`
	Command    string `json:"command"`
	EnvVars    string `json:"env_vars"`

	HelmConfig *HelmConfigDetailResp `json:"helm_config,omitempty"`
}

func NewContainerVersionDetailResp(version *database.ContainerVersion) *ContainerVersionDetailResp {
	return &ContainerVersionDetailResp{
		ContainerVersionResp: *NewContainerVersionResp(version),
		GithubLink:           version.GithubLink,
		Command:              version.Command,
		EnvVars:              version.EnvVars,
	}
}

type ListContainerVersionResp struct {
	Items      []ContainerResp `json:"items"`
	Pagination PaginationInfo  `json:"pagination"`
}

type HelmConfigDetailResp struct {
	ID           int            `json:"id"`
	RepoURL      string         `json:"repo_url"`
	FullChart    string         `json:"full_chart"`
	NsPrefix     string         `json:"ns_prefix"`
	PortTemplate string         `json:"port_template"`
	Values       map[string]any `json:"values"`
}

func NewHelmConfigDetailResp(cfg *database.HelmConfig) (*HelmConfigDetailResp, error) {
	resp := &HelmConfigDetailResp{
		ID:           cfg.ID,
		RepoURL:      cfg.RepoURL,
		FullChart:    cfg.FullChart,
		NsPrefix:     cfg.NsPrefix,
		PortTemplate: cfg.PortTemplate,
	}

	if cfg.Values != "" {
		var valuesMap map[string]any
		if err := json.Unmarshal([]byte(cfg.Values), &valuesMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Helm values JSON for config ID %d: %w", cfg.ID, err)
		}
		resp.Values = valuesMap
	} else {
		resp.Values = make(map[string]any)
	}

	return resp, nil
}

type SubmitContainerBuildResp struct {
	GroupID string `json:"group_id"`
	TraceID string `json:"trace_id"`
	TaskID  string `json:"task_id"`
}

// ---------------------- Container Label DTOs ------------------

// ManageContainerLabelReq represents the request for managing container labels
type ManageContainerLabelReq struct {
	AddLabels    []LabelItem `json:"add_labels" binding:"omitempty"`    // List of labels to add
	RemoveLabels []string    `json:"remove_labels" binding:"omitempty"` // List of label keys to remove
}

func (req *ManageContainerLabelReq) Validate() error {
	if len(req.AddLabels) == 0 && len(req.RemoveLabels) == 0 {
		return fmt.Errorf("at least one of add_labels or remove_labels must be provided")
	}

	if err := validateLabelItemsFiled(req.AddLabels); err != nil {
		return err
	}

	for i, key := range req.RemoveLabels {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("empty label key at index %d in remove_labels", i)
		}
	}

	return nil
}
