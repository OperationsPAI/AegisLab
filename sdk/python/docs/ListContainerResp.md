# ListContainerResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[ContainerResp]**](ContainerResp.md) |  | [optional] 
**pagination** | [**PaginationInfo**](PaginationInfo.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.list_container_resp import ListContainerResp

# TODO update the JSON string below
json = "{}"
# create an instance of ListContainerResp from a JSON string
list_container_resp_instance = ListContainerResp.from_json(json)
# print the JSON string representation of the object
print(ListContainerResp.to_json())

# convert the object into a dict
list_container_resp_dict = list_container_resp_instance.to_dict()
# create an instance of ListContainerResp from a dict
list_container_resp_from_dict = ListContainerResp.from_dict(list_container_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


