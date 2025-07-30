# DtoGranularityResultItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**confidence** | **float** |  | 
**level** | **str** |  | 
**rank** | **int** |  | 
**result** | **str** |  | 

## Example

```python
from rcabench.openapi.models.dto_granularity_result_item import DtoGranularityResultItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGranularityResultItem from a JSON string
dto_granularity_result_item_instance = DtoGranularityResultItem.from_json(json)
# print the JSON string representation of the object
print(DtoGranularityResultItem.to_json())

# convert the object into a dict
dto_granularity_result_item_dict = dto_granularity_result_item_instance.to_dict()
# create an instance of DtoGranularityResultItem from a dict
dto_granularity_result_item_from_dict = DtoGranularityResultItem.from_dict(dto_granularity_result_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


