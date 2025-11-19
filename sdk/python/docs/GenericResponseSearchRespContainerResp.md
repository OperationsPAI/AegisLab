# GenericResponseSearchRespContainerResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**SearchRespContainerResp**](SearchRespContainerResp.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from openapi.models.generic_response_search_resp_container_resp import GenericResponseSearchRespContainerResp

# TODO update the JSON string below
json = "{}"
# create an instance of GenericResponseSearchRespContainerResp from a JSON string
generic_response_search_resp_container_resp_instance = GenericResponseSearchRespContainerResp.from_json(json)
# print the JSON string representation of the object
print(GenericResponseSearchRespContainerResp.to_json())

# convert the object into a dict
generic_response_search_resp_container_resp_dict = generic_response_search_resp_container_resp_instance.to_dict()
# create an instance of GenericResponseSearchRespContainerResp from a dict
generic_response_search_resp_container_resp_from_dict = GenericResponseSearchRespContainerResp.from_dict(generic_response_search_resp_container_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


