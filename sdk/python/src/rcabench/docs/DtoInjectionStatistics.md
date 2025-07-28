# DtoInjectionStatistics


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**completed** | **int** |  | [optional] 
**failed** | **int** |  | [optional] 
**running** | **int** |  | [optional] 
**scheduled** | **int** |  | [optional] 
**total** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_statistics import DtoInjectionStatistics

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionStatistics from a JSON string
dto_injection_statistics_instance = DtoInjectionStatistics.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionStatistics.to_json())

# convert the object into a dict
dto_injection_statistics_dict = dto_injection_statistics_instance.to_dict()
# create an instance of DtoInjectionStatistics from a dict
dto_injection_statistics_from_dict = DtoInjectionStatistics.from_dict(dto_injection_statistics_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


