# DatabaseTask


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** | 添加时间索引 | [optional] 
**cron_expr** | **str** |  | [optional] 
**execute_time** | **int** | 添加执行时间索引 | [optional] 
**group_id** | **str** | 添加组ID索引 | [optional] 
**id** | **str** |  | [optional] 
**immediate** | **bool** |  | [optional] 
**payload** | **str** |  | [optional] 
**project** | [**DatabaseProject**](DatabaseProject.md) | 外键关联 | [optional] 
**project_id** | **int** | 任务可以属于某个项目（可选） | [optional] 
**status** | **str** | 添加多个复合索引 | [optional] 
**trace_id** | **str** | 添加追踪ID索引 | [optional] 
**type** | **str** | 添加复合索引 | [optional] 
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


