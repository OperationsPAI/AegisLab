package container

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	"aegis/utils"
)

// CreateContainerReq represents container creation request.
type CreateContainerReq struct {
	Name     string                `json:"name" binding:"required"`
	Type     *consts.ContainerType `json:"type"`
	README   string                `json:"readme" binding:"omitempty"`
	IsPublic *bool                 `json:"is_public"`

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
	if req.Type == nil {
		return fmt.Errorf("container type is required")
	}
	if err := validateContainerType(req.Type); err != nil {
		return err
	}
	if req.VersionReq != nil {
		if err := req.VersionReq.Validate(); err != nil {
			return fmt.Errorf("invalid container version request: %v", err)
		}
	}

	return nil
}

func (req *CreateContainerReq) ConvertToContainer() *model.Container {
	container := &model.Container{
		Name:     req.Name,
		Type:     *req.Type,
		README:   req.README,
		IsPublic: *req.IsPublic,
		Status:   consts.CommonEnabled,
	}

	if req.VersionReq != nil {
		container.Versions = []model.ContainerVersion{
			*req.VersionReq.ConvertToContainerVersion(),
		}
	}

	return container
}

// ListContainerReq represents container list query parameters.
type ListContainerReq struct {
	dto.PaginationReq
	Type     *consts.ContainerType `form:"type"`
	IsPublic *bool                 `form:"is_public"`
	Status   *consts.StatusType    `form:"status"`
}

func (req *ListContainerReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if err := validateContainerType(req.Type); err != nil {
		return err
	}
	return validateStatus(req.Status, false)
}

// UpdateContainerReq represents container update request.
type UpdateContainerReq struct {
	README   *string            `json:"readme" binding:"omitempty"`
	IsPublic *bool              `json:"is_public" binding:"omitempty"`
	Status   *consts.StatusType `json:"status" binding:"omitempty"`
}

func (req *UpdateContainerReq) Validate() error {
	return validateStatus(req.Status, true)
}

func (req *UpdateContainerReq) PatchContainerModel(target *model.Container) {
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

// ManageContainerLabelReq represents container label management request.
type ManageContainerLabelReq struct {
	AddLabels    []dto.LabelItem `json:"add_labels" binding:"omitempty"`
	RemoveLabels []string        `json:"remove_labels" binding:"omitempty"`
}

func (req *ManageContainerLabelReq) Validate() error {
	if len(req.AddLabels) == 0 && len(req.RemoveLabels) == 0 {
		return fmt.Errorf("at least one of add_labels or remove_labels must be provided")
	}
	if err := validateLabelItems(req.AddLabels); err != nil {
		return err
	}
	for i, key := range req.RemoveLabels {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("empty label key at index %d in remove_labels", i)
		}
	}
	return nil
}

// ContainerResp represents basic container summary information.
type ContainerResp struct {
	ID        int             `json:"id"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	IsPublic  bool            `json:"is_public"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Labels    []dto.LabelItem `json:"labels,omitempty"`
}

func NewContainerResp(container *model.Container) *ContainerResp {
	resp := &ContainerResp{
		ID:        container.ID,
		Name:      container.Name,
		Type:      consts.GetContainerTypeName(container.Type),
		IsPublic:  container.IsPublic,
		Status:    consts.GetStatusTypeName(container.Status),
		CreatedAt: container.CreatedAt,
		UpdatedAt: container.UpdatedAt,
	}

	if len(container.Labels) > 0 {
		resp.Labels = make([]dto.LabelItem, 0, len(container.Labels))
		for _, label := range container.Labels {
			resp.Labels = append(resp.Labels, dto.LabelItem{Key: label.Key, Value: label.Value})
		}
	}
	return resp
}

// ContainerDetailResp represents detailed container information.
type ContainerDetailResp struct {
	ContainerResp

	README string `json:"readme"`

	Versions []ContainerVersionResp `json:"versions"`
}

func NewContainerDetailResp(container *model.Container) *ContainerDetailResp {
	return &ContainerDetailResp{
		ContainerResp: *NewContainerResp(container),
		README:        container.README,
	}
}

type CreateContainerVersionReq struct {
	Name              string                     `json:"name" binding:"required"`
	GithubLink        string                     `json:"github_link" binding:"omitempty"`
	ImageRef          string                     `json:"image_ref" binding:"required"`
	Command           string                     `json:"command" binding:"omitempty"`
	EnvVarRequests    []CreateParameterConfigReq `json:"env_vars" binding:"omitempty"`
	HelmConfigRequest *CreateHelmConfigReq       `json:"helm_config" binding:"omitempty"`
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

	for idx, envVarReq := range req.EnvVarRequests {
		if err := envVarReq.Validate(); err != nil {
			return fmt.Errorf("invalid env var at index %d: %v", idx, err)
		}
	}
	if req.HelmConfigRequest != nil {
		if err := req.HelmConfigRequest.Validate(); err != nil {
			return fmt.Errorf("invalid helm config:  %v", err)
		}
	}

	return nil
}

