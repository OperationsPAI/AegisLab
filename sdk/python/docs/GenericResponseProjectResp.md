# GenericResponseProjectResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**ProjectResp**](ProjectResp.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from openapi.models.generic_response_project_resp import GenericResponseProjectResp

# TODO update the JSON string below
json = "{}"
# create an instance of GenericResponseProjectResp from a JSON string
generic_response_project_resp_instance = GenericResponseProjectResp.from_json(json)
# print the JSON string representation of the object
print(GenericResponseProjectResp.to_json())

# convert the object into a dict
generic_response_project_resp_dict = generic_response_project_resp_instance.to_dict()
# create an instance of GenericResponseProjectResp from a dict
generic_response_project_resp_from_dict = GenericResponseProjectResp.from_dict(generic_response_project_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


