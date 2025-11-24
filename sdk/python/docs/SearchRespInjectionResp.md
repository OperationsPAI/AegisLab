# SearchRespInjectionResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**applied_filters** | [**List[SearchFilter]**](SearchFilter.md) |  | [optional] 
**applied_sort** | [**List[SortOption]**](SortOption.md) |  | [optional] 
**items** | [**List[InjectionResp]**](InjectionResp.md) |  | [optional] 
**pagination** | [**PaginationInfo**](PaginationInfo.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.search_resp_injection_resp import SearchRespInjectionResp

# TODO update the JSON string below
json = "{}"
# create an instance of SearchRespInjectionResp from a JSON string
search_resp_injection_resp_instance = SearchRespInjectionResp.from_json(json)
# print the JSON string representation of the object
print(SearchRespInjectionResp.to_json())

# convert the object into a dict
search_resp_injection_resp_dict = search_resp_injection_resp_instance.to_dict()
# create an instance of SearchRespInjectionResp from a dict
search_resp_injection_resp_from_dict = SearchRespInjectionResp.from_dict(search_resp_injection_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


