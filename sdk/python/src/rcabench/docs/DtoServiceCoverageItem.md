# DtoServiceCoverageItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**coverage** | **float** |  | [optional] 
**not_covered** | **List[str]** |  | [optional] 
**num** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_service_coverage_item import DtoServiceCoverageItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoServiceCoverageItem from a JSON string
dto_service_coverage_item_instance = DtoServiceCoverageItem.from_json(json)
# print the JSON string representation of the object
print(DtoServiceCoverageItem.to_json())

# convert the object into a dict
dto_service_coverage_item_dict = dto_service_coverage_item_instance.to_dict()
# create an instance of DtoServiceCoverageItem from a dict
dto_service_coverage_item_from_dict = DtoServiceCoverageItem.from_dict(dto_service_coverage_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


