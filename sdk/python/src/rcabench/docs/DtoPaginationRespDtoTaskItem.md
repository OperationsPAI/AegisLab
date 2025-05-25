# DtoPaginationRespDtoTaskItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DtoTaskItem]**](DtoTaskItem.md) |  | [optional] 
**total** | **int** |  | [optional] 
**total_pages** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_pagination_resp_dto_task_item import DtoPaginationRespDtoTaskItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoPaginationRespDtoTaskItem from a JSON string
dto_pagination_resp_dto_task_item_instance = DtoPaginationRespDtoTaskItem.from_json(json)
# print the JSON string representation of the object
print(DtoPaginationRespDtoTaskItem.to_json())

# convert the object into a dict
dto_pagination_resp_dto_task_item_dict = dto_pagination_resp_dto_task_item_instance.to_dict()
# create an instance of DtoPaginationRespDtoTaskItem from a dict
dto_pagination_resp_dto_task_item_from_dict = DtoPaginationRespDtoTaskItem.from_dict(dto_pagination_resp_dto_task_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


