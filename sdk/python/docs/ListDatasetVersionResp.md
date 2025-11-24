# ListDatasetVersionResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DatasetVersionResp]**](DatasetVersionResp.md) |  | [optional] 
**pagination** | [**PaginationInfo**](PaginationInfo.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.list_dataset_version_resp import ListDatasetVersionResp

# TODO update the JSON string below
json = "{}"
# create an instance of ListDatasetVersionResp from a JSON string
list_dataset_version_resp_instance = ListDatasetVersionResp.from_json(json)
# print the JSON string representation of the object
print(ListDatasetVersionResp.to_json())

# convert the object into a dict
list_dataset_version_resp_dict = list_dataset_version_resp_instance.to_dict()
# create an instance of ListDatasetVersionResp from a dict
list_dataset_version_resp_from_dict = ListDatasetVersionResp.from_dict(list_dataset_version_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


