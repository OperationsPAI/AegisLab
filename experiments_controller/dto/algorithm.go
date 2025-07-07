package dto

import (
	"fmt"
	"strings"
	"time"

	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/utils"
)

type AlgorithmItem struct {
	Algorithm string    `json:"algorithm"`
	Image     string    `json:"image"`
	Tag       string    `json:"tag"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a *AlgorithmItem) Convert(container database.Container) {
	a.Algorithm = container.Name
	a.Image = container.Image
	a.Tag = container.Tag
	a.UpdatedAt = container.UpdatedAt
}

type ListAlgorithmsResp []AlgorithmItem

type SubmitExecutionReq []ExecutionPayload

func (req *SubmitExecutionReq) Validate() error {
	for _, payload := range *req {
		if err := payload.Validate(); err != nil {
			return fmt.Errorf("Invalid execution payload: %v", err)
		}
	}

	return nil
}

type ExecutionPayload struct {
	Algorithm string            `json:"algorithm" binding:"required"`
	Dataset   string            `json:"dataset" binding:"required"`
	EnvVars   map[string]string `json:"env_vars" binding:"omitempty" swaggertype:"object"`
}

var ExecuteEnvVarNameMap = map[string]struct{}{
	"ALGORITHM": {},
	"SERVICE":   {},
	"VENV":      {},
}

func (p *ExecutionPayload) Validate() error {
	if len(p.EnvVars) != 0 {
		for key := range p.EnvVars {
			if _, exists := ExecuteEnvVarNameMap[key]; !exists {
				return fmt.Errorf("Invalid environment variable name: %s", key)
			}
		}
	}

	return nil
}

type BuildSource struct {
	Type   consts.BuildSourceType `json:"type" binding:"required"`
	File   *FileSource            `json:"file" binding:"omitempty"`
	GitHub *GitHubSource          `json:"github" binding:"omitempty"`
}

func (s *BuildSource) Validate() error {
	if s.Type == consts.BuildSourceTypeGitHub {
		return s.GitHub.Validate()
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

type SubmitBuildingReq struct {
	Algorithm    string        `json:"algorithm" binding:"required"`
	Image        string        `json:"image" binding:"required"`
	Tag          string        `json:"tag" binding:"omitempty"`
	Source       BuildSource   `json:"source" binding:"required"`
	BuildOptions *BuildOptions `json:"build_options" binding:"omitempty"`
}

func (req *SubmitBuildingReq) Validate() error {
	if req.Algorithm == "" {
		return fmt.Errorf("Algorithm name cannot be empty")
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

	if err := req.Source.Validate(); err != nil {
		return fmt.Errorf("Invalid build source: %v", err)
	}

	if err := req.BuildOptions.Validate(); err != nil {
		return fmt.Errorf("Invalid build options: %v", err)
	}

	return nil
}

type DetectorRecord struct {
	SpanName            string   `json:"span_name"`
	Issues              string   `json:"issue"`
	AbnormalAvgDuration *float64 `json:"abnormal_avg_duration" swaggertype:"number" example:"0.5"`
	NormalAvgDuration   *float64 `json:"normal_avg_duration" swaggertype:"number" example:"0.3"`
	AbnormalSuccRate    *float64 `json:"abnormal_succ_rate" swaggertype:"number" example:"0.8"`
	NormalSuccRate      *float64 `json:"normal_succ_rate" swaggertype:"number" example:"0.95"`
	AbnormalP90         *float64 `json:"abnormal_p90" swaggertype:"number" example:"1.2"`
	NormalP90           *float64 `json:"normal_p90" swaggertype:"number" example:"0.8"`
	AbnormalP95         *float64 `json:"abnormal_p95" swaggertype:"number" example:"1.5"`
	NormalP95           *float64 `json:"normal_p95" swaggertype:"number" example:"1.0"`
	AbnormalP99         *float64 `json:"abnormal_p99" swaggertype:"number" example:"2.0"`
	NormalP99           *float64 `json:"normal_p99" swaggertype:"number" example:"1.3"`
}

type ExecutionRecord struct {
	Algorithm          string              `json:"algorithm"`
	GranularityRecords []GranularityRecord `json:"granularity_records"`
}

type ExecutionRecordWithDatasetID struct {
	DatasetID int
	ExecutionRecord
}

type GranularityRecord struct {
	Level      string  `json:"level"`
	Result     string  `json:"result"`
	Rank       int     `json:"rank"`
	Confidence float64 `json:"confidence"`
}

func (g *GranularityRecord) Convert(result database.GranularityResult) {
	g.Level = result.Level
	g.Result = result.Result
	g.Rank = result.Rank
	g.Confidence = result.Confidence
}
