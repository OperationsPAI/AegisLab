# DetectorResultItem


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
from rcabench.openapi.models.detector_result_item import DetectorResultItem

# TODO update the JSON string below
json = "{}"
# create an instance of DetectorResultItem from a JSON string
detector_result_item_instance = DetectorResultItem.from_json(json)
# print the JSON string representation of the object
print(DetectorResultItem.to_json())

# convert the object into a dict
detector_result_item_dict = detector_result_item_instance.to_dict()
# create an instance of DetectorResultItem from a dict
detector_result_item_from_dict = DetectorResultItem.from_dict(detector_result_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


