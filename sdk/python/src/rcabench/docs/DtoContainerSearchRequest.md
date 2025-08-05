# DtoContainerSearchRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**command** | **str** |  | [optional] 
**created_at** | [**DtoDateRange**](DtoDateRange.md) | Common filters shortcuts | [optional] 
**exclude_fields** | **List[str]** |  | [optional] 
**filters** | [**List[DtoSearchFilter]**](DtoSearchFilter.md) | Filters | [optional] 
**image** | **str** |  | [optional] 
**include** | **List[str]** | Include related entities | [optional] 
**include_fields** | **List[str]** | Include/Exclude fields | [optional] 
**is_active** | **bool** |  | [optional] 
**keyword** | **str** | Search keyword (for general text search) | [optional] 
**name** | **str** | Container-specific filters | [optional] 
**page** | **int** | Pagination | [optional] 
**project_id** | **int** |  | [optional] 
**size** | **int** |  | [optional] 
**sort** | [**List[DtoSortOption]**](DtoSortOption.md) | Sort | [optional] 
**status** | **int** |  | [optional] 
**tag** | **str** |  | [optional] 
**type** | **str** |  | [optional] 
**updated_at** | [**DtoDateRange**](DtoDateRange.md) |  | [optional] 
**user_id** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_container_search_request import DtoContainerSearchRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoContainerSearchRequest from a JSON string
dto_container_search_request_instance = DtoContainerSearchRequest.from_json(json)
# print the JSON string representation of the object
print(DtoContainerSearchRequest.to_json())

# convert the object into a dict
dto_container_search_request_dict = dto_container_search_request_instance.to_dict()
# create an instance of DtoContainerSearchRequest from a dict
dto_container_search_request_from_dict = DtoContainerSearchRequest.from_dict(dto_container_search_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


