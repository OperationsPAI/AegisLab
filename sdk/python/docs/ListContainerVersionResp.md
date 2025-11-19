# ListContainerVersionResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[ContainerVersionResp]**](ContainerVersionResp.md) |  | [optional] 
**pagination** | [**PaginationInfo**](PaginationInfo.md) |  | [optional] 

## Example

```python
from openapi.models.list_container_version_resp import ListContainerVersionResp

# TODO update the JSON string below
json = "{}"
# create an instance of ListContainerVersionResp from a JSON string
list_container_version_resp_instance = ListContainerVersionResp.from_json(json)
# print the JSON string representation of the object
print(ListContainerVersionResp.to_json())

# convert the object into a dict
list_container_version_resp_dict = list_container_version_resp_instance.to_dict()
# create an instance of ListContainerVersionResp from a dict
list_container_version_resp_from_dict = ListContainerVersionResp.from_dict(list_container_version_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


