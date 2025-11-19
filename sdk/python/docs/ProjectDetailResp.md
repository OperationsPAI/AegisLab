# ProjectDetailResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**containers** | [**List[ContainerResp]**](ContainerResp.md) |  | [optional] 
**created_at** | **str** |  | [optional] 
**datapacks** | [**List[InjectionResp]**](InjectionResp.md) |  | [optional] 
**datasets** | [**List[DatasetResp]**](DatasetResp.md) |  | [optional] 
**description** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**is_public** | **bool** |  | [optional] 
**labels** | [**List[LabelItem]**](LabelItem.md) |  | [optional] 
**name** | **str** |  | [optional] 
**status** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 
**user_count** | **int** |  | [optional] 

## Example

```python
from openapi.models.project_detail_resp import ProjectDetailResp

# TODO update the JSON string below
json = "{}"
# create an instance of ProjectDetailResp from a JSON string
project_detail_resp_instance = ProjectDetailResp.from_json(json)
# print the JSON string representation of the object
print(ProjectDetailResp.to_json())

# convert the object into a dict
project_detail_resp_dict = project_detail_resp_instance.to_dict()
# create an instance of ProjectDetailResp from a dict
project_detail_resp_from_dict = ProjectDetailResp.from_dict(project_detail_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


