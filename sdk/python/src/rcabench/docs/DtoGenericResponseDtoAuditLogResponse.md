# DtoGenericResponseDtoAuditLogResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | 状态码 | [optional] 
**data** | [**DtoAuditLogResponse**](DtoAuditLogResponse.md) | 泛型类型的数据 | [optional] 
**message** | **str** | 响应消息 | [optional] 
**timestamp** | **int** | 响应生成时间 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_audit_log_response import DtoGenericResponseDtoAuditLogResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoAuditLogResponse from a JSON string
dto_generic_response_dto_audit_log_response_instance = DtoGenericResponseDtoAuditLogResponse.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoAuditLogResponse.to_json())

# convert the object into a dict
dto_generic_response_dto_audit_log_response_dict = dto_generic_response_dto_audit_log_response_instance.to_dict()
# create an instance of DtoGenericResponseDtoAuditLogResponse from a dict
dto_generic_response_dto_audit_log_response_from_dict = DtoGenericResponseDtoAuditLogResponse.from_dict(dto_generic_response_dto_audit_log_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


