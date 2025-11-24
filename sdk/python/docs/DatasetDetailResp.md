# DatasetDetailResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**is_public** | **bool** |  | [optional] 
**labels** | [**List[LabelItem]**](LabelItem.md) |  | [optional] 
**name** | **str** |  | [optional] 
**status** | **str** |  | [optional] 
**type** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 
**versions** | [**List[DatasetVersionResp]**](DatasetVersionResp.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dataset_detail_resp import DatasetDetailResp

# TODO update the JSON string below
json = "{}"
# create an instance of DatasetDetailResp from a JSON string
dataset_detail_resp_instance = DatasetDetailResp.from_json(json)
# print the JSON string representation of the object
print(DatasetDetailResp.to_json())

# convert the object into a dict
dataset_detail_resp_dict = dataset_detail_resp_instance.to_dict()
# create an instance of DatasetDetailResp from a dict
dataset_detail_resp_from_dict = DatasetDetailResp.from_dict(dataset_detail_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


