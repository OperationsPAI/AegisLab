# DtoAttributeCoverageItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**coverage** | **float** |  | [optional] 
**num** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_attribute_coverage_item import DtoAttributeCoverageItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAttributeCoverageItem from a JSON string
dto_attribute_coverage_item_instance = DtoAttributeCoverageItem.from_json(json)
# print the JSON string representation of the object
print(DtoAttributeCoverageItem.to_json())

# convert the object into a dict
dto_attribute_coverage_item_dict = dto_attribute_coverage_item_instance.to_dict()
# create an instance of DtoAttributeCoverageItem from a dict
dto_attribute_coverage_item_from_dict = DtoAttributeCoverageItem.from_dict(dto_attribute_coverage_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


