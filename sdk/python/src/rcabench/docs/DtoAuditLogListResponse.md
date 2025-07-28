# DtoAuditLogListResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DtoAuditLogResponse]**](DtoAuditLogResponse.md) |  | [optional] 
**pagination** | [**DtoPaginationInfo**](DtoPaginationInfo.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_audit_log_list_response import DtoAuditLogListResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAuditLogListResponse from a JSON string
dto_audit_log_list_response_instance = DtoAuditLogListResponse.from_json(json)
# print the JSON string representation of the object
print(DtoAuditLogListResponse.to_json())

# convert the object into a dict
dto_audit_log_list_response_dict = dto_audit_log_list_response_instance.to_dict()
# create an instance of DtoAuditLogListResponse from a dict
dto_audit_log_list_response_from_dict = DtoAuditLogListResponse.from_dict(dto_audit_log_list_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


