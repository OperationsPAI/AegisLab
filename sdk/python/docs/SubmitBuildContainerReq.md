# SubmitBuildContainerReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**build_options** | [**BuildOptions**](BuildOptions.md) |  | [optional] 
**github_branch** | **str** |  | [optional] 
**github_commit** | **str** |  | [optional] 
**github_repository** | **str** | GitHub repository information | 
**github_token** | **str** |  | [optional] 
**image_name** | **str** | Container Meta | 
**sub_path** | **str** |  | [optional] 
**tag** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.submit_build_container_req import SubmitBuildContainerReq

# TODO update the JSON string below
json = "{}"
# create an instance of SubmitBuildContainerReq from a JSON string
submit_build_container_req_instance = SubmitBuildContainerReq.from_json(json)
# print the JSON string representation of the object
print(SubmitBuildContainerReq.to_json())

# convert the object into a dict
submit_build_container_req_dict = submit_build_container_req_instance.to_dict()
# create an instance of SubmitBuildContainerReq from a dict
submit_build_container_req_from_dict = SubmitBuildContainerReq.from_dict(submit_build_container_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


