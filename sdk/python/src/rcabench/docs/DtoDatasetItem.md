# DtoDatasetItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**end_time** | **str** |  | [optional] 
**name** | **str** |  | [optional] 
**param** | **List[object]** |  | [optional] 
**start_time** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_item import DtoDatasetItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetItem from a JSON string
dto_dataset_item_instance = DtoDatasetItem.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetItem.to_json())

# convert the object into a dict
dto_dataset_item_dict = dto_dataset_item_instance.to_dict()
# create an instance of DtoDatasetItem from a dict
dto_dataset_item_from_dict = DtoDatasetItem.from_dict(dto_dataset_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


