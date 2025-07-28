# DtoDatasetV2Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**checksum** | **str** | 文件校验和 | [optional] 
**created_at** | **str** | 创建时间 | [optional] 
**data_source** | **str** | 数据来源描述 | [optional] 
**description** | **str** | 数据集描述 | [optional] 
**download_url** | **str** | 下载链接 | [optional] 
**file_count** | **int** | 文件数量 | [optional] 
**format** | **str** | 数据格式 | [optional] 
**id** | **int** | 唯一标识 | [optional] 
**injections** | [**List[DtoDatasetV2InjectionRelationResponse]**](DtoDatasetV2InjectionRelationResponse.md) | 关联的故障注入 | [optional] 
**is_public** | **bool** | 是否公开 | [optional] 
**labels** | [**List[DatabaseLabel]**](DatabaseLabel.md) | 关联的标签 | [optional] 
**name** | **str** | 数据集名称 | [optional] 
**project** | [**DatabaseProject**](DatabaseProject.md) | 关联项目信息 | [optional] 
**project_id** | **int** | 项目ID | [optional] 
**status** | **int** | 状态 | [optional] 
**type** | **str** | 数据集类型 | [optional] 
**updated_at** | **str** | 更新时间 | [optional] 
**version** | **str** | 数据集版本 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_v2_response import DtoDatasetV2Response

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetV2Response from a JSON string
dto_dataset_v2_response_instance = DtoDatasetV2Response.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetV2Response.to_json())

# convert the object into a dict
dto_dataset_v2_response_dict = dto_dataset_v2_response_instance.to_dict()
# create an instance of DtoDatasetV2Response from a dict
dto_dataset_v2_response_from_dict = DtoDatasetV2Response.from_dict(dto_dataset_v2_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


