# GenericResponseArrayLabelItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**List[LabelItem]**](LabelItem.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from openapi.models.generic_response_array_label_item import GenericResponseArrayLabelItem

# TODO update the JSON string below
json = "{}"
# create an instance of GenericResponseArrayLabelItem from a JSON string
generic_response_array_label_item_instance = GenericResponseArrayLabelItem.from_json(json)
# print the JSON string representation of the object
print(GenericResponseArrayLabelItem.to_json())

# convert the object into a dict
generic_response_array_label_item_dict = generic_response_array_label_item_instance.to_dict()
# create an instance of GenericResponseArrayLabelItem from a dict
generic_response_array_label_item_from_dict = GenericResponseArrayLabelItem.from_dict(generic_response_array_label_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


