# DtoDatasetV2CreateReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**data_source** | **str** | Data source description | [optional] 
**description** | **str** | Dataset description | [optional] 
**format** | **str** | Data format | [optional] 
**injection_ids** | **List[int]** | Associated fault injection ID list | [optional] 
**is_public** | **bool** | Whether public, optional, defaults to false | [optional] 
**label_ids** | **List[int]** | Associated label ID list | [optional] 
**name** | **str** | Dataset name | 
**new_labels** | [**List[DtoDatasetV2LabelCreateReq]**](DtoDatasetV2LabelCreateReq.md) | New label list | [optional] 
**project_id** | **int** | Project ID | 
**type** | **str** | Dataset type | 
**version** | **str** | Dataset version, optional, defaults to v1.0 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_v2_create_req import DtoDatasetV2CreateReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetV2CreateReq from a JSON string
dto_dataset_v2_create_req_instance = DtoDatasetV2CreateReq.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetV2CreateReq.to_json())

# convert the object into a dict
dto_dataset_v2_create_req_dict = dto_dataset_v2_create_req_instance.to_dict()
# create an instance of DtoDatasetV2CreateReq from a dict
dto_dataset_v2_create_req_from_dict = DtoDatasetV2CreateReq.from_dict(dto_dataset_v2_create_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