func (req *CreateContainerVersionReq) ConvertToContainerVersion() *model.ContainerVersion {
	version := &model.ContainerVersion{
		Name:     req.Name,
		ImageRef: req.ImageRef,
		Command:  req.Command,
		Status:   consts.CommonEnabled,
	}

	if len(req.EnvVarRequests) > 0 {
		params := make([]model.ParameterConfig, 0, len(req.EnvVarRequests))
		for _, envVarReq := range req.EnvVarRequests {
			params = append(params, *envVarReq.ConvertToParameterConfig())
		}
		version.EnvVars = params
	}

	if req.HelmConfigRequest != nil {
		version.HelmConfig = req.HelmConfigRequest.ConvertToHelmConfig()
	}

	return version
}

type CreateHelmConfigReq struct {
	Version       string                     `json:"version" binding:"required"`
	ChartName     string                     `json:"chart_name" binding:"required"`
	RepoName      string                     `json:"repo_name" binding:"required"`
	RepoURL       string                     `json:"repo_url" binding:"required"`
	DynamicValues []CreateParameterConfigReq `json:"dynamic_values" binding:"omitempty" swaggertype:"object"`
}

func (req *CreateHelmConfigReq) Validate() error {
	req.Version = strings.TrimSpace(req.Version)
	req.ChartName = strings.TrimSpace(req.ChartName)
	req.RepoName = strings.TrimSpace(req.RepoName)
	req.RepoURL = strings.TrimSpace(req.RepoURL)

	if req.Version == "" {
		if _, _, _, err := utils.ParseSemanticVersion(req.Version); err != nil {
			return fmt.Errorf("invalid semantic version: %s, %v", req.Version, err)
		}
	}
	if req.ChartName == "" {
		return fmt.Errorf("chart name cannot be empty")
	}
	if req.RepoName == "" {
		return fmt.Errorf("repository name cannot be empty")
	}
	if req.RepoURL == "" {
		return fmt.Errorf("repository URL cannot be empty")
	}
	if _, err := url.ParseRequestURI(req.RepoURL); err != nil {
		return fmt.Errorf("invalid repository URL: %s, %w", req.RepoURL, err)
	}
	for i, val := range req.DynamicValues {
		if err := val.Validate(); err != nil {
			return fmt.Errorf("invalid parameter config at index %d: %w", i, err)
		}
	}

	return nil
}

func (req *CreateHelmConfigReq) ConvertToHelmConfig() *model.HelmConfig {
	cfg := &model.HelmConfig{
		Version:   req.Version,
		ChartName: req.ChartName,
		RepoName:  req.RepoName,
		RepoURL:   req.RepoURL,
	}

	if len(req.DynamicValues) > 0 {
		params := make([]model.ParameterConfig, 0, len(req.DynamicValues))
		for _, val := range req.DynamicValues {
			params = append(params, *val.ConvertToParameterConfig())
		}
		cfg.DynamicValues = params
	}

	return cfg
}

type CreateParameterConfigReq struct {
	Key            string                   `json:"key" binding:"required"`
	Type           consts.ParameterType     `json:"type" binding:"required"`
	Category       consts.ParameterCategory `json:"category" binding:"required"`
	ValueType      consts.ValueDataType     `json:"value_type" binding:"omitempty"`
	Description    string                   `json:"description" binding:"omitempty"`
	DefaultValue   *string                  `json:"default_value" binding:"omitempty"`
	TemplateString *string                  `json:"template_string" binding:"omitempty"`
	Required       bool                     `json:"required"`
	Overridable    *bool                    `json:"overridable" binding:"omitempty"`
}

func (req *CreateParameterConfigReq) Validate() error {
	if req.Key == "" {
		return fmt.Errorf("parameter key cannot be empty")
	}
	if _, exists := consts.ValidParameterTypes[req.Type]; !exists {
		return fmt.Errorf("invalid parameter type: %v", req.Type)
	}
	if _, exists := consts.ValidParameterCategories[req.Category]; !exists {
		return fmt.Errorf("invalid parameter category: %v", req.Category)
	}
	if req.Type == consts.ParameterTypeFixed && req.Required && req.DefaultValue == nil {
		return fmt.Errorf("default value is required for fixed parameter type when marked as required")
	}
	if req.Type == consts.ParameterTypeDynamic && req.TemplateString == nil {
		return fmt.Errorf("template string is required for dynamic parameter type")
	}
	return nil
}

