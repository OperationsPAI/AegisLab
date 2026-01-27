package initialization

import (
	"aegis/consts"
	"aegis/database"
)

const AdminUsername = "admin"

type InitialDynamicConfig struct {
	Key          string                 `json:"key"`
	DefaultValue string                 `json:"default_value"`
	ValueType    consts.ConfigValueType `json:"value_type"`
	Scope        consts.ConfigScope     `json:"scope"`
	Category     string                 `json:"category"`
	Description  string                 `json:"description"`
	IsSecret     bool                   `json:"is_secret"`
	MinValue     *float64               `json:"min_value,omitempty"`
	MaxValue     *float64               `json:"max_value,omitempty"`
	Pattern      string                 `json:"pattern,omitempty"`
	Options      string                 `json:"options,omitempty"`
}

func (c *InitialDynamicConfig) ConvertToDBDynamicConfig() *database.DynamicConfig {
	return &database.DynamicConfig{
		Key:          c.Key,
		DefaultValue: c.DefaultValue,
		ValueType:    c.ValueType,
		Scope:        c.Scope,
		Category:     c.Category,
		Description:  c.Description,
		IsSecret:     c.IsSecret,
		MinValue:     c.MinValue,
		MaxValue:     c.MaxValue,
		Pattern:      c.Pattern,
		Options:      c.Options,
	}
}

type InitialDataContainer struct {
	Type     consts.ContainerType      `json:"type"`
	Name     string                    `json:"name"`
	IsPublic bool                      `json:"is_public"`
	Status   consts.StatusType         `json:"status"`
	Versions []InitialContainerVersion `json:"versions"`
}

func (c *InitialDataContainer) ConvertToDBContainer() *database.Container {
	return &database.Container{
		Type:     c.Type,
		Name:     c.Name,
		IsPublic: c.IsPublic,
		Status:   c.Status,
	}
}

type InitialContainerVersion struct {
	Name       string                   `json:"name"`
	GithubLink string                   `json:"github_link"`
	ImageRef   string                   `json:"image_ref"`
	Command    string                   `json:"command"`
	EnvVars    []InitialParameterConfig `json:"env_vars"`
	Status     consts.StatusType        `json:"status"`
	HelmConfig *InitialHelmConfig       `json:"helm_config"`
}

func (cv *InitialContainerVersion) ConvertToDBContainerVersion() *database.ContainerVersion {
	return &database.ContainerVersion{
		Name:       cv.Name,
		GithubLink: cv.GithubLink,
		ImageRef:   cv.ImageRef,
		Command:    cv.Command,
		Status:     cv.Status,
	}
}

type InitialHelmConfig struct {
	Version   string                   `json:"version"`
	ChartName string                   `json:"chart_name"`
	RepoName  string                   `json:"repo_name"`
	RepoURL   string                   `json:"repo_url"`
	Values    []InitialParameterConfig `json:"values"`
}

func (hc *InitialHelmConfig) ConvertToDBHelmConfig() *database.HelmConfig {
	return &database.HelmConfig{
		Version:   hc.Version,
		ChartName: hc.ChartName,
		RepoName:  hc.RepoName,
		RepoURL:   hc.RepoURL,
	}
}

type InitialParameterConfig struct {
	Key            string                   `json:"key"`
	Type           consts.ParameterType     `json:"type"`
	Category       consts.ParameterCategory `json:"category"`
	ValueType      consts.ValueDataType     `json:"value_type"`
	DefaultValue   *string                  `json:"default_value"`
	TemplateString *string                  `json:"template_string"`
	Required       bool                     `json:"required"`
	Overridable    *bool                    `json:"overridable"`
}

func (pc *InitialParameterConfig) ConvertToDBParameterConfig() *database.ParameterConfig {
	config := &database.ParameterConfig{
		Key:            pc.Key,
		Type:           pc.Type,
		Category:       pc.Category,
		ValueType:      pc.ValueType,
		DefaultValue:   pc.DefaultValue,
		TemplateString: pc.TemplateString,
		Required:       pc.Required,
		Overridable:    true,
	}

	if pc.Overridable != nil {
		config.Overridable = *pc.Overridable
	}

	return config
}

type InitialDatasaet struct {
	Name        string                  `json:"name"`
	Type        string                  `json:"type"`
	Description string                  `json:"description"`
	IsPublic    bool                    `json:"is_public"`
	Status      consts.StatusType       `json:"status"`
	Versions    []InitialDatasetVersion `json:"versions"`
}

func (d *InitialDatasaet) ConvertToDBDataset() *database.Dataset {
	return &database.Dataset{
		Name:        d.Name,
		Type:        d.Type,
		Description: d.Description,
		IsPublic:    d.IsPublic,
		Status:      d.Status,
	}
}

type InitialDatasetVersion struct {
	Name   string            `json:"name"`
	Status consts.StatusType `json:"status"`
}

func (dv *InitialDatasetVersion) ConvertToDBDatasetVersion() *database.DatasetVersion {
	return &database.DatasetVersion{
		Name:   dv.Name,
		Status: dv.Status,
	}
}

type InitialDataProject struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Status      consts.StatusType `json:"status"`
}

func (p *InitialDataProject) ConvertToDBProject() *database.Project {
	return &database.Project{
		Name:        p.Name,
		Description: p.Description,
		Status:      p.Status,
	}
}

type InitialUserProject struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type InitialDataUser struct {
	Username string                `json:"username"`
	Email    string                `json:"email"`
	Password string                `json:"password"`
	FullName string                `json:"full_name"`
	Status   consts.StatusType     `json:"status"`
	IsActive bool                  `json:"is_active"`
	Projects []InitialUserProject `json:"project,omitempty"`
}

func (u *InitialDataUser) ConvertToDBUser() *database.User {
	return &database.User{
		Username: u.Username,
		Email:    u.Email,
		Password: u.Password,
		FullName: u.FullName,
		Status:   u.Status,
		IsActive: u.IsActive,
	}
}

type InitialData struct {
	DynamicConfigs []InitialDynamicConfig `json:"dynamic_configs"`
	Containers     []InitialDataContainer `json:"containers"`
	Datasets       []InitialDatasaet      `json:"datasets"`
	Projects       []InitialDataProject   `json:"projects"`
	AdminUser      InitialDataUser        `json:"admin_user"`
	Users          []InitialDataUser      `json:"users"`
}

type ConsumerData struct {
	configs []database.DynamicConfig
}
