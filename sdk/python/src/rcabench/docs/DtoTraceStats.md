# DtoTraceStats


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**avg_duration** | **float** |  | [optional] 
**end_count_map** | **Dict[str, Dict[str, int]]** |  | [optional] 
**fault_injection_traces** | **List[str]** |  | [optional] 
**max_duration** | **float** |  | [optional] 
**min_duration** | **float** |  | [optional] 
**total** | **int** |  | [optional] 
**trace_completed_list** | **List[str]** |  | [optional] 
**trace_errors** | **object** |  | [optional] 
**trace_status_time_map** | **Dict[str, Dict[str, float]]** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_trace_stats import DtoTraceStats

# TODO update the JSON string below
json = "{}"
# create an instance of DtoTraceStats from a JSON string
dto_trace_stats_instance = DtoTraceStats.from_json(json)
# print the JSON string representation of the object
print(DtoTraceStats.to_json())

# convert the object into a dict
dto_trace_stats_dict = dto_trace_stats_instance.to_dict()
# create an instance of DtoTraceStats from a dict
dto_trace_stats_from_dict = DtoTraceStats.from_dict(dto_trace_stats_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


