# DtoDatapackDetectorResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**found_count** | **int** | Number of datapacks with detector results | [optional] 
**items** | [**List[DtoDatapackDetectorItem]**](DtoDatapackDetectorItem.md) | Detector results for each datapack | [optional] 
**not_found_count** | **int** | Number of datapacks without detector results | [optional] 
**total_count** | **int** | Total number of requested datapacks | [optional] 

## Example

```python
from rcabench.openapi.models.dto_datapack_detector_resp import DtoDatapackDetectorResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatapackDetectorResp from a JSON string
dto_datapack_detector_resp_instance = DtoDatapackDetectorResp.from_json(json)
# print the JSON string representation of the object
print(DtoDatapackDetectorResp.to_json())

# convert the object into a dict
dto_datapack_detector_resp_dict = dto_datapack_detector_resp_instance.to_dict()
# create an instance of DtoDatapackDetectorResp from a dict
dto_datapack_detector_resp_from_dict = DtoDatapackDetectorResp.from_dict(dto_datapack_detector_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


