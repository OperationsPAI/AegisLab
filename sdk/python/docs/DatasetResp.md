# DatasetResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**is_public** | **bool** |  | [optional] 
**labels** | [**List[LabelItem]**](LabelItem.md) |  | [optional] 
**name** | **str** |  | [optional] 
**status** | **str** |  | [optional] 
**type** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dataset_resp import DatasetResp

# TODO update the JSON string below
json = "{}"
# create an instance of DatasetResp from a JSON string
dataset_resp_instance = DatasetResp.from_json(json)
# print the JSON string representation of the object
print(DatasetResp.to_json())

# convert the object into a dict
dataset_resp_dict = dataset_resp_instance.to_dict()
# create an instance of DatasetResp from a dict
dataset_resp_from_dict = DatasetResp.from_dict(dataset_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


