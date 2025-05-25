# DtoGenericResponseArrayDtoFaultInjectionNoIssuesResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | 状态码 | [optional] 
**data** | [**List[DtoFaultInjectionNoIssuesResp]**](DtoFaultInjectionNoIssuesResp.md) | 泛型类型的数据 | [optional] 
**message** | **str** | 响应消息 | [optional] 
**timestamp** | **int** | 响应生成时间 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_array_dto_fault_injection_no_issues_resp import DtoGenericResponseArrayDtoFaultInjectionNoIssuesResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseArrayDtoFaultInjectionNoIssuesResp from a JSON string
dto_generic_response_array_dto_fault_injection_no_issues_resp_instance = DtoGenericResponseArrayDtoFaultInjectionNoIssuesResp.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseArrayDtoFaultInjectionNoIssuesResp.to_json())

# convert the object into a dict
dto_generic_response_array_dto_fault_injection_no_issues_resp_dict = dto_generic_response_array_dto_fault_injection_no_issues_resp_instance.to_dict()
# create an instance of DtoGenericResponseArrayDtoFaultInjectionNoIssuesResp from a dict
dto_generic_response_array_dto_fault_injection_no_issues_resp_from_dict = DtoGenericResponseArrayDtoFaultInjectionNoIssuesResp.from_dict(dto_generic_response_array_dto_fault_injection_no_issues_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


