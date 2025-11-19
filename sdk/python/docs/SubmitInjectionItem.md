# SubmitInjectionItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**index** | **int** |  | [optional] 
**task_id** | **str** |  | [optional] 
**trace_id** | **str** |  | [optional] 

## Example

```python
from openapi.models.submit_injection_item import SubmitInjectionItem

# TODO update the JSON string below
json = "{}"
# create an instance of SubmitInjectionItem from a JSON string
submit_injection_item_instance = SubmitInjectionItem.from_json(json)
# print the JSON string representation of the object
print(SubmitInjectionItem.to_json())

# convert the object into a dict
submit_injection_item_dict = submit_injection_item_instance.to_dict()
# create an instance of SubmitInjectionItem from a dict
submit_injection_item_from_dict = SubmitInjectionItem.from_dict(submit_injection_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