func (req *CreateParameterConfigReq) ConvertToParameterConfig() *model.ParameterConfig {
	config := &model.ParameterConfig{
		Key:            req.Key,
		Type:           req.Type,
		Category:       req.Category,
		ValueType:      req.ValueType,
		Description:    req.Description,
		DefaultValue:   req.DefaultValue,
		TemplateString: req.TemplateString,
		Required:       req.Required,
		Overridable:    true,
	}
	if req.Overridable != nil {
		config.Overridable = *req.Overridable
	}
	return config
}

type ListContainerVersionReq struct {
	dto.PaginationReq
	Status *consts.StatusType `json:"status" binding:"omitempty"`
}

func (req *ListContainerVersionReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	return validateStatus(req.Status, false)
}

type SearchContainerReq struct {
	dto.AdvancedSearchReq[string]
	Name    *string `json:"name,omitempty"`
	Image   *string `json:"image,omitempty"`
	Tag     *string `json:"tag,omitempty"`
	Type    *string `json:"type,omitempty"`
	Command *string `json:"command,omitempty"`
	Status  *int    `json:"status,omitempty"`
}

func (csr *SearchContainerReq) ConvertToSearchRequest() *dto.SearchReq[string] {
	sr := csr.ConvertAdvancedToSearch()
	if csr.Name != nil {
		sr.AddFilter("name", dto.OpLike, *csr.Name)
	}
	if csr.Image != nil {
		sr.AddFilter("image", dto.OpLike, *csr.Image)
	}
	if csr.Tag != nil {
		sr.AddFilter("tag", dto.OpEqual, *csr.Tag)
	}
	if csr.Type != nil {
		sr.AddFilter("type", dto.OpEqual, *csr.Type)
	}
	if csr.Command != nil {
		sr.AddFilter("command", dto.OpLike, *csr.Command)
	}
	return sr
}

type SubmitBuildContainerReq struct {
	ImageName        string            `json:"image_name" binding:"required"`
	Tag              string            `json:"tag" binding:"omitempty"`
	GithubRepository string            `json:"github_repository" binding:"required"`
	GithubBranch     string            `json:"github_branch" binding:"omitempty"`
	GithubCommit     string            `json:"github_commit" binding:"omitempty"`
	GithubToken      string            `json:"github_token" binding:"omitempty"`
	SubPath          string            `json:"sub_path" binding:"omitempty"`
	Options          *dto.BuildOptions `json:"build_options" binding:"omitempty"`
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
		tomlPath := filepath.Join(sourcePath, dto.InfoFileName)
		content, err := utils.ReadTomlFile(tomlPath)
		if err != nil {
			return err
		}

		if name, ok := content[dto.InfoNameField].(string); ok && name != "" {
			req.ImageName = name
		} else {
			return fmt.Errorf("%s does not contain a valid name field", dto.InfoFileName)
		}
	}
	return nil
}

