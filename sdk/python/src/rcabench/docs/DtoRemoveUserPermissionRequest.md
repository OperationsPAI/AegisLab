# DtoRemoveUserPermissionRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**permission_id** | **int** |  | 
**project_id** | **int** |  | [optional] 
**user_id** | **int** |  | 

## Example

```python
from rcabench.openapi.models.dto_remove_user_permission_request import DtoRemoveUserPermissionRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoRemoveUserPermissionRequest from a JSON string
dto_remove_user_permission_request_instance = DtoRemoveUserPermissionRequest.from_json(json)
# print the JSON string representation of the object
print(DtoRemoveUserPermissionRequest.to_json())

# convert the object into a dict
dto_remove_user_permission_request_dict = dto_remove_user_permission_request_instance.to_dict()
# create an instance of DtoRemoveUserPermissionRequest from a dict
dto_remove_user_permission_request_from_dict = DtoRemoveUserPermissionRequest.from_dict(dto_remove_user_permission_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


