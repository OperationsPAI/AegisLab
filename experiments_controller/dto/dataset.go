package dto

import (
	"encoding/json"
	"fmt"
	"time"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/utils"
)

type DatasetDeleteReq struct {
	Names []string `form:"names" binding:"required,min=1,dive,required,max=64"`
}

type DatasetDeleteResp struct {
	SuccessCount int64    `json:"success_count"`
	FailedNames  []string `json:"failed_names"`
}

type DatasetDownloadReq struct {
	GroupIDs []string `form:"group_ids"`
	Names    []string `form:"names"`
}

func (r *DatasetDownloadReq) Validate() error {
	hasGroupIDs := len(r.GroupIDs) > 0
	hasNames := len(r.Names) > 0
	if !hasGroupIDs && !hasNames {
		return fmt.Errorf("One of group_ids or names must be provided")
	}

	if hasGroupIDs && hasNames {
		return fmt.Errorf("Only one of group_ids or names must be provided")
	}

	return nil
}

type DatasetItem struct {
	Name      string         `json:"name"`
	Param     map[string]any `json:"param" swaggertype:"array,object"`
	StartTime time.Time      `json:"start_time"`
	EndTime   time.Time      `json:"end_time"`
}

func (d *DatasetItem) Convert(record database.FaultInjectionSchedule) error {
	var param map[string]any
	if err := json.Unmarshal([]byte(record.DisplayConfig), &param); err != nil {
		return fmt.Errorf("faild to unmarshal display config: %v", err)
	}

	param["fault_type"] = chaos.ChaosTypeMap[chaos.ChaosType(record.FaultType)]
	param["pre_duration"] = record.PreDuration

	d.Name = record.InjectionName
	d.Param = param
	d.StartTime = record.StartTime
	d.EndTime = record.EndTime

	return nil
}

type DatasetItemWithID struct {
	ID int
	DatasetItem
}

func (d *DatasetItemWithID) Convert(record database.FaultInjectionSchedule) error {
	var item DatasetItem
	err := item.Convert(record)
	if err != nil {
		return err
	}

	d.ID = record.ID
	d.DatasetItem = item
	return nil
}

type DatasetBuildPayload struct {
	Benchmark   string            `json:"benchmark" binding:"omitempty"`
	Name        string            `json:"name" binding:"required"`
	PreDuration *int              `json:"pre_duration" binding:"omitempty"`
	EnvVars     map[string]string `json:"env_vars" binding:"omitempty" swaggertype:"object"`
}

func (p *DatasetBuildPayload) Validate() error {
	if p.Benchmark == "" {
		p.Benchmark = "clickhouse"
	}

	if p.PreDuration != nil && *p.PreDuration <= 0 {
		return fmt.Errorf("pre_duration must be greater than 0")
	}

	for key := range p.EnvVars {
		if err := utils.IsValidEnvVar(key); err != nil {
			return fmt.Errorf("invalid environment variable key %s: %v", key, err)
		}
	}

	return nil
}

type SubmitDatasetBuildingReq []DatasetBuildPayload

func (req *SubmitDatasetBuildingReq) Validate() error {
	if len(*req) == 0 {
		return fmt.Errorf("at least one dataset build payload is required")
	}

	for _, payload := range *req {
		if err := payload.Validate(); err != nil {
			return fmt.Errorf("invalid dataset build payload: %v", err)
		}
	}

	return nil
}

type DatasetJoinedResult struct {
	GroupID string
	Name    string
}

func (d *DatasetJoinedResult) Convert(groupID, name string) {
	d.GroupID = groupID
	d.Name = name
}

var DatasetStatusMap = map[int]string{
	consts.DatasetInitial:       "initial",
	consts.DatasetInjectSuccess: "inject_success",
	consts.DatasetInjectFailed:  "inject_failed",
	consts.DatasetBuildSuccess:  "build_success",
	consts.DatasetBuildFailed:   "build_failed",
	consts.DatasetDeleted:       "deleted",
}

var DatasetStatusReverseMap = map[string]int{
	"initial":        consts.DatasetInitial,
	"inject_success": consts.DatasetInjectSuccess,
	"inject_failed":  consts.DatasetInjectFailed,
	"build_success":  consts.DatasetBuildSuccess,
	"build_failed":   consts.DatasetBuildFailed,
	"deleted":        consts.DatasetDeleted,
}

// ===================== V2 API DTOs =====================

