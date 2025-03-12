package dto

import (
	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/executor"
)

type DatasetDeleteReq struct {
	IDs []int `form:"ids" binding:"required"`
}

type DatasetDownloadReq struct {
	GroupIDs []string `form:"group_ids" binding:"required"`
}

type DatasetItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type DatasetListReq struct {
	PaginationReq
}

func ConvertToDatasetItem(f *database.FaultInjectionSchedule) *DatasetItem {
	return &DatasetItem{
		ID:   f.ID,
		Name: f.InjectionName,
	}
}

var DatasetStatusMap = map[int]string{
	executor.DatasetInitial: "initial",
	executor.DatasetSuccess: "success",
	executor.DatasetFailed:  "failed",
	executor.DatesetDeleted: "deleted",
}
