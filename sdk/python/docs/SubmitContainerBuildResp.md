# SubmitContainerBuildResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**group_id** | **str** |  | [optional] 
**task_id** | **str** |  | [optional] 
**trace_id** | **str** |  | [optional] 

## Example

```python
from openapi.models.submit_container_build_resp import SubmitContainerBuildResp

# TODO update the JSON string below
json = "{}"
# create an instance of SubmitContainerBuildResp from a JSON string
submit_container_build_resp_instance = SubmitContainerBuildResp.from_json(json)
# print the JSON string representation of the object
print(SubmitContainerBuildResp.to_json())

# convert the object into a dict
submit_container_build_resp_dict = submit_container_build_resp_instance.to_dict()
# create an instance of SubmitContainerBuildResp from a dict
submit_container_build_resp_from_dict = SubmitContainerBuildResp.from_dict(submit_container_build_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


