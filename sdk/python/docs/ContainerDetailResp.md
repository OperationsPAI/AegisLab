# ContainerDetailResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**is_public** | **bool** |  | [optional] 
**labels** | [**List[LabelItem]**](LabelItem.md) |  | [optional] 
**name** | **str** |  | [optional] 
**readme** | **str** |  | [optional] 
**status** | **str** |  | [optional] 
**type** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 
**versions** | [**List[ContainerVersionResp]**](ContainerVersionResp.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.container_detail_resp import ContainerDetailResp

# TODO update the JSON string below
json = "{}"
# create an instance of ContainerDetailResp from a JSON string
container_detail_resp_instance = ContainerDetailResp.from_json(json)
# print the JSON string representation of the object
print(ContainerDetailResp.to_json())

# convert the object into a dict
container_detail_resp_dict = container_detail_resp_instance.to_dict()
# create an instance of ContainerDetailResp from a dict
container_detail_resp_from_dict = ContainerDetailResp.from_dict(container_detail_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