type UpdateContainerVersionReq struct {
	GithubLink        *string              `json:"github_link" binding:"omitempty"`
	Command           *string              `json:"command" binding:"omitempty"`
	Status            *consts.StatusType   `json:"status" binding:"omitempty"`
	HelmConfigRequest *UpdateHelmConfigReq `json:"helm_config" binding:"omitempty"`
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
		if err := validateStatus(req.Status, true); err != nil {
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

func (req *UpdateContainerVersionReq) PatchContainerVersionModel(target *model.ContainerVersion) {
	if req.GithubLink != nil {
		target.GithubLink = *req.GithubLink
	}
	if req.Command != nil {
		target.Command = *req.Command
	}
	if req.Status != nil {
		target.Status = *req.Status
	}
}

type UpdateHelmConfigReq struct {
	RepoURL       *string         `json:"repo_url" binding:"omitempty"`
	RepoName      *string         `json:"repo_name" binding:"omitempty"`
	ChartName     *string         `json:"chart_name" binding:"omitempty"`
	DynamicValues *map[string]any `json:"dynamic_values" binding:"omitempty" swaggertype:"object"`
}

func (req *UpdateHelmConfigReq) Validate() error {
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
	return nil
}

func (req *UpdateHelmConfigReq) PatchHelmConfigModel(target *model.HelmConfig) error {
	if req.RepoURL != nil {
		target.RepoURL = *req.RepoURL
	}
	if req.RepoName != nil {
		target.RepoName = *req.RepoName
	}
	if req.ChartName != nil {
		target.ChartName = *req.ChartName
	}
	return nil
}

type ContainerVersionResp struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	ImageRef  string    `json:"image_ref"`
	Usage     int       `json:"usage"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SetContainerVersionImageReq is the request body for
// PATCH /api/v2/container-versions/:id/image. It rewrites the four image
// reference columns on a container_versions row.
type SetContainerVersionImageReq struct {
	Registry   string `json:"registry"`
	Namespace  string `json:"namespace"`
	Repository string `json:"repository" binding:"required"`
	Tag        string `json:"tag" binding:"required"`
}

func (req *SetContainerVersionImageReq) Validate() error {
	req.Registry = strings.TrimSpace(req.Registry)
	req.Namespace = strings.TrimSpace(req.Namespace)
	req.Repository = strings.TrimSpace(req.Repository)
	req.Tag = strings.TrimSpace(req.Tag)
	if req.Registry == "" {
		req.Registry = "docker.io"
	}
	if req.Repository == "" {
		return fmt.Errorf("repository is required")
	}
	if req.Tag == "" {
		return fmt.Errorf("tag is required")
	}
	return nil
}

// SetContainerVersionImageResp is returned after a successful image rewrite.
type SetContainerVersionImageResp struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Registry   string `json:"registry"`
	Namespace  string `json:"namespace"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	ImageRef   string `json:"image_ref"`
}

func NewSetContainerVersionImageResp(version *model.ContainerVersion) *SetContainerVersionImageResp {
	return &SetContainerVersionImageResp{
		ID:         version.ID,
		Name:       version.Name,
		Registry:   version.Registry,
		Namespace:  version.Namespace,
		Repository: version.Repository,
		Tag:        version.Tag,
		ImageRef:   version.ImageRef,
	}
}

func NewContainerVersionResp(version *model.ContainerVersion) *ContainerVersionResp {
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
	GithubLink string                `json:"github_link"`
	Command    string                `json:"command"`
	EnvVars    string                `json:"env_vars"`
	HelmConfig *HelmConfigDetailResp `json:"helm_config,omitempty"`
}

func NewContainerVersionDetailResp(version *model.ContainerVersion) *ContainerVersionDetailResp {
	return &ContainerVersionDetailResp{
		ContainerVersionResp: *NewContainerVersionResp(version),
		GithubLink:           version.GithubLink,
		Command:              version.Command,
	}
}

type ListContainerVersionResp struct {
	Items      []ContainerVersionResp `json:"items"`
	Pagination dto.PaginationInfo     `json:"pagination"`
}

type HelmConfigDetailResp struct {
	ID        int            `json:"id"`
	Version   string         `json:"version"`
	ChartName string         `json:"chart_name"`
	RepoName  string         `json:"repo_name"`
	RepoURL   string         `json:"repo_url"`
	LocalPath string         `json:"local_path,omitempty"`
	ValueFile string         `json:"value_file,omitempty"`
	Values    map[string]any `json:"values"`
}

func NewHelmConfigDetailResp(cfg *model.HelmConfig) (*HelmConfigDetailResp, error) {
	return &HelmConfigDetailResp{
		ID:        cfg.ID,
		Version:   cfg.Version,
		ChartName: cfg.ChartName,
		RepoName:  cfg.RepoName,
		RepoURL:   cfg.RepoURL,
		LocalPath: cfg.LocalPath,
		ValueFile: cfg.ValueFile,
	}, nil
}

type UploadHelmValueFileResp struct {
	FilePath string `json:"file_path"`
	FileName string `json:"file_name"`
}

type UploadHelmChartResp struct {
	FilePath string `json:"file_path"`
	FileName string `json:"file_name"`
	Checksum string `json:"checksum"`
}

type SubmitContainerBuildResp struct {
	GroupID string `json:"group_id"`
	TraceID string `json:"trace_id"`
	TaskID  string `json:"task_id"`
}

func validateStatus(statusPtr *consts.StatusType, isMutation bool) error {
	if statusPtr == nil {
		return nil
	}
	status := *statusPtr
	if _, exists := consts.ValidStatuses[status]; !exists {
		return fmt.Errorf("invalid status value: %d", status)
	}
	if isMutation && status == consts.CommonDeleted {
		return fmt.Errorf("status value cannot be set to deleted (%d) directly through this update/create operation", consts.CommonDeleted)
	}
	return nil
}

func validateLabelItems(items []dto.LabelItem) error {
	for i, label := range items {
		if strings.TrimSpace(label.Key) == "" {
			return fmt.Errorf("empty label key at index %d in add_labels", i)
		}
		if strings.TrimSpace(label.Value) == "" {
			return fmt.Errorf("empty label value at index %d in add_labels", i)
		}
	}
	return nil
}

func validateContainerType(containerType *consts.ContainerType) error {
	if containerType != nil {
		if _, exists := consts.ValidContainerTypes[*containerType]; !exists {
			return fmt.Errorf("invalid container type: %d", *containerType)
		}
	}
	return nil
}
