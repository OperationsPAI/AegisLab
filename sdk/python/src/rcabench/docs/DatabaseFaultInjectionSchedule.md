# DatabaseFaultInjectionSchedule


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**benchmark** | **str** | Benchmark database, add index | [optional] 
**created_at** | **str** | Creation time, add time index | [optional] 
**description** | **str** | Description (optional field) | [optional] 
**display_config** | **str** | User-facing display configuration | [optional] 
**end_time** | **str** | Expected fault end time, add time index | [optional] 
**engine_config** | **str** | System-facing runtime configuration | [optional] 
**fault_type** | **int** | Fault type, add composite index | [optional] 
**id** | **int** | Unique identifier | [optional] 
**injection_name** | **str** | Name injected in k8s resources | [optional] 
**pre_duration** | **int** | Normal data duration | [optional] 
**start_time** | **str** | Expected fault start time, add time index | [optional] 
**status** | **int** | Status, add composite index | [optional] 
**task** | [**DatabaseTask**](DatabaseTask.md) | Foreign key association | [optional] 
**task_id** | **str** | Associated task ID, add composite index | [optional] 
**updated_at** | **str** | Update time | [optional] 

## Example

```python
from rcabench.openapi.models.database_fault_injection_schedule import DatabaseFaultInjectionSchedule

# TODO update the JSON string below
json = "{}"
# create an instance of DatabaseFaultInjectionSchedule from a JSON string
database_fault_injection_schedule_instance = DatabaseFaultInjectionSchedule.from_json(json)
# print the JSON string representation of the object
print(DatabaseFaultInjectionSchedule.to_json())

# convert the object into a dict
database_fault_injection_schedule_dict = database_fault_injection_schedule_instance.to_dict()
# create an instance of DatabaseFaultInjectionSchedule from a dict
database_fault_injection_schedule_from_dict = DatabaseFaultInjectionSchedule.from_dict(database_fault_injection_schedule_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


