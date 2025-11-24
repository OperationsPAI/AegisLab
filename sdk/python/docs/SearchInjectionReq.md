# SearchInjectionReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**benchmarks** | **List[str]** |  | [optional] 
**created_at_gte** | **str** |  | [optional] 
**created_at_lte** | **str** |  | [optional] 
**end_time_gte** | **str** |  | [optional] 
**end_time_lte** | **str** |  | [optional] 
**fault_types** | **List[int]** |  | [optional] 
**include_labels** | **bool** | Whether to include labels in the response | [optional] 
**include_task** | **bool** | Whether to include task details in the response | [optional] 
**labels** | [**List[LabelItem]**](LabelItem.md) | Custom labels to filter by | [optional] 
**page** | **int** |  | [optional] 
**search** | **str** |  | [optional] 
**size** | **int** |  | [optional] 
**sort_by** | **str** |  | [optional] 
**sort_order** | **str** |  | [optional] 
**start_time_gte** | **str** |  | [optional] 
**start_time_lte** | **str** |  | [optional] 
**statuses** | **List[int]** |  | [optional] 
**tags** | **List[str]** | Tag values to filter by | [optional] 
**task_ids** | **List[str]** |  | [optional] 

## Example

```python
from rcabench.openapi.models.search_injection_req import SearchInjectionReq

# TODO update the JSON string below
json = "{}"
# create an instance of SearchInjectionReq from a JSON string
search_injection_req_instance = SearchInjectionReq.from_json(json)
# print the JSON string representation of the object
print(SearchInjectionReq.to_json())

# convert the object into a dict
search_injection_req_dict = search_injection_req_instance.to_dict()
# create an instance of SearchInjectionReq from a dict
search_injection_req_from_dict = SearchInjectionReq.from_dict(search_injection_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


