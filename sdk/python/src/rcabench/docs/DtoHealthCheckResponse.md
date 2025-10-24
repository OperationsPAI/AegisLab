# DtoHealthCheckResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**services** | **object** |  | [optional] 
**status** | **str** |  | [optional] 
**timestamp** | **str** |  | [optional] 
**uptime** | **str** |  | [optional] 
**version** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_health_check_response import DtoHealthCheckResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoHealthCheckResponse from a JSON string
dto_health_check_response_instance = DtoHealthCheckResponse.from_json(json)
# print the JSON string representation of the object
print(DtoHealthCheckResponse.to_json())

# convert the object into a dict
dto_health_check_response_dict = dto_health_check_response_instance.to_dict()
# create an instance of DtoHealthCheckResponse from a dict
dto_health_check_response_from_dict = DtoHealthCheckResponse.from_dict(dto_health_check_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


