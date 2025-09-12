# DtoPermissionListResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DtoPermissionResponse]**](DtoPermissionResponse.md) |  | [optional] 
**pagination** | [**DtoPaginationInfo**](DtoPaginationInfo.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_permission_list_response import DtoPermissionListResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoPermissionListResponse from a JSON string
dto_permission_list_response_instance = DtoPermissionListResponse.from_json(json)
# print the JSON string representation of the object
print(DtoPermissionListResponse.to_json())

# convert the object into a dict
dto_permission_list_response_dict = dto_permission_list_response_instance.to_dict()
# create an instance of DtoPermissionListResponse from a dict
dto_permission_list_response_from_dict = DtoPermissionListResponse.from_dict(dto_permission_list_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


