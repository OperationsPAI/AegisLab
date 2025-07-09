# DtoGenericResponseDtoPaginationRespDatabaseFaultInjectionSchedule


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | 状态码 | [optional] 
**data** | [**DtoPaginationRespDatabaseFaultInjectionSchedule**](DtoPaginationRespDatabaseFaultInjectionSchedule.md) | 泛型类型的数据 | [optional] 
**message** | **str** | 响应消息 | [optional] 
**timestamp** | **int** | 响应生成时间 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_pagination_resp_database_fault_injection_schedule import DtoGenericResponseDtoPaginationRespDatabaseFaultInjectionSchedule

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoPaginationRespDatabaseFaultInjectionSchedule from a JSON string
dto_generic_response_dto_pagination_resp_database_fault_injection_schedule_instance = DtoGenericResponseDtoPaginationRespDatabaseFaultInjectionSchedule.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoPaginationRespDatabaseFaultInjectionSchedule.to_json())

# convert the object into a dict
dto_generic_response_dto_pagination_resp_database_fault_injection_schedule_dict = dto_generic_response_dto_pagination_resp_database_fault_injection_schedule_instance.to_dict()
# create an instance of DtoGenericResponseDtoPaginationRespDatabaseFaultInjectionSchedule from a dict
dto_generic_response_dto_pagination_resp_database_fault_injection_schedule_from_dict = DtoGenericResponseDtoPaginationRespDatabaseFaultInjectionSchedule.from_dict(dto_generic_response_dto_pagination_resp_database_fault_injection_schedule_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


