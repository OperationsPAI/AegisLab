# DtoHarborSource

Harbor source configuration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**image** | **str** | @Description Harbor image name | 
**tag** | **str** | @Description Image tag (optional, defaults to latest) | [optional] 

## Example

```python
from rcabench.openapi.models.dto_harbor_source import DtoHarborSource

# TODO update the JSON string below
json = "{}"
# create an instance of DtoHarborSource from a JSON string
dto_harbor_source_instance = DtoHarborSource.from_json(json)
# print the JSON string representation of the object
print(DtoHarborSource.to_json())

# convert the object into a dict
dto_harbor_source_dict = dto_harbor_source_instance.to_dict()
# create an instance of DtoHarborSource from a dict
dto_harbor_source_from_dict = DtoHarborSource.from_dict(dto_harbor_source_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


