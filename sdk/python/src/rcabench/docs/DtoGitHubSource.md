# DtoGitHubSource

GitHub source configuration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**branch** | **str** | @Description Branch name (optional, defaults to main) | [optional] 
**commit** | **str** | @Description Specific commit hash (optional) | [optional] 
**path** | **str** | @Description Path within the repository (optional) | [optional] 
**repository** | **str** | @Description GitHub repository in format &#39;owner/repo&#39; | 
**token** | **str** | @Description GitHub access token (optional) | [optional] 

## Example

```python
from rcabench.openapi.models.dto_git_hub_source import DtoGitHubSource

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGitHubSource from a JSON string
dto_git_hub_source_instance = DtoGitHubSource.from_json(json)
# print the JSON string representation of the object
print(DtoGitHubSource.to_json())

# convert the object into a dict
dto_git_hub_source_dict = dto_git_hub_source_instance.to_dict()
# create an instance of DtoGitHubSource from a dict
dto_git_hub_source_from_dict = DtoGitHubSource.from_dict(dto_git_hub_source_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


