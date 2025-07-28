# DtoAuditLogResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**action** | **str** |  | [optional] 
**details** | **str** |  | [optional] 
**error** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**ip_address** | **str** |  | [optional] 
**resource** | **str** |  | [optional] 
**resource_id** | **str** |  | [optional] 
**success** | **bool** |  | [optional] 
**timestamp** | **str** |  | [optional] 
**user_agent** | **str** |  | [optional] 
**user_id** | **int** |  | [optional] 
**username** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_audit_log_response import DtoAuditLogResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAuditLogResponse from a JSON string
dto_audit_log_response_instance = DtoAuditLogResponse.from_json(json)
# print the JSON string representation of the object
print(DtoAuditLogResponse.to_json())

# convert the object into a dict
dto_audit_log_response_dict = dto_audit_log_response_instance.to_dict()
# create an instance of DtoAuditLogResponse from a dict
dto_audit_log_response_from_dict = DtoAuditLogResponse.from_dict(dto_audit_log_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


