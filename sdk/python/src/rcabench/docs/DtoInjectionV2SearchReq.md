# DtoInjectionV2SearchReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**benchmarks** | **List[str]** |  | [optional] 
**created_at_gte** | **str** |  | [optional] 
**created_at_lte** | **str** |  | [optional] 
**end_time_gte** | **str** |  | [optional] 
**end_time_lte** | **str** |  | [optional] 
**fault_types** | **List[int]** |  | [optional] 
**include** | **str** |  | [optional] 
**labels** | [**List[DtoLabelItem]**](DtoLabelItem.md) | Custom labels to filter by | [optional] 
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
from rcabench.openapi.models.dto_injection_v2_search_req import DtoInjectionV2SearchReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionV2SearchReq from a JSON string
dto_injection_v2_search_req_instance = DtoInjectionV2SearchReq.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionV2SearchReq.to_json())

# convert the object into a dict
dto_injection_v2_search_req_dict = dto_injection_v2_search_req_instance.to_dict()
# create an instance of DtoInjectionV2SearchReq from a dict
dto_injection_v2_search_req_from_dict = DtoInjectionV2SearchReq.from_dict(dto_injection_v2_search_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


