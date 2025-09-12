# DtoTaskV2Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** |  | [optional] 
**id** | **str** |  | [optional] 
**status** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_task_v2_response import DtoTaskV2Response

# TODO update the JSON string below
json = "{}"
# create an instance of DtoTaskV2Response from a JSON string
dto_task_v2_response_instance = DtoTaskV2Response.from_json(json)
# print the JSON string representation of the object
print(DtoTaskV2Response.to_json())

# convert the object into a dict
dto_task_v2_response_dict = dto_task_v2_response_instance.to_dict()
# create an instance of DtoTaskV2Response from a dict
dto_task_v2_response_from_dict = DtoTaskV2Response.from_dict(dto_task_v2_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


