# InjectionMetadataResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**config** | [**ChaosNode**](ChaosNode.md) |  | [optional] 
**fault_resource_map** | [**Dict[str, ChaosResourceField]**](ChaosResourceField.md) |  | [optional] 
**fault_type_map** | **Dict[str, str]** |  | [optional] 
**ns_resources** | [**ChaosResources**](ChaosResources.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.injection_metadata_resp import InjectionMetadataResp

# TODO update the JSON string below
json = "{}"
# create an instance of InjectionMetadataResp from a JSON string
injection_metadata_resp_instance = InjectionMetadataResp.from_json(json)
# print the JSON string representation of the object
print(InjectionMetadataResp.to_json())

# convert the object into a dict
injection_metadata_resp_dict = injection_metadata_resp_instance.to_dict()
# create an instance of InjectionMetadataResp from a dict
injection_metadata_resp_from_dict = InjectionMetadataResp.from_dict(injection_metadata_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