// DatasetV2CreateReq 创建数据集请求
type DatasetV2CreateReq struct {
	Name         string                    `json:"name" binding:"required,max=255"`     // 数据集名称
	Version      string                    `json:"version" binding:"max=50"`            // 数据集版本，可选，默认v1.0
	Description  string                    `json:"description" binding:"max=1000"`      // 数据集描述
	Type         string                    `json:"type" binding:"required,max=50"`      // 数据集类型
	DataSource   string                    `json:"data_source" binding:"max=500"`       // 数据来源描述
	Format       string                    `json:"format" binding:"max=50"`             // 数据格式
	ProjectID    int                       `json:"project_id" binding:"required,min=1"` // 项目ID
	IsPublic     *bool                     `json:"is_public"`                           // 是否公开，可选，默认false
	InjectionIDs []int                     `json:"injection_ids"`                       // 关联的故障注入ID列表
	LabelIDs     []int                     `json:"label_ids"`                           // 关联的标签ID列表
	NewLabels    []DatasetV2LabelCreateReq `json:"new_labels"`                          // 新建标签列表
}

// DatasetV2LabelCreateReq 创建标签请求
type DatasetV2LabelCreateReq struct {
	Key         string `json:"key" binding:"required,max=100"`   // 标签键
	Value       string `json:"value" binding:"required,max=255"` // 标签值
	Category    string `json:"category" binding:"max=50"`        // 标签分类
	Description string `json:"description" binding:"max=500"`    // 标签描述
	Color       string `json:"color" binding:"max=7"`            // 标签颜色 (hex格式)
}

// DatasetV2UpdateReq 更新数据集请求
type DatasetV2UpdateReq struct {
	Name         *string                   `json:"name" binding:"omitempty,max=255"`         // 数据集名称
	Version      *string                   `json:"version" binding:"omitempty,max=50"`       // 数据集版本
	Description  *string                   `json:"description" binding:"omitempty,max=1000"` // 数据集描述
	Type         *string                   `json:"type" binding:"omitempty,max=50"`          // 数据集类型
	DataSource   *string                   `json:"data_source" binding:"omitempty,max=500"`  // 数据来源描述
	Format       *string                   `json:"format" binding:"omitempty,max=50"`        // 数据格式
	IsPublic     *bool                     `json:"is_public"`                                // 是否公开
	InjectionIDs []int                     `json:"injection_ids"`                            // 更新关联的故障注入ID列表（完全替换）
	LabelIDs     []int                     `json:"label_ids"`                                // 更新关联的标签ID列表（完全替换）
	NewLabels    []DatasetV2LabelCreateReq `json:"new_labels"`                               // 新建标签列表
}

// DatasetV2InjectionManageReq 管理数据集中的故障注入
type DatasetV2InjectionManageReq struct {
	AddInjections    []int `json:"add_injections"`    // 要添加的故障注入ID列表
	RemoveInjections []int `json:"remove_injections"` // 要移除的故障注入ID列表
}

// DatasetV2LabelManageReq 管理数据集中的标签
type DatasetV2LabelManageReq struct {
	AddLabels    []int                     `json:"add_labels"`    // 要添加的标签ID列表
	RemoveLabels []int                     `json:"remove_labels"` // 要移除的标签ID列表
	NewLabels    []DatasetV2LabelCreateReq `json:"new_labels"`    // 新建标签列表
}

// DatasetV2Response 数据集响应
type DatasetV2Response struct {
	ID          int    `json:"id"`          // 唯一标识
	Name        string `json:"name"`        // 数据集名称
	Version     string `json:"version"`     // 数据集版本
	Description string `json:"description"` // 数据集描述
	Type        string `json:"type"`        // 数据集类型

	FileCount   int                                  `json:"file_count"`             // 文件数量
	DataSource  string                               `json:"data_source"`            // 数据来源描述
	Format      string                               `json:"format"`                 // 数据格式
	ProjectID   int                                  `json:"project_id"`             // 项目ID
	Status      int                                  `json:"status"`                 // 状态
	IsPublic    bool                                 `json:"is_public"`              // 是否公开
	DownloadURL string                               `json:"download_url,omitempty"` // 下载链接
	Checksum    string                               `json:"checksum,omitempty"`     // 文件校验和
	CreatedAt   time.Time                            `json:"created_at"`             // 创建时间
	UpdatedAt   time.Time                            `json:"updated_at"`             // 更新时间
	Project     *database.Project                    `json:"project,omitempty"`      // 关联项目信息
	Injections  []DatasetV2InjectionRelationResponse `json:"injections,omitempty"`   // 关联的故障注入
	Labels      []database.Label                     `json:"labels,omitempty"`       // 关联的标签
}

