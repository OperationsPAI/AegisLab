# DtoUserProjectResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**joined_at** | **str** |  | [optional] 
**project_id** | **int** |  | [optional] 
**project_name** | **str** |  | [optional] 
**role** | [**DtoRoleResponse**](DtoRoleResponse.md) |  | [optional] 
**status** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_user_project_response import DtoUserProjectResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoUserProjectResponse from a JSON string
dto_user_project_response_instance = DtoUserProjectResponse.from_json(json)
# print the JSON string representation of the object
print(DtoUserProjectResponse.to_json())

# convert the object into a dict
dto_user_project_response_dict = dto_user_project_response_instance.to_dict()
# create an instance of DtoUserProjectResponse from a dict
dto_user_project_response_from_dict = DtoUserProjectResponse.from_dict(dto_user_project_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


