# DtoDatasetV2CreateReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**data_source** | **str** | 数据来源描述 | [optional] 
**description** | **str** | 数据集描述 | [optional] 
**format** | **str** | 数据格式 | [optional] 
**injection_ids** | **List[int]** | 关联的故障注入ID列表 | [optional] 
**is_public** | **bool** | 是否公开，可选，默认false | [optional] 
**label_ids** | **List[int]** | 关联的标签ID列表 | [optional] 
**name** | **str** | 数据集名称 | 
**new_labels** | [**List[DtoDatasetV2LabelCreateReq]**](DtoDatasetV2LabelCreateReq.md) | 新建标签列表 | [optional] 
**project_id** | **int** | 项目ID | 
**type** | **str** | 数据集类型 | 
**version** | **str** | 数据集版本，可选，默认v1.0 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_v2_create_req import DtoDatasetV2CreateReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetV2CreateReq from a JSON string
dto_dataset_v2_create_req_instance = DtoDatasetV2CreateReq.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetV2CreateReq.to_json())

# convert the object into a dict
dto_dataset_v2_create_req_dict = dto_dataset_v2_create_req_instance.to_dict()
# create an instance of DtoDatasetV2CreateReq from a dict
dto_dataset_v2_create_req_from_dict = DtoDatasetV2CreateReq.from_dict(dto_dataset_v2_create_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


