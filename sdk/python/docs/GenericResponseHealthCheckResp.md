# GenericResponseHealthCheckResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**HealthCheckResp**](HealthCheckResp.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from openapi.models.generic_response_health_check_resp import GenericResponseHealthCheckResp

# TODO update the JSON string below
json = "{}"
# create an instance of GenericResponseHealthCheckResp from a JSON string
generic_response_health_check_resp_instance = GenericResponseHealthCheckResp.from_json(json)
# print the JSON string representation of the object
print(GenericResponseHealthCheckResp.to_json())

# convert the object into a dict
generic_response_health_check_resp_dict = generic_response_health_check_resp_instance.to_dict()
# create an instance of GenericResponseHealthCheckResp from a dict
generic_response_health_check_resp_from_dict = GenericResponseHealthCheckResp.from_dict(generic_response_health_check_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


