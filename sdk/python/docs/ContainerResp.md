# ContainerResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**is_public** | **bool** |  | [optional] 
**labels** | [**List[LabelItem]**](LabelItem.md) |  | [optional] 
**name** | **str** |  | [optional] 
**status** | **str** |  | [optional] 
**type** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.container_resp import ContainerResp

# TODO update the JSON string below
json = "{}"
# create an instance of ContainerResp from a JSON string
container_resp_instance = ContainerResp.from_json(json)
# print the JSON string representation of the object
print(ContainerResp.to_json())

# convert the object into a dict
container_resp_dict = container_resp_instance.to_dict()
# create an instance of ContainerResp from a dict
container_resp_from_dict = ContainerResp.from_dict(container_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


