# DtoMonitoringQueryRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**end_time** | **str** |  | [optional] 
**query** | **str** |  | 
**start_time** | **str** |  | [optional] 
**step** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_monitoring_query_request import DtoMonitoringQueryRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoMonitoringQueryRequest from a JSON string
dto_monitoring_query_request_instance = DtoMonitoringQueryRequest.from_json(json)
# print the JSON string representation of the object
print(DtoMonitoringQueryRequest.to_json())

# convert the object into a dict
dto_monitoring_query_request_dict = dto_monitoring_query_request_instance.to_dict()
# create an instance of DtoMonitoringQueryRequest from a dict
dto_monitoring_query_request_from_dict = DtoMonitoringQueryRequest.from_dict(dto_monitoring_query_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


