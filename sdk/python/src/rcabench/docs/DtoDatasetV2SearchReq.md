# DtoDatasetV2SearchReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**date_range** | [**DtoDateRangeFilter**](DtoDateRangeFilter.md) | 时间范围过滤 | [optional] 
**include** | **List[str]** | 包含的关联数据 | [optional] 
**is_public** | **bool** | 是否公开 | [optional] 
**label_keys** | **List[str]** | 按标签键过滤 | [optional] 
**label_values** | **List[str]** | 按标签值过滤 | [optional] 
**page** | **int** | 页码 | [optional] 
**project_ids** | **List[int]** | 项目ID列表 | [optional] 
**search** | **str** | 搜索关键词 | [optional] 
**size** | **int** | 每页大小 | [optional] 
**size_range** | [**DtoSizeRangeFilter**](DtoSizeRangeFilter.md) | 大小范围过滤 | [optional] 
**sort_by** | **str** | 排序字段 | [optional] 
**sort_order** | **str** | 排序方向 | [optional] 
**statuses** | **List[int]** | 状态列表 | [optional] 
**types** | **List[str]** | 数据集类型列表 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_v2_search_req import DtoDatasetV2SearchReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetV2SearchReq from a JSON string
dto_dataset_v2_search_req_instance = DtoDatasetV2SearchReq.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetV2SearchReq.to_json())

# convert the object into a dict
dto_dataset_v2_search_req_dict = dto_dataset_v2_search_req_instance.to_dict()
# create an instance of DtoDatasetV2SearchReq from a dict
dto_dataset_v2_search_req_from_dict = DtoDatasetV2SearchReq.from_dict(dto_dataset_v2_search_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


