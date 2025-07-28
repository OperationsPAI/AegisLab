# DtoAssignUserPermissionRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**expires_at** | **str** |  | [optional] 
**grant_type** | **str** |  | 
**permission_id** | **int** |  | 
**project_id** | **int** |  | [optional] 
**user_id** | **int** |  | 

## Example

```python
from rcabench.openapi.models.dto_assign_user_permission_request import DtoAssignUserPermissionRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAssignUserPermissionRequest from a JSON string
dto_assign_user_permission_request_instance = DtoAssignUserPermissionRequest.from_json(json)
# print the JSON string representation of the object
print(DtoAssignUserPermissionRequest.to_json())

# convert the object into a dict
dto_assign_user_permission_request_dict = dto_assign_user_permission_request_instance.to_dict()
# create an instance of DtoAssignUserPermissionRequest from a dict
dto_assign_user_permission_request_from_dict = DtoAssignUserPermissionRequest.from_dict(dto_assign_user_permission_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


