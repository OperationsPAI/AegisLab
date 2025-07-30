# DtoGenericResponseDtoPaginationRespDtoUnifiedTask


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**DtoPaginationRespDtoUnifiedTask**](DtoPaginationRespDtoUnifiedTask.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_pagination_resp_dto_unified_task import DtoGenericResponseDtoPaginationRespDtoUnifiedTask

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoPaginationRespDtoUnifiedTask from a JSON string
dto_generic_response_dto_pagination_resp_dto_unified_task_instance = DtoGenericResponseDtoPaginationRespDtoUnifiedTask.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoPaginationRespDtoUnifiedTask.to_json())

# convert the object into a dict
dto_generic_response_dto_pagination_resp_dto_unified_task_dict = dto_generic_response_dto_pagination_resp_dto_unified_task_instance.to_dict()
# create an instance of DtoGenericResponseDtoPaginationRespDtoUnifiedTask from a dict
dto_generic_response_dto_pagination_resp_dto_unified_task_from_dict = DtoGenericResponseDtoPaginationRespDtoUnifiedTask.from_dict(dto_generic_response_dto_pagination_resp_dto_unified_task_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


