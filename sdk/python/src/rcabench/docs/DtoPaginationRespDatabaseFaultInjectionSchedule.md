# DtoPaginationRespDatabaseFaultInjectionSchedule


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DatabaseFaultInjectionSchedule]**](DatabaseFaultInjectionSchedule.md) |  | [optional] 
**total** | **int** |  | [optional] 
**total_pages** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_pagination_resp_database_fault_injection_schedule import DtoPaginationRespDatabaseFaultInjectionSchedule

# TODO update the JSON string below
json = "{}"
# create an instance of DtoPaginationRespDatabaseFaultInjectionSchedule from a JSON string
dto_pagination_resp_database_fault_injection_schedule_instance = DtoPaginationRespDatabaseFaultInjectionSchedule.from_json(json)
# print the JSON string representation of the object
print(DtoPaginationRespDatabaseFaultInjectionSchedule.to_json())

# convert the object into a dict
dto_pagination_resp_database_fault_injection_schedule_dict = dto_pagination_resp_database_fault_injection_schedule_instance.to_dict()
# create an instance of DtoPaginationRespDatabaseFaultInjectionSchedule from a dict
dto_pagination_resp_database_fault_injection_schedule_from_dict = DtoPaginationRespDatabaseFaultInjectionSchedule.from_dict(dto_pagination_resp_database_fault_injection_schedule_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


