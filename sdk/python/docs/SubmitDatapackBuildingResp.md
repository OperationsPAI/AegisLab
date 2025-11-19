# SubmitDatapackBuildingResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**group_id** | **str** |  | [optional] 
**items** | [**List[SubmitBuildingItem]**](SubmitBuildingItem.md) |  | [optional] 

## Example

```python
from openapi.models.submit_datapack_building_resp import SubmitDatapackBuildingResp

# TODO update the JSON string below
json = "{}"
# create an instance of SubmitDatapackBuildingResp from a JSON string
submit_datapack_building_resp_instance = SubmitDatapackBuildingResp.from_json(json)
# print the JSON string representation of the object
print(SubmitDatapackBuildingResp.to_json())

# convert the object into a dict
submit_datapack_building_resp_dict = submit_datapack_building_resp_instance.to_dict()
# create an instance of SubmitDatapackBuildingResp from a dict
submit_datapack_building_resp_from_dict = SubmitDatapackBuildingResp.from_dict(submit_datapack_building_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


