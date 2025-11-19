# ContainerVersionDetailResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**command** | **str** |  | [optional] 
**env_vars** | **str** |  | [optional] 
**github_link** | **str** |  | [optional] 
**helm_config** | [**HelmConfigDetailResp**](HelmConfigDetailResp.md) |  | [optional] 
**id** | **int** |  | [optional] 
**image_ref** | **str** |  | [optional] 
**name** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 
**usage** | **int** |  | [optional] 

## Example

```python
from openapi.models.container_version_detail_resp import ContainerVersionDetailResp

# TODO update the JSON string below
json = "{}"
# create an instance of ContainerVersionDetailResp from a JSON string
container_version_detail_resp_instance = ContainerVersionDetailResp.from_json(json)
# print the JSON string representation of the object
print(ContainerVersionDetailResp.to_json())

# convert the object into a dict
container_version_detail_resp_dict = container_version_detail_resp_instance.to_dict()
# create an instance of ContainerVersionDetailResp from a dict
container_version_detail_resp_from_dict = ContainerVersionDetailResp.from_dict(container_version_detail_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


