# DtoGenericResponseArrayDtoFaultInjectionWithIssuesResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**List[DtoFaultInjectionWithIssuesResp]**](DtoFaultInjectionWithIssuesResp.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_array_dto_fault_injection_with_issues_resp import DtoGenericResponseArrayDtoFaultInjectionWithIssuesResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseArrayDtoFaultInjectionWithIssuesResp from a JSON string
dto_generic_response_array_dto_fault_injection_with_issues_resp_instance = DtoGenericResponseArrayDtoFaultInjectionWithIssuesResp.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseArrayDtoFaultInjectionWithIssuesResp.to_json())

# convert the object into a dict
dto_generic_response_array_dto_fault_injection_with_issues_resp_dict = dto_generic_response_array_dto_fault_injection_with_issues_resp_instance.to_dict()
# create an instance of DtoGenericResponseArrayDtoFaultInjectionWithIssuesResp from a dict
dto_generic_response_array_dto_fault_injection_with_issues_resp_from_dict = DtoGenericResponseArrayDtoFaultInjectionWithIssuesResp.from_dict(dto_generic_response_array_dto_fault_injection_with_issues_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


