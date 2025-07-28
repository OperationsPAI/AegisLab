# DtoUpdateRoleRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**description** | **str** |  | [optional] 
**display_name** | **str** |  | [optional] 
**status** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_update_role_request import DtoUpdateRoleRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoUpdateRoleRequest from a JSON string
dto_update_role_request_instance = DtoUpdateRoleRequest.from_json(json)
# print the JSON string representation of the object
print(DtoUpdateRoleRequest.to_json())

# convert the object into a dict
dto_update_role_request_dict = dto_update_role_request_instance.to_dict()
# create an instance of DtoUpdateRoleRequest from a dict
dto_update_role_request_from_dict = DtoUpdateRoleRequest.from_dict(dto_update_role_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


