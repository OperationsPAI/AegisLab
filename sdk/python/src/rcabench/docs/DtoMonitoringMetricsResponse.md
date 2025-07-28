# DtoMonitoringMetricsResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**labels** | **Dict[str, str]** |  | [optional] 
**metrics** | [**Dict[str, DtoMetricValue]**](DtoMetricValue.md) |  | [optional] 
**timestamp** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_monitoring_metrics_response import DtoMonitoringMetricsResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoMonitoringMetricsResponse from a JSON string
dto_monitoring_metrics_response_instance = DtoMonitoringMetricsResponse.from_json(json)
# print the JSON string representation of the object
print(DtoMonitoringMetricsResponse.to_json())

# convert the object into a dict
dto_monitoring_metrics_response_dict = dto_monitoring_metrics_response_instance.to_dict()
# create an instance of DtoMonitoringMetricsResponse from a dict
dto_monitoring_metrics_response_from_dict = DtoMonitoringMetricsResponse.from_dict(dto_monitoring_metrics_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


