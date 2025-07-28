# DtoAssignUserToProjectRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**project_id** | **int** |  | 
**role_id** | **int** |  | 

## Example

```python
from rcabench.openapi.models.dto_assign_user_to_project_request import DtoAssignUserToProjectRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAssignUserToProjectRequest from a JSON string
dto_assign_user_to_project_request_instance = DtoAssignUserToProjectRequest.from_json(json)
# print the JSON string representation of the object
print(DtoAssignUserToProjectRequest.to_json())

# convert the object into a dict
dto_assign_user_to_project_request_dict = dto_assign_user_to_project_request_instance.to_dict()
# create an instance of DtoAssignUserToProjectRequest from a dict
dto_assign_user_to_project_request_from_dict = DtoAssignUserToProjectRequest.from_dict(dto_assign_user_to_project_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


