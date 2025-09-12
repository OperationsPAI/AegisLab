# DtoInjectionItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**batch** | **str** |  | [optional] 
**benchmark** | **str** |  | [optional] 
**created_at** | **str** |  | [optional] 
**display_config** | **str** |  | [optional] 
**end_time** | **str** |  | [optional] 
**engine_config** | **str** |  | [optional] 
**env** | **str** |  | [optional] 
**fault_type** | **int** |  | [optional] 
**id** | **int** |  | [optional] 
**injection_name** | **str** |  | [optional] 
**pre_duration** | **int** |  | [optional] 
**start_time** | **str** |  | [optional] 
**status** | **int** |  | [optional] 
**tag** | **str** |  | [optional] 

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


