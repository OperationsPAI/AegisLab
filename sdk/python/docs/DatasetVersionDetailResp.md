# DatasetVersionDetailResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**checksum** | **str** |  | [optional] 
**file_count** | **int** |  | [optional] 
**format** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**name** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dataset_version_detail_resp import DatasetVersionDetailResp

# TODO update the JSON string below
json = "{}"
# create an instance of DatasetVersionDetailResp from a JSON string
dataset_version_detail_resp_instance = DatasetVersionDetailResp.from_json(json)
# print the JSON string representation of the object
print(DatasetVersionDetailResp.to_json())

# convert the object into a dict
dataset_version_detail_resp_dict = dataset_version_detail_resp_instance.to_dict()
# create an instance of DatasetVersionDetailResp from a dict
dataset_version_detail_resp_from_dict = DatasetVersionDetailResp.from_dict(dataset_version_detail_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


