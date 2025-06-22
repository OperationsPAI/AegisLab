# DtoGenericResponseDtoPaginationRespDtoTaskItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | 状态码 | [optional] 
**data** | [**DtoPaginationRespDtoTaskItem**](DtoPaginationRespDtoTaskItem.md) | 泛型类型的数据 | [optional] 
**message** | **str** | 响应消息 | [optional] 
**timestamp** | **int** | 响应生成时间 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_pagination_resp_dto_task_item import DtoGenericResponseDtoPaginationRespDtoTaskItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoPaginationRespDtoTaskItem from a JSON string
dto_generic_response_dto_pagination_resp_dto_task_item_instance = DtoGenericResponseDtoPaginationRespDtoTaskItem.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoPaginationRespDtoTaskItem.to_json())

# convert the object into a dict
dto_generic_response_dto_pagination_resp_dto_task_item_dict = dto_generic_response_dto_pagination_resp_dto_task_item_instance.to_dict()
# create an instance of DtoGenericResponseDtoPaginationRespDtoTaskItem from a dict
dto_generic_response_dto_pagination_resp_dto_task_item_from_dict = DtoGenericResponseDtoPaginationRespDtoTaskItem.from_dict(dto_generic_response_dto_pagination_resp_dto_task_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


