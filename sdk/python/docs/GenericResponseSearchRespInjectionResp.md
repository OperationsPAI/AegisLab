# GenericResponseSearchRespInjectionResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**SearchRespInjectionResp**](SearchRespInjectionResp.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.generic_response_search_resp_injection_resp import GenericResponseSearchRespInjectionResp

# TODO update the JSON string below
json = "{}"
# create an instance of GenericResponseSearchRespInjectionResp from a JSON string
generic_response_search_resp_injection_resp_instance = GenericResponseSearchRespInjectionResp.from_json(json)
# print the JSON string representation of the object
print(GenericResponseSearchRespInjectionResp.to_json())

# convert the object into a dict
generic_response_search_resp_injection_resp_dict = generic_response_search_resp_injection_resp_instance.to_dict()
# create an instance of GenericResponseSearchRespInjectionResp from a dict
generic_response_search_resp_injection_resp_from_dict = GenericResponseSearchRespInjectionResp.from_dict(generic_response_search_resp_injection_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


