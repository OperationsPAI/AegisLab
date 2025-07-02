# DtoGenericResponseArrayDatabaseFaultInjectionSchedule


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | 状态码 | [optional] 
**data** | [**List[DatabaseFaultInjectionSchedule]**](DatabaseFaultInjectionSchedule.md) | 泛型类型的数据 | [optional] 
**message** | **str** | 响应消息 | [optional] 
**timestamp** | **int** | 响应生成时间 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_array_database_fault_injection_schedule import DtoGenericResponseArrayDatabaseFaultInjectionSchedule

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseArrayDatabaseFaultInjectionSchedule from a JSON string
dto_generic_response_array_database_fault_injection_schedule_instance = DtoGenericResponseArrayDatabaseFaultInjectionSchedule.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseArrayDatabaseFaultInjectionSchedule.to_json())

# convert the object into a dict
dto_generic_response_array_database_fault_injection_schedule_dict = dto_generic_response_array_database_fault_injection_schedule_instance.to_dict()
# create an instance of DtoGenericResponseArrayDatabaseFaultInjectionSchedule from a dict
dto_generic_response_array_database_fault_injection_schedule_from_dict = DtoGenericResponseArrayDatabaseFaultInjectionSchedule.from_dict(dto_generic_response_array_database_fault_injection_schedule_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


