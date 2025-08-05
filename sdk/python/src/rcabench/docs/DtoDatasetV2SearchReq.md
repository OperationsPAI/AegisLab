# DtoDatasetV2SearchReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**date_range** | [**DtoDateRangeFilter**](DtoDateRangeFilter.md) | Date range filter | [optional] 
**include** | **List[str]** | Included related data | [optional] 
**is_public** | **bool** | Whether public | [optional] 
**label_keys** | **List[str]** | Filter by label key | [optional] 
**label_values** | **List[str]** | Filter by label value | [optional] 
**page** | **int** | Page number | [optional] 
**search** | **str** | Search keywords | [optional] 
**size** | **int** | Page size | [optional] 
**size_range** | [**DtoSizeRangeFilter**](DtoSizeRangeFilter.md) | Size range filter | [optional] 
**sort_by** | **str** | Sort field | [optional] 
**sort_order** | **str** | Sort direction | [optional] 
**statuses** | **List[int]** | Status list | [optional] 
**types** | **List[str]** | Dataset type list | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_v2_search_req import DtoDatasetV2SearchReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetV2SearchReq from a JSON string
dto_dataset_v2_search_req_instance = DtoDatasetV2SearchReq.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetV2SearchReq.to_json())

# convert the object into a dict
dto_dataset_v2_search_req_dict = dto_dataset_v2_search_req_instance.to_dict()
# create an instance of DtoDatasetV2SearchReq from a dict
dto_dataset_v2_search_req_from_dict = DtoDatasetV2SearchReq.from_dict(dto_dataset_v2_search_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


