# DtoPermissionResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**action** | **str** |  | [optional] 
**created_at** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**display_name** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**is_system** | **bool** |  | [optional] 
**name** | **str** |  | [optional] 
**resource** | [**DtoResourceResponse**](DtoResourceResponse.md) |  | [optional] 
**resource_id** | **int** |  | [optional] 
**roles** | [**List[DtoRoleResponse]**](DtoRoleResponse.md) | Roles that have this permission | [optional] 
**status** | **int** |  | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_permission_response import DtoPermissionResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoPermissionResponse from a JSON string
dto_permission_response_instance = DtoPermissionResponse.from_json(json)
# print the JSON string representation of the object
print(DtoPermissionResponse.to_json())

# convert the object into a dict
dto_permission_response_dict = dto_permission_response_instance.to_dict()
# create an instance of DtoPermissionResponse from a dict
dto_permission_response_from_dict = DtoPermissionResponse.from_dict(dto_permission_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


