# DtoSearchFilter


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**var_field** | **str** | Field name | 
**operator** | [**DtoFilterOperator**](DtoFilterOperator.md) | Operator | 
**value** | **object** | Value | [optional] 
**values** | **List[object]** | Multiple values (for IN operations etc.) | [optional] 

## Example

```python
from rcabench.openapi.models.dto_search_filter import DtoSearchFilter

# TODO update the JSON string below
json = "{}"
# create an instance of DtoSearchFilter from a JSON string
dto_search_filter_instance = DtoSearchFilter.from_json(json)
# print the JSON string representation of the object
print(DtoSearchFilter.to_json())

# convert the object into a dict
dto_search_filter_dict = dto_search_filter_instance.to_dict()
# create an instance of DtoSearchFilter from a dict
dto_search_filter_from_dict = DtoSearchFilter.from_dict(dto_search_filter_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


