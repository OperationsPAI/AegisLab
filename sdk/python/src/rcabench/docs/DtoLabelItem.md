# DtoLabelItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**key** | **str** |  | 
**value** | **str** |  | 

## Example

```python
from rcabench.openapi.models.dto_label_item import DtoLabelItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoLabelItem from a JSON string
dto_label_item_instance = DtoLabelItem.from_json(json)
# print the JSON string representation of the object
print(DtoLabelItem.to_json())

# convert the object into a dict
dto_label_item_dict = dto_label_item_instance.to_dict()
# create an instance of DtoLabelItem from a dict
dto_label_item_from_dict = DtoLabelItem.from_dict(dto_label_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


