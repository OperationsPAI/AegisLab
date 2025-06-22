# DtoTaskItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** |  | [optional] 
**id** | **str** |  | [optional] 
**payload** | **object** |  | [optional] 
**status** | **str** |  | [optional] 
**trace_id** | **str** |  | [optional] 
**type** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_task_item import DtoTaskItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoTaskItem from a JSON string
dto_task_item_instance = DtoTaskItem.from_json(json)
# print the JSON string representation of the object
print(DtoTaskItem.to_json())

# convert the object into a dict
dto_task_item_dict = dto_task_item_instance.to_dict()
# create an instance of DtoTaskItem from a dict
dto_task_item_from_dict = DtoTaskItem.from_dict(dto_task_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


