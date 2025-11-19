# SortOption


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**direction** | [**SortDirection**](SortDirection.md) | Sort direction | 
**var_field** | **str** | Sort field | 

## Example

```python
from openapi.models.sort_option import SortOption

# TODO update the JSON string below
json = "{}"
# create an instance of SortOption from a JSON string
sort_option_instance = SortOption.from_json(json)
# print the JSON string representation of the object
print(SortOption.to_json())

# convert the object into a dict
sort_option_dict = sort_option_instance.to_dict()
# create an instance of SortOption from a dict
sort_option_from_dict = SortOption.from_dict(sort_option_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


