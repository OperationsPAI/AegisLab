# CreateContainerReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**is_public** | **bool** |  | [optional] 
**name** | **str** |  | 
**readme** | **str** |  | [optional] 
**type** | [**ContainerType**](ContainerType.md) |  | [optional] 
**version** | [**CreateContainerVersionReq**](CreateContainerVersionReq.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.create_container_req import CreateContainerReq

# TODO update the JSON string below
json = "{}"
# create an instance of CreateContainerReq from a JSON string
create_container_req_instance = CreateContainerReq.from_json(json)
# print the JSON string representation of the object
print(CreateContainerReq.to_json())

# convert the object into a dict
create_container_req_dict = create_container_req_instance.to_dict()
# create an instance of CreateContainerReq from a dict
create_container_req_from_dict = CreateContainerReq.from_dict(create_container_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


