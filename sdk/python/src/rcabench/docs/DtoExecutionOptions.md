# DtoExecutionOptions


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | [**DtoAlgorithmItem**](DtoAlgorithmItem.md) |  | [optional] 
**dataset** | **str** |  | [optional] 
**execution_id** | **int** |  | [optional] 
**timestamp** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_execution_options import DtoExecutionOptions

# TODO update the JSON string below
json = "{}"
# create an instance of DtoExecutionOptions from a JSON string
dto_execution_options_instance = DtoExecutionOptions.from_json(json)
# print the JSON string representation of the object
print(DtoExecutionOptions.to_json())

# convert the object into a dict
dto_execution_options_dict = dto_execution_options_instance.to_dict()
# create an instance of DtoExecutionOptions from a dict
dto_execution_options_from_dict = DtoExecutionOptions.from_dict(dto_execution_options_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


