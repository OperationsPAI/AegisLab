# CreateContainerVersionReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**command** | **str** |  | [optional] 
**env_vars** | [**List[CreateParameterConfigReq]**](CreateParameterConfigReq.md) |  | [optional] 
**github_link** | **str** |  | [optional] 
**helm_config** | [**CreateHelmConfigReq**](CreateHelmConfigReq.md) |  | [optional] 
**image_ref** | **str** |  | 
**name** | **str** |  | 

## Example

```python
from openapi.models.create_container_version_req import CreateContainerVersionReq

# TODO update the JSON string below
json = "{}"
# create an instance of CreateContainerVersionReq from a JSON string
create_container_version_req_instance = CreateContainerVersionReq.from_json(json)
# print the JSON string representation of the object
print(CreateContainerVersionReq.to_json())

# convert the object into a dict
create_container_version_req_dict = create_container_version_req_instance.to_dict()
# create an instance of CreateContainerVersionReq from a dict
create_container_version_req_from_dict = CreateContainerVersionReq.from_dict(create_container_version_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


