# SearchContainerReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**command** | **str** |  | [optional] 
**created_at** | [**DateRange**](DateRange.md) | Common filters shortcuts | [optional] 
**exclude_fields** | **List[str]** |  | [optional] 
**filters** | [**List[SearchFilter]**](SearchFilter.md) | Filters | [optional] 
**image** | **str** |  | [optional] 
**include** | **List[str]** | Include related entities | [optional] 
**include_fields** | **List[str]** | Include/Exclude fields | [optional] 
**is_active** | **bool** |  | [optional] 
**keyword** | **str** | Search keyword (for general text search) | [optional] 
**name** | **str** | Container-specific filters | [optional] 
**page** | **int** | Pagination | [optional] 
**project_id** | **int** |  | [optional] 
**size** | **int** |  | [optional] 
**sort** | [**List[SortOption]**](SortOption.md) | Sort | [optional] 
**status** | **int** |  | [optional] 
**tag** | **str** |  | [optional] 
**type** | **str** |  | [optional] 
**updated_at** | [**DateRange**](DateRange.md) |  | [optional] 
**user_id** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.search_container_req import SearchContainerReq

# TODO update the JSON string below
json = "{}"
# create an instance of SearchContainerReq from a JSON string
search_container_req_instance = SearchContainerReq.from_json(json)
# print the JSON string representation of the object
print(SearchContainerReq.to_json())

# convert the object into a dict
search_container_req_dict = search_container_req_instance.to_dict()
# create an instance of SearchContainerReq from a dict
search_container_req_from_dict = SearchContainerReq.from_dict(search_container_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


