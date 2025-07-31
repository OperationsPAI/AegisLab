# DtoAlgorithmExecutionRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | [**DtoAlgorithmItem**](DtoAlgorithmItem.md) |  | 
**datapack** | **str** |  | [optional] 
**dataset** | **str** |  | [optional] 
**env_vars** | **object** |  | [optional] 
**project_name** | **str** |  | 

## Example

```python
from rcabench.openapi.models.dto_algorithm_execution_request import DtoAlgorithmExecutionRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAlgorithmExecutionRequest from a JSON string
dto_algorithm_execution_request_instance = DtoAlgorithmExecutionRequest.from_json(json)
# print the JSON string representation of the object
print(DtoAlgorithmExecutionRequest.to_json())

# convert the object into a dict
dto_algorithm_execution_request_dict = dto_algorithm_execution_request_instance.to_dict()
# create an instance of DtoAlgorithmExecutionRequest from a dict
dto_algorithm_execution_request_from_dict = DtoAlgorithmExecutionRequest.from_dict(dto_algorithm_execution_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


