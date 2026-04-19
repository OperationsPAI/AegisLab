package pedestal

// ---------------------- Pedestal Helm DTOs ------------------

// PedestalHelmConfigResp represents a full helm_configs row for CLI/API consumers.
type PedestalHelmConfigResp struct {
	ID                 int    `json:"id"`
	ContainerVersionID int    `json:"container_version_id"`
	ChartName          string `json:"chart_name"`
	Version            string `json:"version"`
	RepoURL            string `json:"repo_url"`
	RepoName           string `json:"repo_name"`
	ValueFile          string `json:"value_file"`
	LocalPath          string `json:"local_path"`
	Checksum           string `json:"checksum"`
}

// UpsertPedestalHelmConfigReq is the body for PUT /api/v2/pedestal/helm/:container_version_id
type UpsertPedestalHelmConfigReq struct {
	ChartName string `json:"chart_name" binding:"required"`
	Version   string `json:"version" binding:"required"`
	RepoURL   string `json:"repo_url" binding:"required"`
	RepoName  string `json:"repo_name" binding:"required"`
	ValueFile string `json:"value_file"`
	LocalPath string `json:"local_path"`
}

// PedestalHelmVerifyCheck is a single step in the verify pipeline.
type PedestalHelmVerifyCheck struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

// PedestalHelmVerifyResp is the aggregated verify response.
type PedestalHelmVerifyResp struct {
	OK     bool                      `json:"ok"`
	Checks []PedestalHelmVerifyCheck `json:"checks"`
}
