# DtoExecutionPayload


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | [**DtoAlgorithmItem**](DtoAlgorithmItem.md) |  | 
**dataset** | **str** |  | 
**env_vars** | **object** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_execution_payload import DtoExecutionPayload

# TODO update the JSON string below
json = "{}"
# create an instance of DtoExecutionPayload from a JSON string
dto_execution_payload_instance = DtoExecutionPayload.from_json(json)
# print the JSON string representation of the object
print(DtoExecutionPayload.to_json())

# convert the object into a dict
dto_execution_payload_dict = dto_execution_payload_instance.to_dict()
# create an instance of DtoExecutionPayload from a dict
dto_execution_payload_from_dict = DtoExecutionPayload.from_dict(dto_execution_payload_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


