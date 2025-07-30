# DtoInjectionV2CreateItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**benchmark** | **str** |  | 
**description** | **str** |  | [optional] 
**display_config** | **str** |  | 
**end_time** | **str** |  | [optional] 
**engine_config** | **str** |  | 
**fault_type** | **int** |  | 
**injection_name** | **str** |  | 
**pre_duration** | **int** |  | 
**start_time** | **str** |  | [optional] 
**status** | **int** |  | [optional] 
**task_id** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_v2_create_item import DtoInjectionV2CreateItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionV2CreateItem from a JSON string
dto_injection_v2_create_item_instance = DtoInjectionV2CreateItem.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionV2CreateItem.to_json())

# convert the object into a dict
dto_injection_v2_create_item_dict = dto_injection_v2_create_item_instance.to_dict()
# create an instance of DtoInjectionV2CreateItem from a dict
dto_injection_v2_create_item_from_dict = DtoInjectionV2CreateItem.from_dict(dto_injection_v2_create_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


