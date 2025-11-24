# CreateDatasetVersionReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**checksum** | **str** |  | [optional] 
**data_source** | **str** |  | [optional] 
**download_url** | **str** |  | [optional] 
**format** | **str** |  | [optional] 
**name** | **str** |  | 

## Example

```python
from rcabench.openapi.models.create_dataset_version_req import CreateDatasetVersionReq

# TODO update the JSON string below
json = "{}"
# create an instance of CreateDatasetVersionReq from a JSON string
create_dataset_version_req_instance = CreateDatasetVersionReq.from_json(json)
# print the JSON string representation of the object
print(CreateDatasetVersionReq.to_json())

# convert the object into a dict
create_dataset_version_req_dict = create_dataset_version_req_instance.to_dict()
# create an instance of CreateDatasetVersionReq from a dict
create_dataset_version_req_from_dict = CreateDatasetVersionReq.from_dict(create_dataset_version_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


