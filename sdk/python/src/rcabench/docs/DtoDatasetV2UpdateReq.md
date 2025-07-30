# DtoDatasetV2UpdateReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**data_source** | **str** | Data source description | [optional] 
**description** | **str** | Dataset description | [optional] 
**format** | **str** | Data format | [optional] 
**injection_ids** | **List[int]** | Update associated fault injection ID list (complete replacement) | [optional] 
**is_public** | **bool** | Whether public | [optional] 
**label_ids** | **List[int]** | Update associated label ID list (complete replacement) | [optional] 
**name** | **str** | Dataset name | [optional] 
**new_labels** | [**List[DtoDatasetV2LabelCreateReq]**](DtoDatasetV2LabelCreateReq.md) | New label list | [optional] 
**type** | **str** | Dataset type | [optional] 
**version** | **str** | Dataset version | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_v2_update_req import DtoDatasetV2UpdateReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetV2UpdateReq from a JSON string
dto_dataset_v2_update_req_instance = DtoDatasetV2UpdateReq.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetV2UpdateReq.to_json())

# convert the object into a dict
dto_dataset_v2_update_req_dict = dto_dataset_v2_update_req_instance.to_dict()
# create an instance of DtoDatasetV2UpdateReq from a dict
dto_dataset_v2_update_req_from_dict = DtoDatasetV2UpdateReq.from_dict(dto_dataset_v2_update_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


