# DtoSizeRangeFilter


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**max_size** | **int** | Maximum size (bytes) | [optional] 
**min_size** | **int** | Minimum size (bytes) | [optional] 

## Example

```python
from rcabench.openapi.models.dto_size_range_filter import DtoSizeRangeFilter

# TODO update the JSON string below
json = "{}"
# create an instance of DtoSizeRangeFilter from a JSON string
dto_size_range_filter_instance = DtoSizeRangeFilter.from_json(json)
# print the JSON string representation of the object
print(DtoSizeRangeFilter.to_json())

# convert the object into a dict
dto_size_range_filter_dict = dto_size_range_filter_instance.to_dict()
# create an instance of DtoSizeRangeFilter from a dict
dto_size_range_filter_from_dict = DtoSizeRangeFilter.from_dict(dto_size_range_filter_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


