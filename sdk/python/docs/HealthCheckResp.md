# HealthCheckResp


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
from openapi.models.health_check_resp import HealthCheckResp

# TODO update the JSON string below
json = "{}"
# create an instance of HealthCheckResp from a JSON string
health_check_resp_instance = HealthCheckResp.from_json(json)
# print the JSON string representation of the object
print(HealthCheckResp.to_json())

# convert the object into a dict
health_check_resp_dict = health_check_resp_instance.to_dict()
# create an instance of HealthCheckResp from a dict
health_check_resp_from_dict = HealthCheckResp.from_dict(health_check_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


