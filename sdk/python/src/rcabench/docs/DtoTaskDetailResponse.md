# DtoTaskDetailResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** |  | [optional] 
**group_id** | **str** |  | [optional] 
**id** | **str** |  | [optional] 
**immediate** | **bool** |  | [optional] 
**logs** | **List[str]** |  | [optional] 
**project** | [**DtoProjectResponse**](DtoProjectResponse.md) | Related entities (only included when specifically requested) | [optional] 
**project_id** | **int** |  | [optional] 
**status** | **str** |  | [optional] 
**trace_id** | **str** |  | [optional] 
**type** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_task_detail_response import DtoTaskDetailResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoTaskDetailResponse from a JSON string
dto_task_detail_response_instance = DtoTaskDetailResponse.from_json(json)
# print the JSON string representation of the object
print(DtoTaskDetailResponse.to_json())

# convert the object into a dict
dto_task_detail_response_dict = dto_task_detail_response_instance.to_dict()
# create an instance of DtoTaskDetailResponse from a dict
dto_task_detail_response_from_dict = DtoTaskDetailResponse.from_dict(dto_task_detail_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


