# DtoInjectionStatsResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**no_issues_injections** | **int** |  | [optional] 
**no_issues_records** | **int** |  | [optional] 
**with_issues_injections** | **int** |  | [optional] 
**with_issues_records** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_stats_resp import DtoInjectionStatsResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionStatsResp from a JSON string
dto_injection_stats_resp_instance = DtoInjectionStatsResp.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionStatsResp.to_json())

# convert the object into a dict
dto_injection_stats_resp_dict = dto_injection_stats_resp_instance.to_dict()
# create an instance of DtoInjectionStatsResp from a dict
dto_injection_stats_resp_from_dict = DtoInjectionStatsResp.from_dict(dto_injection_stats_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


