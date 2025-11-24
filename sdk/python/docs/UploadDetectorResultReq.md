# UploadDetectorResultReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**duration** | **float** | Execution duration in seconds | 
**results** | [**List[DetectorResultItem]**](DetectorResultItem.md) |  | 

## Example

```python
from rcabench.openapi.models.upload_detector_result_req import UploadDetectorResultReq

# TODO update the JSON string below
json = "{}"
# create an instance of UploadDetectorResultReq from a JSON string
upload_detector_result_req_instance = UploadDetectorResultReq.from_json(json)
# print the JSON string representation of the object
print(UploadDetectorResultReq.to_json())

# convert the object into a dict
upload_detector_result_req_dict = upload_detector_result_req_instance.to_dict()
# create an instance of UploadDetectorResultReq from a dict
upload_detector_result_req_from_dict = UploadDetectorResultReq.from_dict(upload_detector_result_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


