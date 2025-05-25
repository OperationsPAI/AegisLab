# DtoFaultInjectionNoIssuesResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**dataset_id** | **int** |  | [optional] 
**display_config** | **str** |  | [optional] 
**engine_config** | [**HandlerNode**](HandlerNode.md) |  | [optional] 
**injection_name** | **str** |  | [optional] 
**pre_duration** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_fault_injection_no_issues_resp import DtoFaultInjectionNoIssuesResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoFaultInjectionNoIssuesResp from a JSON string
dto_fault_injection_no_issues_resp_instance = DtoFaultInjectionNoIssuesResp.from_json(json)
# print the JSON string representation of the object
print(DtoFaultInjectionNoIssuesResp.to_json())

# convert the object into a dict
dto_fault_injection_no_issues_resp_dict = dto_fault_injection_no_issues_resp_instance.to_dict()
# create an instance of DtoFaultInjectionNoIssuesResp from a dict
dto_fault_injection_no_issues_resp_from_dict = DtoFaultInjectionNoIssuesResp.from_dict(dto_fault_injection_no_issues_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


