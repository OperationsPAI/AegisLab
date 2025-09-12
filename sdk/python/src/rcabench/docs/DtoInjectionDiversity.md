# DtoInjectionDiversity


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**attribute_coverages** | **Dict[str, Dict[str, DtoAttributeCoverageItem]]** |  | [optional] 
**fault_distribution** | **Dict[str, int]** |  | [optional] 
**fault_service_coverages** | [**Dict[str, DtoServiceCoverageItem]**](DtoServiceCoverageItem.md) |  | [optional] 
**pair_distribution** | [**List[DtoPairStats]**](DtoPairStats.md) |  | [optional] 
**service_distribution** | **Dict[str, int]** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_injection_diversity import DtoInjectionDiversity

# TODO update the JSON string below
json = "{}"
# create an instance of DtoInjectionDiversity from a JSON string
dto_injection_diversity_instance = DtoInjectionDiversity.from_json(json)
# print the JSON string representation of the object
print(DtoInjectionDiversity.to_json())

# convert the object into a dict
dto_injection_diversity_dict = dto_injection_diversity_instance.to_dict()
# create an instance of DtoInjectionDiversity from a dict
dto_injection_diversity_from_dict = DtoInjectionDiversity.from_dict(dto_injection_diversity_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


