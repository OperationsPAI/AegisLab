# DtoQueryInjectionResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**benchmark** | **str** | Benchmark database, add index and size limit | [optional] 
**created_at** | **str** | Creation time, add time index | [optional] 
**description** | **str** | Description (optional field) | [optional] 
**display_config** | **str** | User-facing display configuration | [optional] 
**end_time** | **str** | Expected fault end time, nullable | [optional] 
**engine_config** | **str** | System-facing runtime configuration | [optional] 
**fault_type** | **int** | Fault type, add composite index | [optional] 
**ground_truth** | [**HandlerGroundtruth**](HandlerGroundtruth.md) |  | [optional] 
**id** | **int** | Unique identifier | [optional] 
**injection_name** | **str** | Name injected in k8s resources with size limit | [optional] 
**pre_duration** | **int** | Normal data duration | [optional] 
**start_time** | **str** | Expected fault start time, nullable with validation | [optional] 
**status** | **int** | Status: -1:deleted 0:disabled 1:enabled | [optional] 
**task** | [**DatabaseTask**](DatabaseTask.md) | Foreign key association | [optional] 
**task_id** | **str** | Associated task ID, add composite index | [optional] 
**updated_at** | **str** | Update time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_query_injection_resp import DtoQueryInjectionResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoQueryInjectionResp from a JSON string
dto_query_injection_resp_instance = DtoQueryInjectionResp.from_json(json)
# print the JSON string representation of the object
print(DtoQueryInjectionResp.to_json())

# convert the object into a dict
dto_query_injection_resp_dict = dto_query_injection_resp_instance.to_dict()
# create an instance of DtoQueryInjectionResp from a dict
dto_query_injection_resp_from_dict = DtoQueryInjectionResp.from_dict(dto_query_injection_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


