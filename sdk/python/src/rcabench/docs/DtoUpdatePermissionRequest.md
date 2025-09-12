# DtoUpdatePermissionRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**action** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**display_name** | **str** |  | [optional] 
**resource_id** | **int** |  | [optional] 
**status** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_update_permission_request import DtoUpdatePermissionRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoUpdatePermissionRequest from a JSON string
dto_update_permission_request_instance = DtoUpdatePermissionRequest.from_json(json)
# print the JSON string representation of the object
print(DtoUpdatePermissionRequest.to_json())

# convert the object into a dict
dto_update_permission_request_dict = dto_update_permission_request_instance.to_dict()
# create an instance of DtoUpdatePermissionRequest from a dict
dto_update_permission_request_from_dict = DtoUpdatePermissionRequest.from_dict(dto_update_permission_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


