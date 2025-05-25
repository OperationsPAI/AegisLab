# DtoPaginationRespDtoDatasetItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DtoDatasetItem]**](DtoDatasetItem.md) |  | [optional] 
**total** | **int** |  | [optional] 
**total_pages** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_pagination_resp_dto_dataset_item import DtoPaginationRespDtoDatasetItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoPaginationRespDtoDatasetItem from a JSON string
dto_pagination_resp_dto_dataset_item_instance = DtoPaginationRespDtoDatasetItem.from_json(json)
# print the JSON string representation of the object
print(DtoPaginationRespDtoDatasetItem.to_json())

# convert the object into a dict
dto_pagination_resp_dto_dataset_item_dict = dto_pagination_resp_dto_dataset_item_instance.to_dict()
# create an instance of DtoPaginationRespDtoDatasetItem from a dict
dto_pagination_resp_dto_dataset_item_from_dict = DtoPaginationRespDtoDatasetItem.from_dict(dto_pagination_resp_dto_dataset_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


