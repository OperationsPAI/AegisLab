# DtoAlgorithmExecutionResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm_id** | **int** |  | [optional] 
**datapack_id** | **int** |  | [optional] 
**dataset_id** | **int** |  | [optional] 
**status** | **str** |  | [optional] 
**task_id** | **str** |  | [optional] 
**trace_id** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_algorithm_execution_response import DtoAlgorithmExecutionResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAlgorithmExecutionResponse from a JSON string
dto_algorithm_execution_response_instance = DtoAlgorithmExecutionResponse.from_json(json)
# print the JSON string representation of the object
print(DtoAlgorithmExecutionResponse.to_json())

# convert the object into a dict
dto_algorithm_execution_response_dict = dto_algorithm_execution_response_instance.to_dict()
# create an instance of DtoAlgorithmExecutionResponse from a dict
dto_algorithm_execution_response_from_dict = DtoAlgorithmExecutionResponse.from_dict(dto_algorithm_execution_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


