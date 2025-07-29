# DtoContainerResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**command** | **str** |  | [optional] 
**created_at** | **str** |  | [optional] 
**env_vars** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**image** | **str** |  | [optional] 
**is_public** | **bool** |  | [optional] 
**name** | **str** |  | [optional] 
**status** | **bool** |  | [optional] 
**tag** | **str** |  | [optional] 
**type** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 
**user** | [**DtoUserResponse**](DtoUserResponse.md) | Related entities (only included when specifically requested) | [optional] 
**user_id** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_container_response import DtoContainerResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoContainerResponse from a JSON string
dto_container_response_instance = DtoContainerResponse.from_json(json)
# print the JSON string representation of the object
print(DtoContainerResponse.to_json())

# convert the object into a dict
dto_container_response_dict = dto_container_response_instance.to_dict()
# create an instance of DtoContainerResponse from a dict
dto_container_response_from_dict = DtoContainerResponse.from_dict(dto_container_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


