# UploadExecutionResultResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**has_anomalies** | **bool** | Only included for detector results | [optional] 
**result_count** | **int** |  | [optional] 
**uploaded_at** | **str** |  | [optional] 

## Example

```python
from openapi.models.upload_execution_result_resp import UploadExecutionResultResp

# TODO update the JSON string below
json = "{}"
# create an instance of UploadExecutionResultResp from a JSON string
upload_execution_result_resp_instance = UploadExecutionResultResp.from_json(json)
# print the JSON string representation of the object
print(UploadExecutionResultResp.to_json())

# convert the object into a dict
upload_execution_result_resp_dict = upload_execution_result_resp_instance.to_dict()
# create an instance of UploadExecutionResultResp from a dict
upload_execution_result_resp_from_dict = UploadExecutionResultResp.from_dict(upload_execution_result_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


