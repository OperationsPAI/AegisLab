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
