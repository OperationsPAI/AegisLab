# DtoProjectResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**is_public** | **bool** |  | [optional] 
**members** | [**List[DtoUserProjectResponse]**](DtoUserProjectResponse.md) | Related entities (only included when specifically requested) | [optional] 
**name** | **str** |  | [optional] 
**status** | **int** |  | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_project_response import DtoProjectResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoProjectResponse from a JSON string
dto_project_response_instance = DtoProjectResponse.from_json(json)
# print the JSON string representation of the object
print(DtoProjectResponse.to_json())

# convert the object into a dict
dto_project_response_dict = dto_project_response_instance.to_dict()
# create an instance of DtoProjectResponse from a dict
dto_project_response_from_dict = DtoProjectResponse.from_dict(dto_project_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


