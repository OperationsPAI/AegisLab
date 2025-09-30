# DtoInjectionV2CascadeDeleteStats


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**dataset_fault_injections** | **int** |  | [optional] 
**detectors** | **int** |  | [optional] 
**execution_result_labels** | **int** |  | [optional] 
**execution_results** | **int** |  | [optional] 
**fault_injection_labels** | **int** |  | [optional] 
**granularity_results** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_v2_cascade_delete_stats import DtoInjectionV2CascadeDeleteStats

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionV2CascadeDeleteStats from a JSON string
dto_injection_v2_cascade_delete_stats_instance = DtoInjectionV2CascadeDeleteStats.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionV2CascadeDeleteStats.to_json())

# convert the object into a dict
dto_injection_v2_cascade_delete_stats_dict = dto_injection_v2_cascade_delete_stats_instance.to_dict()
# create an instance of DtoInjectionV2CascadeDeleteStats from a dict
dto_injection_v2_cascade_delete_stats_from_dict = DtoInjectionV2CascadeDeleteStats.from_dict(dto_injection_v2_cascade_delete_stats_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


