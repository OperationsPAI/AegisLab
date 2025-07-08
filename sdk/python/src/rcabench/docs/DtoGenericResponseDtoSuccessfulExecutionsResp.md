# DtoGenericResponseDtoSuccessfulExecutionsResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | 状态码 | [optional] 
**data** | [**List[DtoSuccessfulExecutionItem]**](DtoSuccessfulExecutionItem.md) | 泛型类型的数据 | [optional] 
**message** | **str** | 响应消息 | [optional] 
**timestamp** | **int** | 响应生成时间 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_successful_executions_resp import DtoGenericResponseDtoSuccessfulExecutionsResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoSuccessfulExecutionsResp from a JSON string
dto_generic_response_dto_successful_executions_resp_instance = DtoGenericResponseDtoSuccessfulExecutionsResp.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoSuccessfulExecutionsResp.to_json())

# convert the object into a dict
dto_generic_response_dto_successful_executions_resp_dict = dto_generic_response_dto_successful_executions_resp_instance.to_dict()
# create an instance of DtoGenericResponseDtoSuccessfulExecutionsResp from a dict
dto_generic_response_dto_successful_executions_resp_from_dict = DtoGenericResponseDtoSuccessfulExecutionsResp.from_dict(dto_generic_response_dto_successful_executions_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


