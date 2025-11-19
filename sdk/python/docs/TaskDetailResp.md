# TaskDetailResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** |  | [optional] 
**cron_expr** | **str** |  | [optional] 
**execute_time** | **int** |  | [optional] 
**group_id** | **str** |  | [optional] 
**id** | **str** |  | [optional] 
**immediate** | **bool** |  | [optional] 
**logs** | **List[str]** |  | [optional] 
**payload** | **object** |  | [optional] 
**project_id** | **int** |  | [optional] 
**project_name** | **str** |  | [optional] 
**state** | **str** |  | [optional] 
**status** | **str** |  | [optional] 
**trace_id** | **str** |  | [optional] 
**type** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from openapi.models.task_detail_resp import TaskDetailResp

# TODO update the JSON string below
json = "{}"
# create an instance of TaskDetailResp from a JSON string
task_detail_resp_instance = TaskDetailResp.from_json(json)
# print the JSON string representation of the object
print(TaskDetailResp.to_json())

# convert the object into a dict
task_detail_resp_dict = task_detail_resp_instance.to_dict()
# create an instance of TaskDetailResp from a dict
task_detail_resp_from_dict = TaskDetailResp.from_dict(task_detail_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


