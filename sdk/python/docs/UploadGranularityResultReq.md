# UploadGranularityResultReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**duration** | **float** | Execution duration in seconds | 
**results** | [**List[GranularityResultItem]**](GranularityResultItem.md) |  | 

## Example

```python
from openapi.models.upload_granularity_result_req import UploadGranularityResultReq

# TODO update the JSON string below
json = "{}"
# create an instance of UploadGranularityResultReq from a JSON string
upload_granularity_result_req_instance = UploadGranularityResultReq.from_json(json)
# print the JSON string representation of the object
print(UploadGranularityResultReq.to_json())

# convert the object into a dict
upload_granularity_result_req_dict = upload_granularity_result_req_instance.to_dict()
# create an instance of UploadGranularityResultReq from a dict
upload_granularity_result_req_from_dict = UploadGranularityResultReq.from_dict(upload_granularity_result_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


