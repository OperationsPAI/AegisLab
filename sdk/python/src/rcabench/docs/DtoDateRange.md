# DtoDateRange


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**var_from** | **str** |  | [optional] 
**to** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_date_range import DtoDateRange

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDateRange from a JSON string
dto_date_range_instance = DtoDateRange.from_json(json)
# print the JSON string representation of the object
print(DtoDateRange.to_json())

# convert the object into a dict
dto_date_range_dict = dto_date_range_instance.to_dict()
# create an instance of DtoDateRange from a dict
dto_date_range_from_dict = DtoDateRange.from_dict(dto_date_range_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


