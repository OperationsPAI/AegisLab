# GenericResponseSubmitDatapackBuildingResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**SubmitDatapackBuildingResp**](SubmitDatapackBuildingResp.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.generic_response_submit_datapack_building_resp import GenericResponseSubmitDatapackBuildingResp

# TODO update the JSON string below
json = "{}"
# create an instance of GenericResponseSubmitDatapackBuildingResp from a JSON string
generic_response_submit_datapack_building_resp_instance = GenericResponseSubmitDatapackBuildingResp.from_json(json)
# print the JSON string representation of the object
print(GenericResponseSubmitDatapackBuildingResp.to_json())

# convert the object into a dict
generic_response_submit_datapack_building_resp_dict = generic_response_submit_datapack_building_resp_instance.to_dict()
# create an instance of GenericResponseSubmitDatapackBuildingResp from a dict
generic_response_submit_datapack_building_resp_from_dict = GenericResponseSubmitDatapackBuildingResp.from_dict(generic_response_submit_datapack_building_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


