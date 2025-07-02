# DtoGenericResponseArrayString


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | 状态码 | [optional] 
**data** | **List[str]** | 泛型类型的数据 | [optional] 
**message** | **str** | 响应消息 | [optional] 
**timestamp** | **int** | 响应生成时间 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_array_string import DtoGenericResponseArrayString

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseArrayString from a JSON string
dto_generic_response_array_string_instance = DtoGenericResponseArrayString.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseArrayString.to_json())

# convert the object into a dict
dto_generic_response_array_string_dict = dto_generic_response_array_string_instance.to_dict()
# create an instance of DtoGenericResponseArrayString from a dict
dto_generic_response_array_string_from_dict = DtoGenericResponseArrayString.from_dict(dto_generic_response_array_string_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


