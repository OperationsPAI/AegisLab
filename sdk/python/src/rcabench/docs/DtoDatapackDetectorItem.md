# DtoDatapackDetectorItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**datapack** | **str** | Datapack name (from FaultInjectionSchedule) | [optional] 
**executed_at** | **str** | Execution time | [optional] 
**execution_id** | **int** | Execution ID (0 if no execution found) | [optional] 
**found** | **bool** | Whether detector result was found | [optional] 
**results** | [**List[DtoDetectorRecord]**](DtoDetectorRecord.md) | Detector analysis results | [optional] 

## Example

```python
from rcabench.openapi.models.dto_datapack_detector_item import DtoDatapackDetectorItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatapackDetectorItem from a JSON string
dto_datapack_detector_item_instance = DtoDatapackDetectorItem.from_json(json)
# print the JSON string representation of the object
print(DtoDatapackDetectorItem.to_json())

# convert the object into a dict
dto_datapack_detector_item_dict = dto_datapack_detector_item_instance.to_dict()
# create an instance of DtoDatapackDetectorItem from a dict
dto_datapack_detector_item_from_dict = DtoDatapackDetectorItem.from_dict(dto_datapack_detector_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


