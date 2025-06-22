# DtoExecutionRecord


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | **str** |  | [optional] 
**granularity_records** | [**List[DtoGranularityRecord]**](DtoGranularityRecord.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_execution_record import DtoExecutionRecord

# TODO update the JSON string below
json = "{}"
# create an instance of DtoExecutionRecord from a JSON string
dto_execution_record_instance = DtoExecutionRecord.from_json(json)
# print the JSON string representation of the object
print(DtoExecutionRecord.to_json())

# convert the object into a dict
dto_execution_record_dict = dto_execution_record_instance.to_dict()
# create an instance of DtoExecutionRecord from a dict
dto_execution_record_from_dict = DtoExecutionRecord.from_dict(dto_execution_record_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


