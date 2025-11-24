# ContainerVersionResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **int** |  | [optional] 
**image_ref** | **str** |  | [optional] 
**name** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 
**usage** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.container_version_resp import ContainerVersionResp

# TODO update the JSON string below
json = "{}"
# create an instance of ContainerVersionResp from a JSON string
container_version_resp_instance = ContainerVersionResp.from_json(json)
# print the JSON string representation of the object
print(ContainerVersionResp.to_json())

# convert the object into a dict
container_version_resp_dict = container_version_resp_instance.to_dict()
# create an instance of ContainerVersionResp from a dict
container_version_resp_from_dict = ContainerVersionResp.from_dict(container_version_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


