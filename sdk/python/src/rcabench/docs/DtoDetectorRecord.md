# DtoDetectorRecord


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**abnormal_avg_duration** | **float** |  | [optional] 
**abnormal_p90** | **float** |  | [optional] 
**abnormal_p95** | **float** |  | [optional] 
**abnormal_p99** | **float** |  | [optional] 
**abnormal_succ_rate** | **float** |  | [optional] 
**issue** | **str** |  | [optional] 
**normal_avg_duration** | **float** |  | [optional] 
**normal_p90** | **float** |  | [optional] 
**normal_p95** | **float** |  | [optional] 
**normal_p99** | **float** |  | [optional] 
**normal_succ_rate** | **float** |  | [optional] 
**span_name** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_detector_record import DtoDetectorRecord

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDetectorRecord from a JSON string
dto_detector_record_instance = DtoDetectorRecord.from_json(json)
# print the JSON string representation of the object
print(DtoDetectorRecord.to_json())

# convert the object into a dict
dto_detector_record_dict = dto_detector_record_instance.to_dict()
# create an instance of DtoDetectorRecord from a dict
dto_detector_record_from_dict = DtoDetectorRecord.from_dict(dto_detector_record_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


