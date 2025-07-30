# DtoGenericResponseDtoMonitoringMetricsResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**DtoMonitoringMetricsResponse**](DtoMonitoringMetricsResponse.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_monitoring_metrics_response import DtoGenericResponseDtoMonitoringMetricsResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoMonitoringMetricsResponse from a JSON string
dto_generic_response_dto_monitoring_metrics_response_instance = DtoGenericResponseDtoMonitoringMetricsResponse.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoMonitoringMetricsResponse.to_json())

# convert the object into a dict
dto_generic_response_dto_monitoring_metrics_response_dict = dto_generic_response_dto_monitoring_metrics_response_instance.to_dict()
# create an instance of DtoGenericResponseDtoMonitoringMetricsResponse from a dict
dto_generic_response_dto_monitoring_metrics_response_from_dict = DtoGenericResponseDtoMonitoringMetricsResponse.from_dict(dto_generic_response_dto_monitoring_metrics_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


