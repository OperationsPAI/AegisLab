# DtoGenericResponseDtoRoleResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**DtoRoleResponse**](DtoRoleResponse.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_role_response import DtoGenericResponseDtoRoleResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoRoleResponse from a JSON string
dto_generic_response_dto_role_response_instance = DtoGenericResponseDtoRoleResponse.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoRoleResponse.to_json())

# convert the object into a dict
dto_generic_response_dto_role_response_dict = dto_generic_response_dto_role_response_instance.to_dict()
# create an instance of DtoGenericResponseDtoRoleResponse from a dict
dto_generic_response_dto_role_response_from_dict = DtoGenericResponseDtoRoleResponse.from_dict(dto_generic_response_dto_role_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


