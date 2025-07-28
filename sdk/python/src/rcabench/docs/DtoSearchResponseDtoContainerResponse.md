# DtoSearchResponseDtoContainerResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**applied_filters** | [**List[DtoSearchFilter]**](DtoSearchFilter.md) |  | [optional] 
**applied_sort** | [**List[DtoSortOption]**](DtoSortOption.md) |  | [optional] 
**items** | [**List[DtoContainerResponse]**](DtoContainerResponse.md) |  | [optional] 
**pagination** | [**DtoPaginationInfo**](DtoPaginationInfo.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_search_response_dto_container_response import DtoSearchResponseDtoContainerResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoSearchResponseDtoContainerResponse from a JSON string
dto_search_response_dto_container_response_instance = DtoSearchResponseDtoContainerResponse.from_json(json)
# print the JSON string representation of the object
print(DtoSearchResponseDtoContainerResponse.to_json())

# convert the object into a dict
dto_search_response_dto_container_response_dict = dto_search_response_dto_container_response_instance.to_dict()
# create an instance of DtoSearchResponseDtoContainerResponse from a dict
dto_search_response_dto_container_response_from_dict = DtoSearchResponseDtoContainerResponse.from_dict(dto_search_response_dto_container_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


