# DtoUnifiedTask


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**cron_expr** | **str** | Cron expression for recurring tasks | [optional] 
**execute_time** | **int** | Unix timestamp for delayed execution | [optional] 
**group_carrier** | **Dict[str, str]** | Carrier for group context | [optional] 
**group_id** | **str** | ID for grouping tasks | [optional] 
**immediate** | **bool** | Whether to execute immediately | [optional] 
**payload** | **object** | Task-specific data | [optional] 
**restart_num** | **int** | Number of restarts for the task | [optional] 
**retry_policy** | [**DtoRetryPolicy**](DtoRetryPolicy.md) | Policy for retrying failed tasks | [optional] 
**task_id** | **str** | Unique identifier for the task | [optional] 
**trace_carrier** | **Dict[str, str]** | Carrier for trace context | [optional] 
**trace_id** | **str** | ID for tracing related tasks | [optional] 
**type** | [**ConstsTaskType**](ConstsTaskType.md) | Task type (determines how it&#39;s processed) | [optional] 

## Example

```python
from rcabench.openapi.models.dto_unified_task import DtoUnifiedTask

# TODO update the JSON string below
json = "{}"
# create an instance of DtoUnifiedTask from a JSON string
dto_unified_task_instance = DtoUnifiedTask.from_json(json)
# print the JSON string representation of the object
print(DtoUnifiedTask.to_json())

# convert the object into a dict
dto_unified_task_dict = dto_unified_task_instance.to_dict()
# create an instance of DtoUnifiedTask from a dict
dto_unified_task_from_dict = DtoUnifiedTask.from_dict(dto_unified_task_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


