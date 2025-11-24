# GranularityResultItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**confidence** | **float** |  | [optional] 
**level** | **str** |  | 
**rank** | **int** |  | 
**result** | **str** |  | 

## Example

```python
from rcabench.openapi.models.granularity_result_item import GranularityResultItem

# TODO update the JSON string below
json = "{}"
# create an instance of GranularityResultItem from a JSON string
granularity_result_item_instance = GranularityResultItem.from_json(json)
# print the JSON string representation of the object
print(GranularityResultItem.to_json())

# convert the object into a dict
granularity_result_item_dict = granularity_result_item_instance.to_dict()
# create an instance of GranularityResultItem from a dict
granularity_result_item_from_dict = GranularityResultItem.from_dict(granularity_result_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


