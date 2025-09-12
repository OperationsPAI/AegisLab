# DtoDatasetV2LabelCreateReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**category** | **str** | Label category | [optional] 
**color** | **str** | Label color (hex format) | [optional] 
**description** | **str** | Label description | [optional] 
**key** | **str** | Label key | 
**value** | **str** | Label value | 

## Example

```python
from rcabench.openapi.models.dto_dataset_v2_label_create_req import DtoDatasetV2LabelCreateReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetV2LabelCreateReq from a JSON string
dto_dataset_v2_label_create_req_instance = DtoDatasetV2LabelCreateReq.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetV2LabelCreateReq.to_json())

# convert the object into a dict
dto_dataset_v2_label_create_req_dict = dto_dataset_v2_label_create_req_instance.to_dict()
# create an instance of DtoDatasetV2LabelCreateReq from a dict
dto_dataset_v2_label_create_req_from_dict = DtoDatasetV2LabelCreateReq.from_dict(dto_dataset_v2_label_create_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


