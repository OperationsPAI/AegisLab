# DtoCreateContainerRequest

Container creation request for v2 API

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**build_options** | [**DtoBuildOptions**](DtoBuildOptions.md) | @Description Container build options | [optional] 
**build_source** | [**DtoBuildSource**](DtoBuildSource.md) | @Description Container build source configuration | [optional] 
**command** | **str** | @Description Container startup command @example /bin/bash | [optional] 
**env_vars** | **List[str]** | @Description Environment variables | [optional] 
**image** | **str** | @Description Docker image name @example my-image | 
**is_public** | **bool** | @Description Whether the container is public | [optional] 
**name** | **str** | @Description Container name @example my-container | 
**tag** | **str** | @Description Docker image tag @example latest | [optional] 
**type** | **str** | @Description Container type (algorithm or benchmark) @example algorithm | 

## Example

```python
from rcabench.openapi.models.dto_create_container_request import DtoCreateContainerRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoCreateContainerRequest from a JSON string
dto_create_container_request_instance = DtoCreateContainerRequest.from_json(json)
# print the JSON string representation of the object
print(DtoCreateContainerRequest.to_json())

# convert the object into a dict
dto_create_container_request_dict = dto_create_container_request_instance.to_dict()
# create an instance of DtoCreateContainerRequest from a dict
dto_create_container_request_from_dict = DtoCreateContainerRequest.from_dict(dto_create_container_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


