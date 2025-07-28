# DtoCreateRoleRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**description** | **str** |  | [optional] 
**display_name** | **str** |  | 
**name** | **str** |  | 
**type** | **str** |  | 

## Example

```python
from rcabench.openapi.models.dto_create_role_request import DtoCreateRoleRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoCreateRoleRequest from a JSON string
dto_create_role_request_instance = DtoCreateRoleRequest.from_json(json)
# print the JSON string representation of the object
print(DtoCreateRoleRequest.to_json())

# convert the object into a dict
dto_create_role_request_dict = dto_create_role_request_instance.to_dict()
# create an instance of DtoCreateRoleRequest from a dict
dto_create_role_request_from_dict = DtoCreateRoleRequest.from_dict(dto_create_role_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


