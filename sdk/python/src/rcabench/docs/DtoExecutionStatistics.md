# DtoExecutionStatistics


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**failed** | **int** |  | [optional] 
**successful** | **int** |  | [optional] 
**total** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_execution_statistics import DtoExecutionStatistics

# TODO update the JSON string below
json = "{}"
# create an instance of DtoExecutionStatistics from a JSON string
dto_execution_statistics_instance = DtoExecutionStatistics.from_json(json)
# print the JSON string representation of the object
print(DtoExecutionStatistics.to_json())

# convert the object into a dict
dto_execution_statistics_dict = dto_execution_statistics_instance.to_dict()
# create an instance of DtoExecutionStatistics from a dict
dto_execution_statistics_from_dict = DtoExecutionStatistics.from_dict(dto_execution_statistics_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


