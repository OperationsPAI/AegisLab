# DatasetVersionResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **int** |  | [optional] 
**name** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from openapi.models.dataset_version_resp import DatasetVersionResp

# TODO update the JSON string below
json = "{}"
# create an instance of DatasetVersionResp from a JSON string
dataset_version_resp_instance = DatasetVersionResp.from_json(json)
# print the JSON string representation of the object
print(DatasetVersionResp.to_json())

# convert the object into a dict
dataset_version_resp_dict = dataset_version_resp_instance.to_dict()
# create an instance of DatasetVersionResp from a dict
dataset_version_resp_from_dict = DatasetVersionResp.from_dict(dataset_version_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


