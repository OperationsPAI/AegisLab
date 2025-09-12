# DtoDatasetV2LabelManageReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**add_labels** | **List[int]** | List of label IDs to add | [optional] 
**new_labels** | [**List[DtoDatasetV2LabelCreateReq]**](DtoDatasetV2LabelCreateReq.md) | New label list | [optional] 
**remove_labels** | **List[int]** | List of label IDs to remove | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_v2_label_manage_req import DtoDatasetV2LabelManageReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetV2LabelManageReq from a JSON string
dto_dataset_v2_label_manage_req_instance = DtoDatasetV2LabelManageReq.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetV2LabelManageReq.to_json())

# convert the object into a dict
dto_dataset_v2_label_manage_req_dict = dto_dataset_v2_label_manage_req_instance.to_dict()
# create an instance of DtoDatasetV2LabelManageReq from a dict
dto_dataset_v2_label_manage_req_from_dict = DtoDatasetV2LabelManageReq.from_dict(dto_dataset_v2_label_manage_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


