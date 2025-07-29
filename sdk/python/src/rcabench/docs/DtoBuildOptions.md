# DtoBuildOptions

Build options for container creation

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**build_args** | **object** | @Description Build arguments (optional) | [optional] 
**context_dir** | **str** | @Description Context directory for build (optional) | [optional] 
**dockerfile_path** | **str** | @Description Path to Dockerfile (optional, defaults to Dockerfile) | [optional] 
**force_rebuild** | **bool** | @Description Force rebuild even if image exists | [optional] 
**labels** | **object** | @Description Build labels (optional) | [optional] 
**target** | **str** | @Description Build target (optional) | [optional] 

## Example

```python
from rcabench.openapi.models.dto_build_options import DtoBuildOptions

# TODO update the JSON string below
json = "{}"
# create an instance of DtoBuildOptions from a JSON string
dto_build_options_instance = DtoBuildOptions.from_json(json)
# print the JSON string representation of the object
print(DtoBuildOptions.to_json())

# convert the object into a dict
dto_build_options_dict = dto_build_options_instance.to_dict()
# create an instance of DtoBuildOptions from a dict
dto_build_options_from_dict = DtoBuildOptions.from_dict(dto_build_options_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


