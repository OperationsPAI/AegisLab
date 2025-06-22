# DtoGenericResponseAny


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | 状态码 | [optional] 
**data** | **object** | 泛型类型的数据 | [optional] 
**message** | **str** | 响应消息 | [optional] 
**timestamp** | **int** | 响应生成时间 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_any import DtoGenericResponseAny

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseAny from a JSON string
dto_generic_response_any_instance = DtoGenericResponseAny.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseAny.to_json())

# convert the object into a dict
dto_generic_response_any_dict = dto_generic_response_any_instance.to_dict()
# create an instance of DtoGenericResponseAny from a dict
dto_generic_response_any_from_dict = DtoGenericResponseAny.from_dict(dto_generic_response_any_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