// DatasetV2InjectionRelationResponse 数据集故障注入关联响应
type DatasetV2InjectionRelationResponse struct {
	ID               int                              `json:"id"`                        // 关联ID
	FaultInjectionID int                              `json:"fault_injection_id"`        // 故障注入ID
	CreatedAt        time.Time                        `json:"created_at"`                // 创建时间
	UpdatedAt        time.Time                        `json:"updated_at"`                // 更新时间
	FaultInjection   *database.FaultInjectionSchedule `json:"fault_injection,omitempty"` // 故障注入详情
}

// DatasetV2ListReq 数据集列表查询请求
type DatasetV2ListReq struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`         // 页码，默认1
	Size      int    `form:"size" binding:"omitempty,min=1,max=100"` // 每页大小，默认20
	ProjectID *int   `form:"project_id" binding:"omitempty,min=1"`   // 项目ID过滤
	Type      string `form:"type"`                                   // 数据集类型过滤
	Status    *int   `form:"status"`                                 // 状态过滤
	IsPublic  *bool  `form:"is_public"`                              // 是否公开过滤
	Search    string `form:"search"`                                 // 搜索关键词（名称、描述）
	SortBy    string `form:"sort_by"`                                // 排序字段（id, name, created_at, updated_at）
	SortOrder string `form:"sort_order"`                             // 排序方向（asc, desc）
	Include   string `form:"include"`                                // 包含的关联数据（project, injections, labels）
}

// DatasetV2SearchReq 数据集搜索请求（POST方式，支持复杂条件）
type DatasetV2SearchReq struct {
	Page        int              `json:"page" binding:"omitempty,min=1"`         // 页码
	Size        int              `json:"size" binding:"omitempty,min=1,max=100"` // 每页大小
	ProjectIDs  []int            `json:"project_ids"`                            // 项目ID列表
	Types       []string         `json:"types"`                                  // 数据集类型列表
	Statuses    []int            `json:"statuses"`                               // 状态列表
	IsPublic    *bool            `json:"is_public"`                              // 是否公开
	Search      string           `json:"search"`                                 // 搜索关键词
	DateRange   *DateRangeFilter `json:"date_range"`                             // 时间范围过滤
	SizeRange   *SizeRangeFilter `json:"size_range"`                             // 大小范围过滤
	Include     []string         `json:"include"`                                // 包含的关联数据
	SortBy      string           `json:"sort_by"`                                // 排序字段
	SortOrder   string           `json:"sort_order"`                             // 排序方向
	LabelKeys   []string         `json:"label_keys"`                             // 按标签键过滤
	LabelValues []string         `json:"label_values"`                           // 按标签值过滤
}

// DateRangeFilter 时间范围过滤器
type DateRangeFilter struct {
	StartTime *time.Time `json:"start_time"` // 开始时间
	EndTime   *time.Time `json:"end_time"`   // 结束时间
}

// SizeRangeFilter 大小范围过滤器
type SizeRangeFilter struct {
	MinSize *int64 `json:"min_size"` // 最小大小（字节）
	MaxSize *int64 `json:"max_size"` // 最大大小（字节）
}

// DatasetSearchResponse 数据集搜索响应结构
type DatasetSearchResponse struct {
	Items      []DatasetV2Response `json:"items"`      // 结果列表
	Pagination PaginationInfo      `json:"pagination"` // 分页信息
}

// ToDatasetV2Response 将Database.Dataset转换为DatasetV2Response
func ToDatasetV2Response(dataset *database.Dataset, includeRelations bool) *DatasetV2Response {
	resp := &DatasetV2Response{
		ID:          dataset.ID,
		Name:        dataset.Name,
		Version:     dataset.Version,
		Description: dataset.Description,
		Type:        dataset.Type,
		FileCount:   dataset.FileCount,
		DataSource:  dataset.DataSource,
		Format:      dataset.Format,
		ProjectID:   dataset.ProjectID,
		Status:      dataset.Status,
		IsPublic:    dataset.IsPublic,
		DownloadURL: dataset.DownloadURL,
		Checksum:    dataset.Checksum,
		CreatedAt:   dataset.CreatedAt,
		UpdatedAt:   dataset.UpdatedAt,
	}

	// 包含项目信息
	if dataset.Project != nil {
		resp.Project = dataset.Project
	}

	return resp
}
