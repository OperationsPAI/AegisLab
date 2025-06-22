# DtoGranularityRecord


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**confidence** | **float** |  | [optional] 
**level** | **str** |  | [optional] 
**rank** | **int** |  | [optional] 
**result** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_granularity_record import DtoGranularityRecord

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGranularityRecord from a JSON string
dto_granularity_record_instance = DtoGranularityRecord.from_json(json)
# print the JSON string representation of the object
print(DtoGranularityRecord.to_json())

# convert the object into a dict
dto_granularity_record_dict = dto_granularity_record_instance.to_dict()
# create an instance of DtoGranularityRecord from a dict
dto_granularity_record_from_dict = DtoGranularityRecord.from_dict(dto_granularity_record_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


