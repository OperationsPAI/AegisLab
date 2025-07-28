# DtoCreatePermissionRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**action** | **str** |  | 
**description** | **str** |  | [optional] 
**display_name** | **str** |  | 
**name** | **str** |  | 
**resource_id** | **int** |  | 

## Example

```python
from rcabench.openapi.models.dto_create_permission_request import DtoCreatePermissionRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoCreatePermissionRequest from a JSON string
dto_create_permission_request_instance = DtoCreatePermissionRequest.from_json(json)
# print the JSON string representation of the object
print(DtoCreatePermissionRequest.to_json())

# convert the object into a dict
dto_create_permission_request_dict = dto_create_permission_request_instance.to_dict()
# create an instance of DtoCreatePermissionRequest from a dict
dto_create_permission_request_from_dict = DtoCreatePermissionRequest.from_dict(dto_create_permission_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


