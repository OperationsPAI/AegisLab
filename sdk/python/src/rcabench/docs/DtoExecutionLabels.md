# DtoExecutionLabels


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**tag** | **str** | user-defined tag | [optional] 

## Example

```python
from rcabench.openapi.models.dto_execution_labels import DtoExecutionLabels

# TODO update the JSON string below
json = "{}"
# create an instance of DtoExecutionLabels from a JSON string
dto_execution_labels_instance = DtoExecutionLabels.from_json(json)
# print the JSON string representation of the object
print(DtoExecutionLabels.to_json())

# convert the object into a dict
dto_execution_labels_dict = dto_execution_labels_instance.to_dict()
# create an instance of DtoExecutionLabels from a dict
dto_execution_labels_from_dict = DtoExecutionLabels.from_dict(dto_execution_labels_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


