package dto

import (
	"encoding/json"
	"fmt"
	"time"

	chaos "github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
)

type DatasetDeleteReq struct {
	Names []string `form:"names" binding:"required,min=1,dive,required,max=64"`
}

type DatasetDeleteResp struct {
	SuccessCount int64    `json:"success_count"`
	FailedNames  []string `json:"failed_names"`
}

type DatasetDownloadReq struct {
	GroupIDs []string `form:"group_ids" binding:"required"`
}

type DatasetItem struct {
	Name      string         `json:"name"`
	Param     map[string]any `json:"param"`
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

type DatasetListReq struct {
	PaginationReq
}

type DatasetPayload struct {
	Benchmark   string     `json:"benchmark"`
	Name        string     `json:"name"`
	PreDuration int        `json:"pre_duration"`
	Service     string     `json:"service"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
}

type DatasetJoinedResult struct {
	GroupID string
	Name    string
}

func (d *DatasetJoinedResult) Convert(groupID, name string) {
	d.GroupID = groupID
	d.Name = name
}

type QueryDatasetReq struct {
	Name string `form:"name" binding:"required,max=64"`
	Sort string `form:"sort" binding:"oneof=desc asc"`
}

type QueryDatasetResp struct {
	DatasetItem
	DetectorResult   DetectorRecord    `json:"detector_result"`
	ExecutionResults []ExecutionRecord `json:"execution_results"`
}

var DatasetStatusMap = map[int]string{
	consts.DatasetInitial:       "initial",
	consts.DatasetInjectSuccess: "inject_success",
	consts.DatasetInjectFailed:  "inject_failed",
	consts.DatasetBuildSuccess:  "build_success",
	consts.DatasetBuildFailed:   "build_failed",
	consts.DatasetDeleted:       "deleted",
}
