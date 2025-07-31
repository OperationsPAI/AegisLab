# DtoAlgorithmExecutionPayload


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | **str** |  | [optional] 
**dataset** | **str** |  | [optional] 
**env_vars** | **object** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_algorithm_execution_payload import DtoAlgorithmExecutionPayload

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAlgorithmExecutionPayload from a JSON string
dto_algorithm_execution_payload_instance = DtoAlgorithmExecutionPayload.from_json(json)
# print the JSON string representation of the object
print(DtoAlgorithmExecutionPayload.to_json())

# convert the object into a dict
dto_algorithm_execution_payload_dict = dto_algorithm_execution_payload_instance.to_dict()
# create an instance of DtoAlgorithmExecutionPayload from a dict
dto_algorithm_execution_payload_from_dict = DtoAlgorithmExecutionPayload.from_dict(dto_algorithm_execution_payload_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


