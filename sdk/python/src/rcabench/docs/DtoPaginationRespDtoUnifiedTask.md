# DtoPaginationRespDtoUnifiedTask


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DtoUnifiedTask]**](DtoUnifiedTask.md) |  | [optional] 
**total** | **int** |  | [optional] 
**total_pages** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_pagination_resp_dto_unified_task import DtoPaginationRespDtoUnifiedTask

# TODO update the JSON string below
json = "{}"
# create an instance of DtoPaginationRespDtoUnifiedTask from a JSON string
dto_pagination_resp_dto_unified_task_instance = DtoPaginationRespDtoUnifiedTask.from_json(json)
# print the JSON string representation of the object
print(DtoPaginationRespDtoUnifiedTask.to_json())

# convert the object into a dict
dto_pagination_resp_dto_unified_task_dict = dto_pagination_resp_dto_unified_task_instance.to_dict()
# create an instance of DtoPaginationRespDtoUnifiedTask from a dict
dto_pagination_resp_dto_unified_task_from_dict = DtoPaginationRespDtoUnifiedTask.from_dict(dto_pagination_resp_dto_unified_task_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


