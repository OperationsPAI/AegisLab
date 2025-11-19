# GenericResponseInjectionMetadataResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**InjectionMetadataResp**](InjectionMetadataResp.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from openapi.models.generic_response_injection_metadata_resp import GenericResponseInjectionMetadataResp

# TODO update the JSON string below
json = "{}"
# create an instance of GenericResponseInjectionMetadataResp from a JSON string
generic_response_injection_metadata_resp_instance = GenericResponseInjectionMetadataResp.from_json(json)
# print the JSON string representation of the object
print(GenericResponseInjectionMetadataResp.to_json())

# convert the object into a dict
generic_response_injection_metadata_resp_dict = generic_response_injection_metadata_resp_instance.to_dict()
# create an instance of GenericResponseInjectionMetadataResp from a dict
generic_response_injection_metadata_resp_from_dict = GenericResponseInjectionMetadataResp.from_dict(generic_response_injection_metadata_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


