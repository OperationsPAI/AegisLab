# DtoSystemStatisticsResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**containers** | [**DtoContainerStatistics**](DtoContainerStatistics.md) |  | [optional] 
**datasets** | [**DtoDatasetStatistics**](DtoDatasetStatistics.md) |  | [optional] 
**executions** | [**DtoExecutionStatistics**](DtoExecutionStatistics.md) |  | [optional] 
**generated_at** | **str** |  | [optional] 
**injections** | [**DtoInjectionStatistics**](DtoInjectionStatistics.md) |  | [optional] 
**projects** | [**DtoProjectStatistics**](DtoProjectStatistics.md) |  | [optional] 
**tasks** | [**DtoTaskStatistics**](DtoTaskStatistics.md) |  | [optional] 
**users** | [**DtoUserStatistics**](DtoUserStatistics.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_system_statistics_response import DtoSystemStatisticsResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoSystemStatisticsResponse from a JSON string
dto_system_statistics_response_instance = DtoSystemStatisticsResponse.from_json(json)
# print the JSON string representation of the object
print(DtoSystemStatisticsResponse.to_json())

# convert the object into a dict
dto_system_statistics_response_dict = dto_system_statistics_response_instance.to_dict()
# create an instance of DtoSystemStatisticsResponse from a dict
dto_system_statistics_response_from_dict = DtoSystemStatisticsResponse.from_dict(dto_system_statistics_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


