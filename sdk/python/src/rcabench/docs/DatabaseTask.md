# DatabaseTask


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** |  | [optional] 
**cron_expr** | **str** |  | [optional] 
**execute_time** | **int** |  | [optional] 
**group_id** | **str** |  | [optional] 
**id** | **str** |  | [optional] 
**immediate** | **bool** |  | [optional] 
**payload** | **str** |  | [optional] 
**project_id** | **int** |  | [optional] 
**status** | **str** |  | [optional] 
**trace_id** | **str** |  | [optional] 
**type** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 

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


