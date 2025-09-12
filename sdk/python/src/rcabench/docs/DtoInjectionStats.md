# DtoInjectionStats


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**diversity** | [**DtoInjectionDiversity**](DtoInjectionDiversity.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_stats import DtoInjectionStats

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionStats from a JSON string
dto_injection_stats_instance = DtoInjectionStats.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionStats.to_json())

# convert the object into a dict
dto_injection_stats_dict = dto_injection_stats_instance.to_dict()
# create an instance of DtoInjectionStats from a dict
dto_injection_stats_from_dict = DtoInjectionStats.from_dict(dto_injection_stats_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


