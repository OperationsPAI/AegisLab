# DtoDateRangeFilter


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**end_time** | **str** | 结束时间 | [optional] 
**start_time** | **str** | 开始时间 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_date_range_filter import DtoDateRangeFilter

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDateRangeFilter from a JSON string
dto_date_range_filter_instance = DtoDateRangeFilter.from_json(json)
# print the JSON string representation of the object
print(DtoDateRangeFilter.to_json())

# convert the object into a dict
dto_date_range_filter_dict = dto_date_range_filter_instance.to_dict()
# create an instance of DtoDateRangeFilter from a dict
dto_date_range_filter_from_dict = DtoDateRangeFilter.from_dict(dto_date_range_filter_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


