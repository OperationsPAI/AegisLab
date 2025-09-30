# DtoDetectorResultRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**duration** | **float** | Execution duration in seconds | 
**results** | [**List[DtoDetectorResultItem]**](DtoDetectorResultItem.md) |  | 

## Example

```python
from rcabench.openapi.models.dto_detector_result_request import DtoDetectorResultRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDetectorResultRequest from a JSON string
dto_detector_result_request_instance = DtoDetectorResultRequest.from_json(json)
# print the JSON string representation of the object
print(DtoDetectorResultRequest.to_json())

# convert the object into a dict
dto_detector_result_request_dict = dto_detector_result_request_instance.to_dict()
# create an instance of DtoDetectorResultRequest from a dict
dto_detector_result_request_from_dict = DtoDetectorResultRequest.from_dict(dto_detector_result_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


