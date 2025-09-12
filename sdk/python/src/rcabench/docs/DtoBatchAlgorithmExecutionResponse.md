# DtoBatchAlgorithmExecutionResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**executions** | [**List[DtoAlgorithmExecutionResponse]**](DtoAlgorithmExecutionResponse.md) |  | [optional] 
**group_id** | **str** |  | [optional] 
**message** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_batch_algorithm_execution_response import DtoBatchAlgorithmExecutionResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoBatchAlgorithmExecutionResponse from a JSON string
dto_batch_algorithm_execution_response_instance = DtoBatchAlgorithmExecutionResponse.from_json(json)
# print the JSON string representation of the object
print(DtoBatchAlgorithmExecutionResponse.to_json())

# convert the object into a dict
dto_batch_algorithm_execution_response_dict = dto_batch_algorithm_execution_response_instance.to_dict()
# create an instance of DtoBatchAlgorithmExecutionResponse from a dict
dto_batch_algorithm_execution_response_from_dict = DtoBatchAlgorithmExecutionResponse.from_dict(dto_batch_algorithm_execution_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


