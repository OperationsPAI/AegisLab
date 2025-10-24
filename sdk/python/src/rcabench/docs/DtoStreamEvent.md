# DtoStreamEvent


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**event_name** | [**ConstsEventType**](ConstsEventType.md) |  | [optional] 
**payload** | **object** |  | [optional] 
**task_id** | **str** |  | [optional] 
**task_type** | [**ConstsTaskType**](ConstsTaskType.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_stream_event import DtoStreamEvent

# TODO update the JSON string below
json = "{}"
# create an instance of DtoStreamEvent from a JSON string
dto_stream_event_instance = DtoStreamEvent.from_json(json)
# print the JSON string representation of the object
print(DtoStreamEvent.to_json())

# convert the object into a dict
dto_stream_event_dict = dto_stream_event_instance.to_dict()
# create an instance of DtoStreamEvent from a dict
dto_stream_event_from_dict = DtoStreamEvent.from_dict(dto_stream_event_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


