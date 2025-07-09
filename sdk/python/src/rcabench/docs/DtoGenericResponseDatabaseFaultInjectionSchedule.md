# DtoGenericResponseDatabaseFaultInjectionSchedule


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | 状态码 | [optional] 
**data** | [**DatabaseFaultInjectionSchedule**](DatabaseFaultInjectionSchedule.md) | 泛型类型的数据 | [optional] 
**message** | **str** | 响应消息 | [optional] 
**timestamp** | **int** | 响应生成时间 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_database_fault_injection_schedule import DtoGenericResponseDatabaseFaultInjectionSchedule

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDatabaseFaultInjectionSchedule from a JSON string
dto_generic_response_database_fault_injection_schedule_instance = DtoGenericResponseDatabaseFaultInjectionSchedule.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDatabaseFaultInjectionSchedule.to_json())

# convert the object into a dict
dto_generic_response_database_fault_injection_schedule_dict = dto_generic_response_database_fault_injection_schedule_instance.to_dict()
# create an instance of DtoGenericResponseDatabaseFaultInjectionSchedule from a dict
dto_generic_response_database_fault_injection_schedule_from_dict = DtoGenericResponseDatabaseFaultInjectionSchedule.from_dict(dto_generic_response_database_fault_injection_schedule_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


