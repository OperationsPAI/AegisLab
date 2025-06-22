# DtoFaultInjectionStatisticsResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**no_issues_count** | **int** |  | [optional] 
**total_count** | **int** |  | [optional] 
**with_issues_count** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_fault_injection_statistics_resp import DtoFaultInjectionStatisticsResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoFaultInjectionStatisticsResp from a JSON string
dto_fault_injection_statistics_resp_instance = DtoFaultInjectionStatisticsResp.from_json(json)
# print the JSON string representation of the object
print(DtoFaultInjectionStatisticsResp.to_json())

# convert the object into a dict
dto_fault_injection_statistics_resp_dict = dto_fault_injection_statistics_resp_instance.to_dict()
# create an instance of DtoFaultInjectionStatisticsResp from a dict
dto_fault_injection_statistics_resp_from_dict = DtoFaultInjectionStatisticsResp.from_dict(dto_fault_injection_statistics_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


