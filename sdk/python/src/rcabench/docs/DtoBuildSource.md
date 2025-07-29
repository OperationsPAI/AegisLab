# DtoBuildSource

Build source configuration with different source types

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**file** | [**DtoFileSource**](DtoFileSource.md) | @Description File source configuration (for file uploads) | [optional] 
**github** | [**DtoGitHubSource**](DtoGitHubSource.md) | @Description GitHub source configuration | [optional] 
**harbor** | [**DtoHarborSource**](DtoHarborSource.md) | @Description Harbor source configuration | [optional] 
**type** | **str** | @Description Build source type (file, github, or harbor) | 

## Example

```python
from rcabench.openapi.models.dto_build_source import DtoBuildSource

# TODO update the JSON string below
json = "{}"
# create an instance of DtoBuildSource from a JSON string
dto_build_source_instance = DtoBuildSource.from_json(json)
# print the JSON string representation of the object
print(DtoBuildSource.to_json())

# convert the object into a dict
dto_build_source_dict = dto_build_source_instance.to_dict()
# create an instance of DtoBuildSource from a dict
dto_build_source_from_dict = DtoBuildSource.from_dict(dto_build_source_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


