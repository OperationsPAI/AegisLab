# DtoFaultInjectionWithIssuesResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**dataset_id** | **int** |  | [optional] 
**display_config** | **str** |  | [optional] 
**engine_config** | [**HandlerNode**](HandlerNode.md) |  | [optional] 
**injection_name** | **str** |  | [optional] 
**issues** | **str** |  | [optional] 
**pre_duration** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_fault_injection_with_issues_resp import DtoFaultInjectionWithIssuesResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoFaultInjectionWithIssuesResp from a JSON string
dto_fault_injection_with_issues_resp_instance = DtoFaultInjectionWithIssuesResp.from_json(json)
# print the JSON string representation of the object
print(DtoFaultInjectionWithIssuesResp.to_json())

# convert the object into a dict
dto_fault_injection_with_issues_resp_dict = dto_fault_injection_with_issues_resp_instance.to_dict()
# create an instance of DtoFaultInjectionWithIssuesResp from a dict
dto_fault_injection_with_issues_resp_from_dict = DtoFaultInjectionWithIssuesResp.from_dict(dto_fault_injection_with_issues_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


