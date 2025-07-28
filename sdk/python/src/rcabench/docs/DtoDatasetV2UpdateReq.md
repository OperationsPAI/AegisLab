# DtoDatasetV2UpdateReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**data_source** | **str** | 数据来源描述 | [optional] 
**description** | **str** | 数据集描述 | [optional] 
**format** | **str** | 数据格式 | [optional] 
**injection_ids** | **List[int]** | 更新关联的故障注入ID列表（完全替换） | [optional] 
**is_public** | **bool** | 是否公开 | [optional] 
**label_ids** | **List[int]** | 更新关联的标签ID列表（完全替换） | [optional] 
**name** | **str** | 数据集名称 | [optional] 
**new_labels** | [**List[DtoDatasetV2LabelCreateReq]**](DtoDatasetV2LabelCreateReq.md) | 新建标签列表 | [optional] 
**type** | **str** | 数据集类型 | [optional] 
**version** | **str** | 数据集版本 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_v2_update_req import DtoDatasetV2UpdateReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetV2UpdateReq from a JSON string
dto_dataset_v2_update_req_instance = DtoDatasetV2UpdateReq.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetV2UpdateReq.to_json())

# convert the object into a dict
dto_dataset_v2_update_req_dict = dto_dataset_v2_update_req_instance.to_dict()
# create an instance of DtoDatasetV2UpdateReq from a dict
dto_dataset_v2_update_req_from_dict = DtoDatasetV2UpdateReq.from_dict(dto_dataset_v2_update_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


