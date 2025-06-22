# DtoDatasetBuildPayload


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**benchmark** | **str** |  | [optional] 
**env_vars** | **object** |  | [optional] 
**name** | **str** |  | [optional] 
**pre_duration** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_build_payload import DtoDatasetBuildPayload

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetBuildPayload from a JSON string
dto_dataset_build_payload_instance = DtoDatasetBuildPayload.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetBuildPayload.to_json())

# convert the object into a dict
dto_dataset_build_payload_dict = dto_dataset_build_payload_instance.to_dict()
# create an instance of DtoDatasetBuildPayload from a dict
dto_dataset_build_payload_from_dict = DtoDatasetBuildPayload.from_dict(dto_dataset_build_payload_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


