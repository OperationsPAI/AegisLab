# DtoSearchResponseDtoPermissionResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**applied_filters** | [**List[DtoSearchFilter]**](DtoSearchFilter.md) |  | [optional] 
**applied_sort** | [**List[DtoSortOption]**](DtoSortOption.md) |  | [optional] 
**items** | [**List[DtoPermissionResponse]**](DtoPermissionResponse.md) |  | [optional] 
**pagination** | [**DtoPaginationInfo**](DtoPaginationInfo.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_search_response_dto_permission_response import DtoSearchResponseDtoPermissionResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoSearchResponseDtoPermissionResponse from a JSON string
dto_search_response_dto_permission_response_instance = DtoSearchResponseDtoPermissionResponse.from_json(json)
# print the JSON string representation of the object
print(DtoSearchResponseDtoPermissionResponse.to_json())

# convert the object into a dict
dto_search_response_dto_permission_response_dict = dto_search_response_dto_permission_response_instance.to_dict()
# create an instance of DtoSearchResponseDtoPermissionResponse from a dict
dto_search_response_dto_permission_response_from_dict = DtoSearchResponseDtoPermissionResponse.from_dict(dto_search_response_dto_permission_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


