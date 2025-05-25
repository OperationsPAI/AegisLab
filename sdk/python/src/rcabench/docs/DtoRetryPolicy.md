# DtoRetryPolicy


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**backoff_sec** | **int** | Seconds to wait between retries | [optional] 
**max_attempts** | **int** | Maximum number of retry attempts | [optional] 

## Example

```python
from rcabench.openapi.models.dto_retry_policy import DtoRetryPolicy

# TODO update the JSON string below
json = "{}"
# create an instance of DtoRetryPolicy from a JSON string
dto_retry_policy_instance = DtoRetryPolicy.from_json(json)
# print the JSON string representation of the object
print(DtoRetryPolicy.to_json())

# convert the object into a dict
dto_retry_policy_dict = dto_retry_policy_instance.to_dict()
# create an instance of DtoRetryPolicy from a dict
dto_retry_policy_from_dict = DtoRetryPolicy.from_dict(dto_retry_policy_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


