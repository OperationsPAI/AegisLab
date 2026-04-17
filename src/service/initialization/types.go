package initialization

import (
	"aegis/consts"
	"aegis/model"
	"gorm.io/gorm"
)

const AdminUsername = "admin"

type InitialDynamicConfig struct {
	Key          string                 `yaml:"key"`
	DefaultValue string                 `yaml:"default_value"`
	ValueType    consts.ConfigValueType `yaml:"value_type"`
	Scope        consts.ConfigScope     `yaml:"scope"`
	Category     string                 `yaml:"category"`
	Description  string                 `yaml:"description"`
	IsSecret     bool                   `yaml:"is_secret"`
	MinValue     *float64               `yaml:"min_value"`
	MaxValue     *float64               `yaml:"max_value"`
	Pattern      string                 `yaml:"pattern"`
	Options      string                 `yaml:"options"`
}

func (c *InitialDynamicConfig) ConvertToDBDynamicConfig() *model.DynamicConfig {
	return &model.DynamicConfig{
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
	Type     consts.ContainerType      `yaml:"type"`
	Name     string                    `yaml:"name"`
	IsPublic bool                      `yaml:"is_public"`
	Status   consts.StatusType         `yaml:"status"`
	Versions []InitialContainerVersion `yaml:"versions"`
}

func (c *InitialDataContainer) ConvertToDBContainer() *model.Container {
	return &model.Container{
		Type:     c.Type,
		Name:     c.Name,
		IsPublic: c.IsPublic,
		Status:   c.Status,
	}
}

type InitialContainerVersion struct {
	Name       string                   `yaml:"name"`
	GithubLink string                   `yaml:"github_link"`
	ImageRef   string                   `yaml:"image_ref"`
	Command    string                   `yaml:"command"`
	EnvVars    []InitialParameterConfig `yaml:"env_vars"`
	Status     consts.StatusType        `yaml:"status"`
	HelmConfig *InitialHelmConfig       `yaml:"helm_config"`
}

func (cv *InitialContainerVersion) ConvertToDBContainerVersion() *model.ContainerVersion {
	return &model.ContainerVersion{
		Name:       cv.Name,
		GithubLink: cv.GithubLink,
		ImageRef:   cv.ImageRef,
		Command:    cv.Command,
		Status:     cv.Status,
	}
}

type InitialHelmConfig struct {
	Version   string                   `yaml:"version"`
	ChartName string                   `yaml:"chart_name"`
	RepoName  string                   `yaml:"repo_name"`
	RepoURL   string                   `yaml:"repo_url"`
	Values    []InitialParameterConfig `yaml:"values"`
}

func (hc *InitialHelmConfig) ConvertToDBHelmConfig() *model.HelmConfig {
	return &model.HelmConfig{
		Version:   hc.Version,
		ChartName: hc.ChartName,
		RepoName:  hc.RepoName,
		RepoURL:   hc.RepoURL,
	}
}

type InitialParameterConfig struct {
	Key            string                   `yaml:"key"`
	Type           consts.ParameterType     `yaml:"type"`
	Category       consts.ParameterCategory `yaml:"category"`
	ValueType      consts.ValueDataType     `yaml:"value_type"`
	DefaultValue   *string                  `yaml:"default_value"`
	TemplateString *string                  `yaml:"template_string"`
	Required       bool                     `yaml:"required"`
	Overridable    *bool                    `yaml:"overridable"`
}

func (pc *InitialParameterConfig) ConvertToDBParameterConfig() *model.ParameterConfig {
	config := &model.ParameterConfig{
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
	Name        string                  `yaml:"name"`
	Type        string                  `yaml:"type"`
	Description string                  `yaml:"description"`
	IsPublic    bool                    `yaml:"is_public"`
	Status      consts.StatusType       `yaml:"status"`
	Versions    []InitialDatasetVersion `yaml:"versions"`
}

func (d *InitialDatasaet) ConvertToDBDataset() *model.Dataset {
	return &model.Dataset{
		Name:        d.Name,
		Type:        d.Type,
		Description: d.Description,
		IsPublic:    d.IsPublic,
		Status:      d.Status,
	}
}

type InitialDatasetVersion struct {
	Name   string            `yaml:"name"`
	Status consts.StatusType `yaml:"status"`
}

func (dv *InitialDatasetVersion) ConvertToDBDatasetVersion() *model.DatasetVersion {
	return &model.DatasetVersion{
		Name:   dv.Name,
		Status: dv.Status,
	}
}

type InitialDataProject struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Status      consts.StatusType `yaml:"status"`
}

func (p *InitialDataProject) ConvertToDBProject() *model.Project {
	return &model.Project{
		Name:        p.Name,
		Description: p.Description,
		Status:      p.Status,
	}
}

type InitialDataTeam struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	IsPublic    bool              `yaml:"is_public"`
	Status      consts.StatusType `yaml:"status"`
}

func (t *InitialDataTeam) ConvertToDBTeam() *model.Team {
	return &model.Team{
		Name:        t.Name,
		Description: t.Description,
		IsPublic:    t.IsPublic,
		Status:      t.Status,
	}
}

type InitialUserProject struct {
	Name string `yaml:"name"`
	Role string `yaml:"role"`
}

type InitialUserTeam struct {
	Name     string               `yaml:"name"`
	Role     string               `yaml:"role"`
	Projects []InitialUserProject `yaml:"projects"`
}

type InitialDataUser struct {
	Username string               `yaml:"username"`
	Email    string               `yaml:"email"`
	Password string               `yaml:"password"`
	FullName string               `yaml:"full_name"`
	Status   consts.StatusType    `yaml:"status"`
	IsActive bool                 `yaml:"is_active"`
	Projects []InitialUserProject `yaml:"projects"`
	Teams    []InitialUserTeam    `yaml:"teams"`
}

func (u *InitialDataUser) ConvertToDBUser() *model.User {
	return &model.User{
		Username: u.Username,
		Email:    u.Email,
		Password: u.Password,
		FullName: u.FullName,
		Status:   u.Status,
		IsActive: u.IsActive,
	}
}

type InitialData struct {
	DynamicConfigs []InitialDynamicConfig `yaml:"dynamic_configs"`
	Containers     []InitialDataContainer `yaml:"containers"`
	Datasets       []InitialDatasaet      `yaml:"datasets"`
	Projects       []InitialDataProject   `yaml:"projects"`
	Teams          []InitialDataTeam      `yaml:"teams"`
	AdminUser      InitialDataUser        `yaml:"admin_user"`
	Users          []InitialDataUser      `yaml:"users"`
}

type configData struct {
	scope   consts.ConfigScope
	configs []model.DynamicConfig
}

func newConfigDataWithDB(db *gorm.DB, scope consts.ConfigScope) (*configData, error) {
	configs, err := newBootstrapStore(db).listExistingConfigs()
	if err != nil {
		return nil, err
	}

	return &configData{
		scope:   scope,
		configs: configs,
	}, nil
}
