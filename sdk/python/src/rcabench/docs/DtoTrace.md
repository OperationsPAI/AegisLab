# DtoTrace


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**head_task_id** | **str** |  | [optional] 
**index** | **int** |  | [optional] 
**trace_id** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_trace import DtoTrace

# TODO update the JSON string below
json = "{}"
# create an instance of DtoTrace from a JSON string
dto_trace_instance = DtoTrace.from_json(json)
# print the JSON string representation of the object
print(DtoTrace.to_json())

# convert the object into a dict
dto_trace_dict = dto_trace_instance.to_dict()
# create an instance of DtoTrace from a dict
dto_trace_from_dict = DtoTrace.from_dict(dto_trace_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


