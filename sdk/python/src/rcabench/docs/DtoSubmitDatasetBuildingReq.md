# DtoSubmitDatasetBuildingReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**payloads** | [**List[DtoDatasetBuildPayload]**](DtoDatasetBuildPayload.md) |  | 
**project_name** | **str** |  | 

## Example

```python
from rcabench.openapi.models.dto_submit_dataset_building_req import DtoSubmitDatasetBuildingReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoSubmitDatasetBuildingReq from a JSON string
dto_submit_dataset_building_req_instance = DtoSubmitDatasetBuildingReq.from_json(json)
# print the JSON string representation of the object
print(DtoSubmitDatasetBuildingReq.to_json())

# convert the object into a dict
dto_submit_dataset_building_req_dict = dto_submit_dataset_building_req_instance.to_dict()
# create an instance of DtoSubmitDatasetBuildingReq from a dict
dto_submit_dataset_building_req_from_dict = DtoSubmitDatasetBuildingReq.from_dict(dto_submit_dataset_building_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


