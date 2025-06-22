# DtoInjectionItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**end_time** | **str** |  | [optional] 
**fault_type** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**pre_duration** | **int** |  | [optional] 
**spec** | **object** |  | [optional] 
**start_time** | **str** |  | [optional] 
**status** | **str** |  | [optional] 
**task_id** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_item import DtoInjectionItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionItem from a JSON string
dto_injection_item_instance = DtoInjectionItem.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionItem.to_json())

# convert the object into a dict
dto_injection_item_dict = dto_injection_item_instance.to_dict()
# create an instance of DtoInjectionItem from a dict
dto_injection_item_from_dict = DtoInjectionItem.from_dict(dto_injection_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


