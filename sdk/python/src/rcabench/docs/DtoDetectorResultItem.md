# DtoDetectorResultItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**abnormal_avg_duration** | **float** |  | [optional] 
**abnormal_p90** | **float** |  | [optional] 
**abnormal_p95** | **float** |  | [optional] 
**abnormal_p99** | **float** |  | [optional] 
**abnormal_succ_rate** | **float** |  | [optional] 
**issues** | **str** |  | 
**normal_avg_duration** | **float** |  | [optional] 
**normal_p90** | **float** |  | [optional] 
**normal_p95** | **float** |  | [optional] 
**normal_p99** | **float** |  | [optional] 
**normal_succ_rate** | **float** |  | [optional] 
**span_name** | **str** |  | 

## Example

```python
from rcabench.openapi.models.dto_detector_result_item import DtoDetectorResultItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDetectorResultItem from a JSON string
dto_detector_result_item_instance = DtoDetectorResultItem.from_json(json)
# print the JSON string representation of the object
print(DtoDetectorResultItem.to_json())

# convert the object into a dict
dto_detector_result_item_dict = dto_detector_result_item_instance.to_dict()
# create an instance of DtoDetectorResultItem from a dict
dto_detector_result_item_from_dict = DtoDetectorResultItem.from_dict(dto_detector_result_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


