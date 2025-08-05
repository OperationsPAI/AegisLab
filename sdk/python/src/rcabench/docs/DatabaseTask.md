# DatabaseTask


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** | Creation time with index | [optional] 
**cron_expr** | **str** | Cron expression with size limit | [optional] 
**execute_time** | **int** | Execution time timestamp | [optional] 
**group_id** | **str** | Group ID with size limit | [optional] 
**id** | **str** | Task ID with size limit | [optional] 
**immediate** | **bool** | Whether to execute immediately | [optional] 
**payload** | **str** | Task payload | [optional] 
**project** | [**DatabaseProject**](DatabaseProject.md) | Foreign key association | [optional] 
**project_id** | **int** | Task can belong to a project (optional) | [optional] 
**status** | **str** | Status: Pending, Running, Completed, Error, Cancelled, Rescheduled | [optional] 
**trace_id** | **str** | Trace ID with size limit | [optional] 
**type** | **str** | Task type with size limit | [optional] 
**updated_at** | **str** | Update time | [optional] 

## Example

```python
from rcabench.openapi.models.database_task import DatabaseTask

# TODO update the JSON string below
json = "{}"
# create an instance of DatabaseTask from a JSON string
database_task_instance = DatabaseTask.from_json(json)
# print the JSON string representation of the object
print(DatabaseTask.to_json())

# convert the object into a dict
database_task_dict = database_task_instance.to_dict()
# create an instance of DatabaseTask from a dict
database_task_from_dict = DatabaseTask.from_dict(database_task_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


