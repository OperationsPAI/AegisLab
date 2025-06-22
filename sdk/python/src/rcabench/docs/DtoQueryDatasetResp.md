# DtoQueryDatasetResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**detector_result** | [**DtoDetectorRecord**](DtoDetectorRecord.md) |  | [optional] 
**end_time** | **str** |  | [optional] 
**execution_results** | [**List[DtoExecutionRecord]**](DtoExecutionRecord.md) |  | [optional] 
**name** | **str** |  | [optional] 
**param** | **List[object]** |  | [optional] 
**start_time** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_query_dataset_resp import DtoQueryDatasetResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoQueryDatasetResp from a JSON string
dto_query_dataset_resp_instance = DtoQueryDatasetResp.from_json(json)
# print the JSON string representation of the object
print(DtoQueryDatasetResp.to_json())

# convert the object into a dict
dto_query_dataset_resp_dict = dto_query_dataset_resp_instance.to_dict()
# create an instance of DtoQueryDatasetResp from a dict
dto_query_dataset_resp_from_dict = DtoQueryDatasetResp.from_dict(dto_query_dataset_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


