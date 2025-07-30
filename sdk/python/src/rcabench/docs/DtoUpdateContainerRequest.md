# DtoUpdateContainerRequest

Request structure for updating container information

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**command** | **str** | @Description Container startup command (optional) | [optional] 
**env_vars** | **str** | @Description Environment variables (optional) | [optional] 
**image** | **str** | @Description Docker image name (optional) | [optional] 
**is_public** | **bool** | @Description Whether the container is public (optional) | [optional] 
**name** | **str** | @Description Container name (optional) | [optional] 
**status** | **bool** | @Description Container status (optional) | [optional] 
**tag** | **str** | @Description Docker image tag (optional) | [optional] 
**type** | **str** | @Description Container type (optional) | [optional] 

## Example

```python
from rcabench.openapi.models.dto_update_container_request import DtoUpdateContainerRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoUpdateContainerRequest from a JSON string
dto_update_container_request_instance = DtoUpdateContainerRequest.from_json(json)
# print the JSON string representation of the object
print(DtoUpdateContainerRequest.to_json())

# convert the object into a dict
dto_update_container_request_dict = dto_update_container_request_instance.to_dict()
# create an instance of DtoUpdateContainerRequest from a dict
dto_update_container_request_from_dict = DtoUpdateContainerRequest.from_dict(dto_update_container_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


