# ListDatasetResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**items** | [**List[DatasetResp]**](DatasetResp.md) |  | [optional] 
**pagination** | [**PaginationInfo**](PaginationInfo.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.list_dataset_resp import ListDatasetResp

# TODO update the JSON string below
json = "{}"
# create an instance of ListDatasetResp from a JSON string
list_dataset_resp_instance = ListDatasetResp.from_json(json)
# print the JSON string representation of the object
print(ListDatasetResp.to_json())

# convert the object into a dict
list_dataset_resp_dict = list_dataset_resp_instance.to_dict()
# create an instance of ListDatasetResp from a dict
list_dataset_resp_from_dict = ListDatasetResp.from_dict(list_dataset_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


