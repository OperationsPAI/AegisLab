# DtoBatchAlgorithmExecutionRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**executions** | [**List[DtoAlgorithmExecutionRequest]**](DtoAlgorithmExecutionRequest.md) |  | 
**labels** | [**DtoExecutionLabels**](DtoExecutionLabels.md) | 预置的执行标签 | [optional] 
**project_name** | **str** |  | 

## Example

```python
from rcabench.openapi.models.dto_batch_algorithm_execution_request import DtoBatchAlgorithmExecutionRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DtoBatchAlgorithmExecutionRequest from a JSON string
dto_batch_algorithm_execution_request_instance = DtoBatchAlgorithmExecutionRequest.from_json(json)
# print the JSON string representation of the object
print(DtoBatchAlgorithmExecutionRequest.to_json())

# convert the object into a dict
dto_batch_algorithm_execution_request_dict = dto_batch_algorithm_execution_request_instance.to_dict()
# create an instance of DtoBatchAlgorithmExecutionRequest from a dict
dto_batch_algorithm_execution_request_from_dict = DtoBatchAlgorithmExecutionRequest.from_dict(dto_batch_algorithm_execution_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


