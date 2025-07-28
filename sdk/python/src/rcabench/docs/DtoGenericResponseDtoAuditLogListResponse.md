# DtoGenericResponseDtoAuditLogListResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | 状态码 | [optional] 
**data** | [**DtoAuditLogListResponse**](DtoAuditLogListResponse.md) | 泛型类型的数据 | [optional] 
**message** | **str** | 响应消息 | [optional] 
**timestamp** | **int** | 响应生成时间 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_audit_log_list_response import DtoGenericResponseDtoAuditLogListResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoAuditLogListResponse from a JSON string
dto_generic_response_dto_audit_log_list_response_instance = DtoGenericResponseDtoAuditLogListResponse.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoAuditLogListResponse.to_json())

# convert the object into a dict
dto_generic_response_dto_audit_log_list_response_dict = dto_generic_response_dto_audit_log_list_response_instance.to_dict()
# create an instance of DtoGenericResponseDtoAuditLogListResponse from a dict
dto_generic_response_dto_audit_log_list_response_from_dict = DtoGenericResponseDtoAuditLogListResponse.from_dict(dto_generic_response_dto_audit_log_list_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


