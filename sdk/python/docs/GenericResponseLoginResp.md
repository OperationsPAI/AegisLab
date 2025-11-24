# GenericResponseLoginResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**LoginResp**](LoginResp.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.generic_response_login_resp import GenericResponseLoginResp

# TODO update the JSON string below
json = "{}"
# create an instance of GenericResponseLoginResp from a JSON string
generic_response_login_resp_instance = GenericResponseLoginResp.from_json(json)
# print the JSON string representation of the object
print(GenericResponseLoginResp.to_json())

# convert the object into a dict
generic_response_login_resp_dict = generic_response_login_resp_instance.to_dict()
# create an instance of GenericResponseLoginResp from a dict
generic_response_login_resp_from_dict = GenericResponseLoginResp.from_dict(generic_response_login_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


