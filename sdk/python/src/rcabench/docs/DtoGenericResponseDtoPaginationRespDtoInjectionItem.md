# DtoGenericResponseDtoPaginationRespDtoInjectionItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | 状态码 | [optional] 
**data** | [**DtoPaginationRespDtoInjectionItem**](DtoPaginationRespDtoInjectionItem.md) | 泛型类型的数据 | [optional] 
**message** | **str** | 响应消息 | [optional] 
**timestamp** | **int** | 响应生成时间 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_pagination_resp_dto_injection_item import DtoGenericResponseDtoPaginationRespDtoInjectionItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoPaginationRespDtoInjectionItem from a JSON string
dto_generic_response_dto_pagination_resp_dto_injection_item_instance = DtoGenericResponseDtoPaginationRespDtoInjectionItem.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoPaginationRespDtoInjectionItem.to_json())

# convert the object into a dict
dto_generic_response_dto_pagination_resp_dto_injection_item_dict = dto_generic_response_dto_pagination_resp_dto_injection_item_instance.to_dict()
# create an instance of DtoGenericResponseDtoPaginationRespDtoInjectionItem from a dict
dto_generic_response_dto_pagination_resp_dto_injection_item_from_dict = DtoGenericResponseDtoPaginationRespDtoInjectionItem.from_dict(dto_generic_response_dto_pagination_resp_dto_injection_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


