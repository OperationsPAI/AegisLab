# SearchRespContainerResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**applied_filters** | [**List[SearchFilter]**](SearchFilter.md) |  | [optional] 
**applied_sort** | [**List[SortOption]**](SortOption.md) |  | [optional] 
**items** | [**List[ContainerResp]**](ContainerResp.md) |  | [optional] 
**pagination** | [**PaginationInfo**](PaginationInfo.md) |  | [optional] 

## Example

```python
from openapi.models.search_resp_container_resp import SearchRespContainerResp

# TODO update the JSON string below
json = "{}"
# create an instance of SearchRespContainerResp from a JSON string
search_resp_container_resp_instance = SearchRespContainerResp.from_json(json)
# print the JSON string representation of the object
print(SearchRespContainerResp.to_json())

# convert the object into a dict
search_resp_container_resp_dict = search_resp_container_resp_instance.to_dict()
# create an instance of SearchRespContainerResp from a dict
search_resp_container_resp_from_dict = SearchRespContainerResp.from_dict(search_resp_container_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


