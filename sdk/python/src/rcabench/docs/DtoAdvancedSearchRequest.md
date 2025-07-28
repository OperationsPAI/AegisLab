# DtoAdvancedSearchRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | [**DtoDateRange**](DtoDateRange.md) | Common filters shortcuts | [optional] 
**exclude_fields** | **List[str]** |  | [optional] 
**filters** | [**List[DtoSearchFilter]**](DtoSearchFilter.md) | Filters | [optional] 
**include** | **List[str]** | Include related entities | [optional] 
**include_fields** | **List[str]** | Include/Exclude fields | [optional] 
**is_active** | **bool** |  | [optional] 
**keyword** | **str** | Search keyword (for general text search) | [optional] 
**page** | **int** | Pagination | [optional] 
**project_id** | **int** |  | [optional] 
**size** | **int** |  | [optional] 
**sort** | [**List[DtoSortOption]**](DtoSortOption.md) | Sort | [optional] 
**status** | **List[int]** |  | [optional] 
**updated_at** | [**DtoDateRange**](DtoDateRange.md) |  | [optional] 
**user_id** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_advanced_search_request import DtoAdvancedSearchRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAdvancedSearchRequest from a JSON string
dto_advanced_search_request_instance = DtoAdvancedSearchRequest.from_json(json)
# print the JSON string representation of the object
print(DtoAdvancedSearchRequest.to_json())

# convert the object into a dict
dto_advanced_search_request_dict = dto_advanced_search_request_instance.to_dict()
# create an instance of DtoAdvancedSearchRequest from a dict
dto_advanced_search_request_from_dict = DtoAdvancedSearchRequest.from_dict(dto_advanced_search_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


