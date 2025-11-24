# SubmitDatapackBuildingReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**labels** | [**List[LabelItem]**](LabelItem.md) |  | [optional] 
**project_name** | **str** |  | 
**specs** | [**List[BuildingSpec]**](BuildingSpec.md) |  | 

## Example

```python
from rcabench.openapi.models.submit_datapack_building_req import SubmitDatapackBuildingReq

# TODO update the JSON string below
json = "{}"
# create an instance of SubmitDatapackBuildingReq from a JSON string
submit_datapack_building_req_instance = SubmitDatapackBuildingReq.from_json(json)
# print the JSON string representation of the object
print(SubmitDatapackBuildingReq.to_json())

# convert the object into a dict
submit_datapack_building_req_dict = submit_datapack_building_req_instance.to_dict()
# create an instance of SubmitDatapackBuildingReq from a dict
submit_datapack_building_req_from_dict = SubmitDatapackBuildingReq.from_dict(submit_datapack_building_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


