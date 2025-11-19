# ProjectResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**is_public** | **bool** |  | [optional] 
**labels** | [**List[LabelItem]**](LabelItem.md) |  | [optional] 
**name** | **str** |  | [optional] 
**status** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from openapi.models.project_resp import ProjectResp

# TODO update the JSON string below
json = "{}"
# create an instance of ProjectResp from a JSON string
project_resp_instance = ProjectResp.from_json(json)
# print the JSON string representation of the object
print(ProjectResp.to_json())

# convert the object into a dict
project_resp_dict = project_resp_instance.to_dict()
# create an instance of ProjectResp from a dict
project_resp_from_dict = ProjectResp.from_dict(project_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


