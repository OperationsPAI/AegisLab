# DtoExecution


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**dataset** | [**DtoDatasetItem**](DtoDatasetItem.md) |  | [optional] 
**granularity_records** | [**List[DtoGranularityRecord]**](DtoGranularityRecord.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_execution import DtoExecution

# TODO update the JSON string below
json = "{}"
# create an instance of DtoExecution from a JSON string
dto_execution_instance = DtoExecution.from_json(json)
# print the JSON string representation of the object
print(DtoExecution.to_json())

# convert the object into a dict
dto_execution_dict = dto_execution_instance.to_dict()
# create an instance of DtoExecution from a dict
dto_execution_from_dict = DtoExecution.from_dict(dto_execution_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


