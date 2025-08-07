# DtoProjectV2Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**containers** | [**List[DatabaseContainer]**](DatabaseContainer.md) | Associated containers | [optional] 
**created_at** | **str** | Creation time | [optional] 
**datasets** | [**List[DatabaseDataset]**](DatabaseDataset.md) | Associated datasets | [optional] 
**description** | **str** | Project description | [optional] 
**id** | **int** | Unique identifier | [optional] 
**injections** | [**List[DtoInjectionV2Response]**](DtoInjectionV2Response.md) | Associated fault injections | [optional] 
**is_public** | **bool** | Whether public | [optional] 
**labels** | [**List[DatabaseLabel]**](DatabaseLabel.md) | Associated labels | [optional] 
**name** | **str** | Project name | [optional] 
**status** | **int** | Status | [optional] 
**updated_at** | **str** | Update time | [optional] 

## Example

```python
from rcabench.openapi.models.dto_project_v2_response import DtoProjectV2Response

# TODO update the JSON string below
json = "{}"
# create an instance of DtoProjectV2Response from a JSON string
dto_project_v2_response_instance = DtoProjectV2Response.from_json(json)
# print the JSON string representation of the object
print(DtoProjectV2Response.to_json())

# convert the object into a dict
dto_project_v2_response_dict = dto_project_v2_response_instance.to_dict()
# create an instance of DtoProjectV2Response from a dict
dto_project_v2_response_from_dict = DtoProjectV2Response.from_dict(dto_project_v2_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


