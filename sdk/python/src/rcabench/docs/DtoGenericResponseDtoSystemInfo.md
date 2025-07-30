# DtoGenericResponseDtoSystemInfo


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**DtoSystemInfo**](DtoSystemInfo.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_system_info import DtoGenericResponseDtoSystemInfo

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoSystemInfo from a JSON string
dto_generic_response_dto_system_info_instance = DtoGenericResponseDtoSystemInfo.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoSystemInfo.to_json())

# convert the object into a dict
dto_generic_response_dto_system_info_dict = dto_generic_response_dto_system_info_instance.to_dict()
# create an instance of DtoGenericResponseDtoSystemInfo from a dict
dto_generic_response_dto_system_info_from_dict = DtoGenericResponseDtoSystemInfo.from_dict(dto_generic_response_dto_system_info_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


