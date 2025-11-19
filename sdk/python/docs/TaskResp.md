# TaskResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** |  | [optional] 
**cron_expr** | **str** |  | [optional] 
**execute_time** | **int** |  | [optional] 
**group_id** | **str** |  | [optional] 
**id** | **str** |  | [optional] 
**immediate** | **bool** |  | [optional] 
**project_id** | **int** |  | [optional] 
**project_name** | **str** |  | [optional] 
**state** | **str** |  | [optional] 
**status** | **str** |  | [optional] 
**trace_id** | **str** |  | [optional] 
**type** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from openapi.models.task_resp import TaskResp

# TODO update the JSON string below
json = "{}"
# create an instance of TaskResp from a JSON string
task_resp_instance = TaskResp.from_json(json)
# print the JSON string representation of the object
print(TaskResp.to_json())

# convert the object into a dict
task_resp_dict = task_resp_instance.to_dict()
# create an instance of TaskResp from a dict
task_resp_from_dict = TaskResp.from_dict(task_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


