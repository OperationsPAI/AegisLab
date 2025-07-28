# DtoTimeRangeQuery


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**custom_end_time** | **str** |  | [optional] 
**custom_start_time** | **str** |  | [optional] 
**lookback** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_time_range_query import DtoTimeRangeQuery

# TODO update the JSON string below
json = "{}"
# create an instance of DtoTimeRangeQuery from a JSON string
dto_time_range_query_instance = DtoTimeRangeQuery.from_json(json)
# print the JSON string representation of the object
print(DtoTimeRangeQuery.to_json())

# convert the object into a dict
dto_time_range_query_dict = dto_time_range_query_instance.to_dict()
# create an instance of DtoTimeRangeQuery from a dict
dto_time_range_query_from_dict = DtoTimeRangeQuery.from_dict(dto_time_range_query_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


