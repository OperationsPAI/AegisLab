# DtoGenericResponseDtoListTasksResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**List[DatabaseTask]**](DatabaseTask.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_list_tasks_resp import DtoGenericResponseDtoListTasksResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoListTasksResp from a JSON string
dto_generic_response_dto_list_tasks_resp_instance = DtoGenericResponseDtoListTasksResp.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoListTasksResp.to_json())

# convert the object into a dict
dto_generic_response_dto_list_tasks_resp_dict = dto_generic_response_dto_list_tasks_resp_instance.to_dict()
# create an instance of DtoGenericResponseDtoListTasksResp from a dict
dto_generic_response_dto_list_tasks_resp_from_dict = DtoGenericResponseDtoListTasksResp.from_dict(dto_generic_response_dto_list_tasks_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


