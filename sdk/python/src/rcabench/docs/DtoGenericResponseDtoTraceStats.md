# DtoGenericResponseDtoTraceStats


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**DtoTraceStats**](DtoTraceStats.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_trace_stats import DtoGenericResponseDtoTraceStats

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoTraceStats from a JSON string
dto_generic_response_dto_trace_stats_instance = DtoGenericResponseDtoTraceStats.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoTraceStats.to_json())

# convert the object into a dict
dto_generic_response_dto_trace_stats_dict = dto_generic_response_dto_trace_stats_instance.to_dict()
# create an instance of DtoGenericResponseDtoTraceStats from a dict
dto_generic_response_dto_trace_stats_from_dict = DtoGenericResponseDtoTraceStats.from_dict(dto_generic_response_dto_trace_stats_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


