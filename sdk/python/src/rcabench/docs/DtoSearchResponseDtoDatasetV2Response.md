# DtoSearchResponseDtoDatasetV2Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**applied_filters** | [**List[DtoSearchFilter]**](DtoSearchFilter.md) |  | [optional] 
**applied_sort** | [**List[DtoSortOption]**](DtoSortOption.md) |  | [optional] 
**items** | [**List[DtoDatasetV2Response]**](DtoDatasetV2Response.md) |  | [optional] 
**pagination** | [**DtoPaginationInfo**](DtoPaginationInfo.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_search_response_dto_dataset_v2_response import DtoSearchResponseDtoDatasetV2Response

# TODO update the JSON string below
json = "{}"
# create an instance of DtoSearchResponseDtoDatasetV2Response from a JSON string
dto_search_response_dto_dataset_v2_response_instance = DtoSearchResponseDtoDatasetV2Response.from_json(json)
# print the JSON string representation of the object
print(DtoSearchResponseDtoDatasetV2Response.to_json())

# convert the object into a dict
dto_search_response_dto_dataset_v2_response_dict = dto_search_response_dto_dataset_v2_response_instance.to_dict()
# create an instance of DtoSearchResponseDtoDatasetV2Response from a dict
dto_search_response_dto_dataset_v2_response_from_dict = DtoSearchResponseDtoDatasetV2Response.from_dict(dto_search_response_dto_dataset_v2_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


