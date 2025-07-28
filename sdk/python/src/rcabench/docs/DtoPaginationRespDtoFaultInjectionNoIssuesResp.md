# DtoPaginationRespDtoFaultInjectionNoIssuesResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DtoFaultInjectionNoIssuesResp]**](DtoFaultInjectionNoIssuesResp.md) |  | [optional] 
**total** | **int** |  | [optional] 
**total_pages** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_pagination_resp_dto_fault_injection_no_issues_resp import DtoPaginationRespDtoFaultInjectionNoIssuesResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoPaginationRespDtoFaultInjectionNoIssuesResp from a JSON string
dto_pagination_resp_dto_fault_injection_no_issues_resp_instance = DtoPaginationRespDtoFaultInjectionNoIssuesResp.from_json(json)
# print the JSON string representation of the object
print(DtoPaginationRespDtoFaultInjectionNoIssuesResp.to_json())

# convert the object into a dict
dto_pagination_resp_dto_fault_injection_no_issues_resp_dict = dto_pagination_resp_dto_fault_injection_no_issues_resp_instance.to_dict()
# create an instance of DtoPaginationRespDtoFaultInjectionNoIssuesResp from a dict
dto_pagination_resp_dto_fault_injection_no_issues_resp_from_dict = DtoPaginationRespDtoFaultInjectionNoIssuesResp.from_dict(dto_pagination_resp_dto_fault_injection_no_issues_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


