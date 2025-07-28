# DtoRoleResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**display_name** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**is_system** | **bool** |  | [optional] 
**name** | **str** |  | [optional] 
**permissions** | [**List[DtoPermissionResponse]**](DtoPermissionResponse.md) |  | [optional] 
**status** | **int** |  | [optional] 
**type** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 
**user_count** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_role_response import DtoRoleResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoRoleResponse from a JSON string
dto_role_response_instance = DtoRoleResponse.from_json(json)
# print the JSON string representation of the object
print(DtoRoleResponse.to_json())

# convert the object into a dict
dto_role_response_dict = dto_role_response_instance.to_dict()
# create an instance of DtoRoleResponse from a dict
dto_role_response_from_dict = DtoRoleResponse.from_dict(dto_role_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


