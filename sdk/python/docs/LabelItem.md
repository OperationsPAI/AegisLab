# LabelItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**key** | **str** |  | [optional] 
**value** | **str** |  | [optional] 

## Example

```python
from openapi.models.label_item import LabelItem

# TODO update the JSON string below
json = "{}"
# create an instance of LabelItem from a JSON string
label_item_instance = LabelItem.from_json(json)
# print the JSON string representation of the object
print(LabelItem.to_json())

# convert the object into a dict
label_item_dict = label_item_instance.to_dict()
# create an instance of LabelItem from a dict
label_item_from_dict = LabelItem.from_dict(label_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


