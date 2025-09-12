# DtoAssignUserRoleRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**role_id** | **int** |  | 
**user_id** | **int** |  | 

## Example

```python
from rcabench.openapi.models.dto_assign_user_role_request import DtoAssignUserRoleRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAssignUserRoleRequest from a JSON string
dto_assign_user_role_request_instance = DtoAssignUserRoleRequest.from_json(json)
# print the JSON string representation of the object
print(DtoAssignUserRoleRequest.to_json())

# convert the object into a dict
dto_assign_user_role_request_dict = dto_assign_user_role_request_instance.to_dict()
# create an instance of DtoAssignUserRoleRequest from a dict
dto_assign_user_role_request_from_dict = DtoAssignUserRoleRequest.from_dict(dto_assign_user_role_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


