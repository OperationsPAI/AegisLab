# BuildOptions


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**build_args** | **object** |  | [optional] 
**context_dir** | **str** |  | [optional] [default to '.']
**dockerfile_path** | **str** |  | [optional] [default to 'Dockerfile']
**force_rebuild** | **bool** |  | [optional] 
**target** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.build_options import BuildOptions

# TODO update the JSON string below
json = "{}"
# create an instance of BuildOptions from a JSON string
build_options_instance = BuildOptions.from_json(json)
# print the JSON string representation of the object
print(BuildOptions.to_json())

# convert the object into a dict
build_options_dict = build_options_instance.to_dict()
# create an instance of BuildOptions from a dict
build_options_from_dict = BuildOptions.from_dict(build_options_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


